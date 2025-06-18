package loadtest

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
	"github.com/orijtech/cosmosloadtester/pkg/httprpc"
)

// SimpleHybridTransactor is a simple wrapper that delegates to WebSocket or HTTP
type SimpleHybridTransactor struct {
	remoteAddr        string
	protocol          string
	config            *loadtest.Config
	wsTransactor      *loadtest.Transactor
	httpClient        *httprpc.HTTPRPCClient
	logger            *logrus.Logger
	broadcastTxMethod string
	
	// Stats tracking
	statsMtx  sync.RWMutex
	startTime time.Time
	txCount   int
	txBytes   int64
	txRate    float64
	
	// Progress callback
	progressCallbackMtx      sync.RWMutex
	progressCallbackID       int
	progressCallbackInterval time.Duration
	progressCallback         func(id int, txCount int, txBytes int64)
	
	// Control
	stopMtx sync.RWMutex
	stop    bool
	stopErr error
}

// NewHybridTransactor creates a new hybrid transactor
func NewHybridTransactor(remoteAddr string, config *loadtest.Config) (*SimpleHybridTransactor, error) {
	u, err := url.Parse(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	protocol := u.Scheme
	if protocol != "ws" && protocol != "wss" && protocol != "http" && protocol != "https" {
		return nil, fmt.Errorf("unsupported protocol: %s (supported: ws://, wss://, http://, https://)", protocol)
	}

	logger := logrus.WithField("component", fmt.Sprintf("simple-hybrid-transactor[%s]", u.String())).Logger
	
	transactor := &SimpleHybridTransactor{
		remoteAddr:               remoteAddr,
		protocol:                 protocol,
		config:                   config,
		logger:                   logger,
		broadcastTxMethod:        "broadcast_tx_" + config.BroadcastTxMethod,
		progressCallbackInterval: 5 * time.Second,
	}

	// Initialize based on protocol
	switch protocol {
	case "ws", "wss":
		wsTransactor, err := loadtest.NewTransactor(remoteAddr, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create WebSocket transactor: %w", err)
		}
		transactor.wsTransactor = wsTransactor
	case "http", "https":
		httpClient, err := httprpc.NewHTTPRPCClient(remoteAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP RPC client: %w", err)
		}
		transactor.httpClient = httpClient
	}

	logger.Infof("Created hybrid transactor for %s protocol", protocol)
	return transactor, nil
}

// SetProgressCallback sets the progress callback
func (t *SimpleHybridTransactor) SetProgressCallback(id int, interval time.Duration, callback func(int, int, int64)) {
	t.progressCallbackMtx.Lock()
	defer t.progressCallbackMtx.Unlock()
	t.progressCallbackID = id
	t.progressCallbackInterval = interval
	t.progressCallback = callback
	
	// Also set on WebSocket transactor if it exists
	if t.wsTransactor != nil {
		t.wsTransactor.SetProgressCallback(id, interval, callback)
	}
}

// Start starts the transactor
func (t *SimpleHybridTransactor) Start() {
	t.logger.Info("Starting hybrid transactor")
	
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		t.wsTransactor.Start()
		return
	}
	
	// For HTTP, we need to handle this ourselves
	if t.httpClient != nil {
		t.statsMtx.Lock()
		t.startTime = time.Now()
		t.statsMtx.Unlock()
		
		// TODO: Implement HTTP transaction sending
		// For now, just log that this is not fully implemented
		t.logger.Warn("HTTP transaction sending not yet implemented - requires client factory access")
		return
	}
	
	t.logger.Error("No transactor or client available")
}

// Cancel cancels the transactor
func (t *SimpleHybridTransactor) Cancel() {
	t.logger.Info("Cancelling hybrid transactor")
	
	t.stopMtx.Lock()
	t.stop = true
	t.stopMtx.Unlock()
	
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		t.wsTransactor.Cancel()
		return
	}
	
	// For HTTP, we handle the cancellation ourselves
	t.logger.Info("HTTP transactor cancelled")
}

// Wait waits for the transactor to finish
func (t *SimpleHybridTransactor) Wait() error {
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		return t.wsTransactor.Wait()
	}
	
	// For HTTP, close the client
	if t.httpClient != nil {
		return t.httpClient.Close()
	}
	
	return nil
}

// GetTxCount returns the transaction count
func (t *SimpleHybridTransactor) GetTxCount() int {
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		return t.wsTransactor.GetTxCount()
	}
	
	// For HTTP, return our own stats
	t.statsMtx.RLock()
	defer t.statsMtx.RUnlock()
	return t.txCount
}

// GetTxBytes returns the transaction bytes
func (t *SimpleHybridTransactor) GetTxBytes() int64 {
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		return t.wsTransactor.GetTxBytes()
	}
	
	// For HTTP, return our own stats
	t.statsMtx.RLock()
	defer t.statsMtx.RUnlock()
	return t.txBytes
}

// GetTxRate returns the transaction rate
func (t *SimpleHybridTransactor) GetTxRate() float64 {
	// For WebSocket, delegate to the original transactor
	if t.wsTransactor != nil {
		return t.wsTransactor.GetTxRate()
	}
	
	// For HTTP, return our own stats
	t.statsMtx.RLock()
	defer t.statsMtx.RUnlock()
	return t.txRate
}