package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	tmloadtest "github.com/informalsystems/tm-load-test/pkg/loadtest"
	"github.com/orijtech/cosmosloadtester/pkg/loadtest"
	loadtestpb "github.com/orijtech/cosmosloadtester/proto/orijtech/cosmosloadtester/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HybridServer extends the original server with HTTPS protocol support
type HybridServer struct {
	*Server
	transactorFactory *loadtest.TransactorFactory
}

// NewHybridServer creates a new server that supports both WebSocket and HTTP(S) protocols
func NewHybridServer() *HybridServer {
	return &HybridServer{
		Server:            NewServer(),
		transactorFactory: loadtest.NewTransactorFactory(),
	}
}

// RunLoadtest runs a load test with hybrid protocol support
func (s *HybridServer) RunLoadtest(ctx context.Context, req *loadtestpb.RunLoadtestRequest) (*loadtestpb.RunLoadtestResponse, error) {
	logrus.Info("Starting hybrid load test with protocol auto-detection")

	// Validate and convert endpoints
	if len(req.Endpoints) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one endpoint must be specified")
	}

	// Log detected protocols
	for _, endpoint := range req.Endpoints {
		protocol := detectProtocol(endpoint)
		logrus.Infof("Detected protocol for %s: %s", endpoint, protocol)
	}

	// Use enhanced configuration validation
	config, err := s.buildHybridConfig(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid configuration: %v", err)
	}

	// Validate that all client factories support the requested configuration
	for _, clientFactoryName := range []string{config.ClientFactory} {
		if err := s.validateClientFactory(clientFactoryName, *config); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "client factory validation failed: %v", err)
		}
	}

	// Create and run hybrid load test
	return s.runHybridLoadTest(ctx, config)
}

func (s *HybridServer) buildHybridConfig(req *loadtestpb.RunLoadtestRequest) (*tmloadtest.Config, error) {
	broadcastTxMethod, err := mapBroadcastTxMethod(req.BroadcastTxMethod)
	if err != nil {
		return nil, err
	}

	endpointSelectMethod, err := mapEndpointSelectMethod(req.EndpointSelectMethod)
	if err != nil {
		return nil, err
	}

	config := &tmloadtest.Config{
		ClientFactory:        req.ClientFactory,
		Connections:          int(req.ConnectionCount),
		Time:                 int(req.Duration.GetSeconds()),
		SendPeriod:           int(req.SendPeriod.GetSeconds()),
		Rate:                 int(req.TransactionsPerSecond),
		Size:                 int(req.TransactionSizeBytes),
		Count:                int(req.TransactionCount),
		BroadcastTxMethod:    broadcastTxMethod,
		Endpoints:            req.Endpoints,
		EndpointSelectMethod: endpointSelectMethod,
		ExpectPeers:          int(req.ExpectPeersCount),
		MaxEndpoints:         int(req.MaxEndpointCount),
		MinConnectivity:      int(req.MinPeerConnectivityCount),
		PeerConnectTimeout:   int(req.PeerConnectTimeout.GetSeconds()),
		StatsOutputFile:      req.StatsOutputFilePath,
		NoTrapInterrupts:     true,
	}

	return config, nil
}

func (s *HybridServer) validateClientFactory(factoryName string, config tmloadtest.Config) error {
	// This would validate that the client factory exists and supports the configuration
	// For now, we'll do basic validation
	if strings.TrimSpace(factoryName) == "" {
		return fmt.Errorf("client factory name cannot be empty")
	}
	return nil
}

func (s *HybridServer) runHybridLoadTest(ctx context.Context, config *tmloadtest.Config) (*loadtestpb.RunLoadtestResponse, error) {
	logrus.Infof("Running hybrid load test with %d endpoints", len(config.Endpoints))

	// Create transactors for each endpoint using the factory
	var transactors []loadtest.TransactorInterface
	for i, endpoint := range config.Endpoints {
		transactor, err := s.transactorFactory.CreateTransactor(endpoint, config)
		if err != nil {
			logrus.Errorf("Failed to create transactor for endpoint %s: %v", endpoint, err)
			// Clean up previously created transactors
			for _, t := range transactors {
				t.Cancel()
				t.Wait()
			}
			return nil, status.Errorf(codes.Internal, "failed to create transactor for endpoint %s: %v", endpoint, err)
		}

		// Set progress callback
		transactor.SetProgressCallback(i, 5*time.Second, func(id int, txCount int, txBytes int64) {
			logrus.Infof("Transactor %d progress: %d transactions, %d bytes", id, txCount, txBytes)
		})

		transactors = append(transactors, transactor)
	}

	// Start all transactors
	for i, transactor := range transactors {
		transactor.Start()
		logrus.Infof("Started transactor %d", i)
	}

	// Wait for the test duration
	testDuration := time.Duration(config.Time) * time.Second
	logrus.Infof("Running load test for %v", testDuration)

	select {
	case <-time.After(testDuration):
		logrus.Info("Load test duration completed")
	case <-ctx.Done():
		logrus.Info("Load test cancelled by context")
	}

	// Cancel all transactors and collect stats
	var totalTxCount int
	var totalTxBytes int64
	var avgTxRate float64

	for i, transactor := range transactors {
		transactor.Cancel()
		if err := transactor.Wait(); err != nil {
			logrus.Errorf("Error waiting for transactor %d: %v", i, err)
		}

		txCount := transactor.GetTxCount()
		txBytes := transactor.GetTxBytes()
		txRate := transactor.GetTxRate()
		totalTxCount += txCount
		totalTxBytes += txBytes
		avgTxRate += txRate

		logrus.Infof("Transactor %d final stats: %d transactions, %d bytes, %.2f tx/s", 
			i, txCount, txBytes, txRate)
	}

	if len(transactors) > 0 {
		avgTxRate = avgTxRate / float64(len(transactors))
	}

	logrus.Infof("Hybrid load test completed: %d total transactions, %d total bytes, %.2f avg tx/s", 
		totalTxCount, totalTxBytes, avgTxRate)

	// Build response
	response := &loadtestpb.RunLoadtestResponse{
		TotalTxs:           int64(totalTxCount),
		TotalBytes:         totalTxBytes,
		AvgTxsPerSecond:    avgTxRate,
		AvgBytesPerSecond:  float64(totalTxBytes) / float64(config.Time),
	}

	return response, nil
}

func detectProtocol(endpoint string) string {
	if strings.HasPrefix(endpoint, "ws://") {
		return "WebSocket"
	} else if strings.HasPrefix(endpoint, "wss://") {
		return "WebSocket Secure"
	} else if strings.HasPrefix(endpoint, "http://") {
		return "HTTP"
	} else if strings.HasPrefix(endpoint, "https://") {
		return "HTTPS"
	}
	return "Unknown"
}