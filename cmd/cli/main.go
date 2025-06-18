package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/fatih/color"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"

	"github.com/orijtech/cosmosloadtester/clients/aiw3defi"
	"github.com/orijtech/cosmosloadtester/clients/myabciapp"
)

// CLI flags
var (
	clientFactory        = flag.String("client-factory", "test-cosmos-client-factory", "Client factory to use for generating transactions")
	connections          = flag.Int("connections", 1, "Number of connections to open to each endpoint")
	duration             = flag.Duration("duration", 60*time.Second, "Duration for the load test")
	sendPeriod           = flag.Duration("send-period", 1*time.Second, "Period at which to send batches of transactions")
	transactionsPerSecond = flag.Int("rate", 1000, "Number of transactions to generate per second per connection")
	transactionSize      = flag.Int("size", 250, "Size of each transaction in bytes (min 40)")
	transactionCount     = flag.Int("count", -1, "Maximum number of transactions (-1 for unlimited)")
	broadcastMethod      = flag.String("broadcast-method", "sync", "Broadcast method: sync, async, or commit")
	endpoints            = flag.String("endpoints", "", "Comma-separated list of RPC endpoints (ws:// or http://)")
	endpointSelectMethod = flag.String("endpoint-select-method", "supplied", "Endpoint selection method: supplied, discovered, or any")
	expectPeers          = flag.Int("expect-peers", 0, "Expected number of peers for P2P crawling")
	maxEndpoints         = flag.Int("max-endpoints", 0, "Maximum number of endpoints (0 for unlimited)")
	minConnectivity      = flag.Int("min-connectivity", 0, "Minimum peer connectivity")
	peerConnectTimeout   = flag.Duration("peer-connect-timeout", 5*time.Second, "Timeout for peer connections")
	statsOutputFile      = flag.String("stats-output", "", "File to store statistics (CSV format)")
	outputFormat         = flag.String("output-format", "live", "Output format: live, json, csv, or summary")
	quiet                = flag.Bool("quiet", false, "Suppress progress output")
	logLevel             = flag.String("log-level", "info", "Log level: debug, info, warn, error")
	listFactories        = flag.Bool("list-factories", false, "List available client factories")
	showVersion          = flag.Bool("version", false, "Show version information")
	listProfiles         = flag.Bool("list-profiles", false, "List available profiles")
	showProfile          = flag.String("show-profile", "", "Show details for a specific profile")
	deleteProfile        = flag.String("delete-profile", "", "Delete a specific profile")
	generateTemplate     = flag.String("generate-template", "", "Generate a new template profile")
	exportProfiles       = flag.String("export-profiles", "", "Export profiles to a file")
	importProfiles       = flag.String("import-profiles", "", "Import profiles from a file")
	interactive          = flag.Bool("interactive", false, "Run in interactive mode")
	validateConfig       = flag.Bool("validate-config", false, "Validate configuration")
	dryRun               = flag.Bool("dry-run", false, "Run without actually executing transactions")
	checkEndpoints       = flag.Bool("check-endpoints", false, "Check endpoint connectivity")
	benchmark            = flag.String("benchmark", "", "Run a specific benchmark")
	profile              = flag.String("profile", "", "Use a specific profile for the load test")
)

const (
	version = "1.0.0"
	banner  = `
   ____                              _                    _ _____         _            
  / ___|___  ___ _ __ ___   ___  ___| |    ___   __ _  __| |_   _|__  ___| |_ ___ _ __ 
 | |   / _ \/ __| '_ ` + "`" + ` _ \ / _ \/ __| |   / _ \ / _` + "`" + ` |/ _` + "`" + ` | | |/ _ \/ __| __/ _ \ '__|
 | |__| (_) \__ \ | | | | | (_) \__ \ |__| (_) | (_| | (_| | | |  __/\__ \ ||  __/ |   
  \____\___/|___/_| |_| |_|\___/|___/____|\___/ \__,_|\__,_| |_|\___||___/\__\___|_|   
                                                                                       
                       Terminal-based Cosmos Load Testing Tool
`
)

// Stats represents load test statistics
type Stats struct {
	TotalTxs            int64                    `json:"total_txs"`
	TotalTime           time.Duration            `json:"total_time"`
	TotalBytes          int64                    `json:"total_bytes"`
	AvgTxsPerSecond     float64                  `json:"avg_txs_per_second"`
	AvgBytesPerSecond   float64                  `json:"avg_bytes_per_second"`
	PerSecondStats      []PerSecondStats         `json:"per_second_stats"`
	EndpointStats       map[string]EndpointStats `json:"endpoint_stats"`
	ClientFactoryUsed   string                   `json:"client_factory_used"`
	ConfigurationUsed   loadtest.Config          `json:"configuration_used"`
}

// PerSecondStats represents per-second statistics
type PerSecondStats struct {
	Second          int64              `json:"second"`
	TxsPerSecond    float64            `json:"txs_per_second"`
	BytesPerSecond  float64            `json:"bytes_per_second"`
	LatencyP50      time.Duration      `json:"latency_p50"`
	LatencyP75      time.Duration      `json:"latency_p75"`
	LatencyP90      time.Duration      `json:"latency_p90"`
	LatencyP95      time.Duration      `json:"latency_p95"`
	LatencyP99      time.Duration      `json:"latency_p99"`
	SuccessRate     float64            `json:"success_rate"`
	ErrorCount      int64              `json:"error_count"`
}

// EndpointStats represents statistics for a specific endpoint
type EndpointStats struct {
	Endpoint        string        `json:"endpoint"`
	Protocol        string        `json:"protocol"`
	TotalTxs        int64         `json:"total_txs"`
	TotalBytes      int64         `json:"total_bytes"`
	AvgLatency      time.Duration `json:"avg_latency"`
	ErrorCount      int64         `json:"error_count"`
	ConnectionCount int           `json:"connection_count"`
}

// ProgressReporter handles real-time progress reporting
type ProgressReporter struct {
	startTime    time.Time
	progressBar  *progressbar.ProgressBar
	stats        *Stats
	mu           sync.RWMutex
	quiet        bool
	outputFormat string
}

func main() {
	flag.Parse()

	// Setup logging
	setupLogging()

	// Show version
	if *showVersion {
		fmt.Printf("cosmosloadtester-cli version %s\n", version)
		return
	}

	// Show banner
	if !*quiet && *outputFormat == "live" {
		color.Cyan(banner)
	}

	// Register client factories
	if err := registerClientFactories(); err != nil {
		log.Fatalf("Failed to register client factories: %v", err)
	}

	// List factories if requested
	if *listFactories {
		listAvailableFactories()
		return
	}

	// Initialize enhanced CLI
	cli, err := NewCLI()
	if err != nil {
		log.Fatalf("Failed to initialize CLI: %v", err)
	}

	// Process CLI commands first
	if err := cli.Run(); err != nil {
		log.Fatalf("CLI command failed: %v", err)
	}

	// If no CLI commands were processed, run standard load test
	if !shouldRunStandardLoadTest() {
		return
	}

	// Validate configuration
	config, err := buildConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Run load test
	if err := runLoadTest(config); err != nil {
		log.Fatalf("Load test failed: %v", err)
	}
}

// shouldRunStandardLoadTest determines if we should run the standard load test
// based on which flags were provided
func shouldRunStandardLoadTest() bool {
	// Don't run standard load test if any of these management commands were used
	if *listProfiles || *showProfile != "" || *deleteProfile != "" || 
	   *generateTemplate != "" || *exportProfiles != "" || *importProfiles != "" ||
	   *interactive || *validateConfig || *dryRun || *checkEndpoints || *benchmark != "" {
		return false
	}

	// Run standard load test if profile is specified or basic parameters are provided
	return *profile != "" || *endpoints != ""
}

func setupLogging() {
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})
}

func registerClientFactories() error {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	// Register the default test client factory
	cosmosClientFactory := myabciapp.NewCosmosClientFactory(txConfig)
	if err := loadtest.RegisterClientFactory("test-cosmos-client-factory", cosmosClientFactory); err != nil {
		return fmt.Errorf("failed to register client factory %s: %w", "test-cosmos-client-factory", err)
	}

	// Register the AIW3 DeFi client factory
	aiw3defiClientFactory := aiw3defi.NewAIW3DefiClientFactory(txConfig)
	if err := loadtest.RegisterClientFactory("aiw3defi-bank-send", aiw3defiClientFactory); err != nil {
		return fmt.Errorf("failed to register client factory %s: %w", "aiw3defi-bank-send", err)
	}

	return nil
}

func listAvailableFactories() {
	// Since there's no public API to get registered factories, 
	// we'll list the ones we know are registered
	factories := []string{"test-cosmos-client-factory", "aiw3defi-bank-send"}
	color.Green("Available Client Factories:")
	for _, factory := range factories {
		color.White("  • %s", factory)
	}
	
	if len(factories) == 0 {
		color.Yellow("No client factories registered")
	}
}

func buildConfig() (loadtest.Config, error) {
	var config loadtest.Config

	// Validate endpoints
	if *endpoints == "" {
		return config, fmt.Errorf("endpoints are required")
	}
	endpointList := strings.Split(*endpoints, ",")
	for i, endpoint := range endpointList {
		endpointList[i] = strings.TrimSpace(endpoint)
	}

	// Validate broadcast method
	validBroadcastMethods := map[string]bool{
		"sync":   true,
		"async":  true,
		"commit": true,
	}
	if !validBroadcastMethods[*broadcastMethod] {
		return config, fmt.Errorf("invalid broadcast method: %s (valid: sync, async, commit)", *broadcastMethod)
	}

	// Validate endpoint select method
	validEndpointSelectMethods := map[string]bool{
		"supplied":   true,
		"discovered": true,
		"any":        true,
	}
	if !validEndpointSelectMethods[*endpointSelectMethod] {
		return config, fmt.Errorf("invalid endpoint select method: %s (valid: supplied, discovered, any)", *endpointSelectMethod)
	}

	// Create temporary stats file if not provided
	statsFile := *statsOutputFile
	if statsFile == "" {
		tmpFile, err := os.CreateTemp("", "cosmosloadtester-*.csv")
		if err != nil {
			return config, fmt.Errorf("failed to create temporary stats file: %w", err)
		}
		statsFile = tmpFile.Name()
		tmpFile.Close()
	}

	config = loadtest.Config{
		ClientFactory:        *clientFactory,
		Connections:          *connections,
		Time:                 int(duration.Seconds()),
		SendPeriod:           int(sendPeriod.Seconds()),
		Rate:                 *transactionsPerSecond,
		Size:                 *transactionSize,
		Count:                *transactionCount,
		BroadcastTxMethod:    *broadcastMethod,
		Endpoints:            endpointList,
		EndpointSelectMethod: *endpointSelectMethod,
		ExpectPeers:          *expectPeers,
		MaxEndpoints:         *maxEndpoints,
		MinConnectivity:      *minConnectivity,
		PeerConnectTimeout:   int(peerConnectTimeout.Seconds()),
		StatsOutputFile:      statsFile,
		NoTrapInterrupts:     false,
	}

	return config, config.Validate()
}

func runLoadTest(config loadtest.Config) error {
	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup progress reporter
	reporter := &ProgressReporter{
		startTime:    time.Now(),
		quiet:        *quiet,
		outputFormat: *outputFormat,
		stats: &Stats{
			EndpointStats:     make(map[string]EndpointStats),
			ClientFactoryUsed: config.ClientFactory,
			ConfigurationUsed: config,
		},
	}

	// Show configuration
	if !*quiet {
		displayConfiguration(config)
	}

	// Setup progress bar for live output
	if *outputFormat == "live" && !*quiet {
		reporter.progressBar = progressbar.NewOptions(int(config.Time),
			progressbar.OptionSetDescription("Load Testing Progress"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerHead:    "█",
				SaucerPadding: "░",
				BarStart:      "│",
				BarEnd:        "│",
			}),
		)
	}

	// Start load test
	go func() {
		if err := executeLoadTest(ctx, config, reporter); err != nil {
			logrus.Errorf("Load test execution failed: %v", err)
			cancel()
		}
	}()

	// Wait for completion or interruption
	select {
	case <-ctx.Done():
		// Test completed
	case <-sigChan:
		color.Yellow("\nReceived interrupt signal, stopping load test...")
		cancel()
		time.Sleep(2 * time.Second) // Give time for cleanup
	}

	// Display final results
	return displayResults(reporter.stats)
}

func displayConfiguration(config loadtest.Config) {
	color.Green("\n=== Load Test Configuration ===")
	color.White("Client Factory: %s", config.ClientFactory)
	color.White("Connections: %d per endpoint", config.Connections)
	color.White("Duration: %s", time.Duration(config.Time)*time.Second)
	color.White("Send Period: %s", time.Duration(config.SendPeriod)*time.Second)
	color.White("Rate: %d transactions/second per connection", config.Rate)
	color.White("Transaction Size: %d bytes", config.Size)
	if config.Count == -1 {
		color.White("Transaction Count: unlimited")
	} else {
		color.White("Transaction Count: %d", config.Count)
	}
	color.White("Broadcast Method: %s", config.BroadcastTxMethod)
	color.White("Endpoints:")
	for _, endpoint := range config.Endpoints {
		protocol := "http"
		if strings.HasPrefix(endpoint, "ws") {
			protocol = "websocket"
		}
		color.White("  • %s (%s)", endpoint, protocol)
	}
	color.White("Endpoint Selection: %s", config.EndpointSelectMethod)
	color.Green("================================\n")
}

func executeLoadTest(ctx context.Context, config loadtest.Config, reporter *ProgressReporter) error {
	// Start periodic reporting
	if *outputFormat == "live" && !*quiet {
		go reporter.startPeriodicReporting()
	}

	// Execute the load test using the existing tm-load-test framework
	psL, err := loadtest.ExecuteStandaloneWithStats(config)
	if err != nil {
		return fmt.Errorf("load test execution failed: %w", err)
	}

	// Process results
	reporter.mu.Lock()
	defer reporter.mu.Unlock()

	for _, ps := range psL {
		reporter.stats.TotalTxs += int64(ps.TotalTxs)
		reporter.stats.TotalBytes += int64(ps.TotalBytes)
		reporter.stats.TotalTime = ps.TotalTime
		reporter.stats.AvgTxsPerSecond = ps.AvgTxPerSecond
		reporter.stats.AvgBytesPerSecond = ps.AvgBytesPerSecond

		// Process per-second stats
		for _, perSec := range ps.PerSecond {
			stats := PerSecondStats{
				Second:         int64(perSec.Sec),
				TxsPerSecond:   float64(perSec.QPS),
				BytesPerSecond: float64(perSec.Bytes),
			}

			// Extract latency percentiles if available
			if perSec.LatencyRankings != nil {
				if perSec.LatencyRankings.P50thLatency != nil {
					stats.LatencyP50 = perSec.LatencyRankings.P50thLatency.Latency
				}
				if perSec.LatencyRankings.P75thLatency != nil {
					stats.LatencyP75 = perSec.LatencyRankings.P75thLatency.Latency
				}
				if perSec.LatencyRankings.P90thLatency != nil {
					stats.LatencyP90 = perSec.LatencyRankings.P90thLatency.Latency
				}
				if perSec.LatencyRankings.P95thLatency != nil {
					stats.LatencyP95 = perSec.LatencyRankings.P95thLatency.Latency
				}
				if perSec.LatencyRankings.P99thLatency != nil {
					stats.LatencyP99 = perSec.LatencyRankings.P99thLatency.Latency
				}
			}

			reporter.stats.PerSecondStats = append(reporter.stats.PerSecondStats, stats)
		}
	}

	return nil
}

func (r *ProgressReporter) startPeriodicReporting() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.updateProgress()
		}
	}
}

func (r *ProgressReporter) updateProgress() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.progressBar != nil {
		elapsed := time.Since(r.startTime)
		r.progressBar.Set(int(elapsed.Seconds()))
	}
}

func displayResults(stats *Stats) error {
	switch *outputFormat {
	case "json":
		return displayJSONResults(stats)
	case "csv":
		return displayCSVResults(stats)
	case "summary":
		return displaySummaryResults(stats)
	default: // "live"
		return displayLiveResults(stats)
	}
}

func displayLiveResults(stats *Stats) error {
	color.Green("\n=== Load Test Results ===")
	color.White("Total Transactions: %s", formatNumber(stats.TotalTxs))
	color.White("Total Time: %s", stats.TotalTime.Round(time.Millisecond))
	color.White("Total Bytes: %s", formatBytes(stats.TotalBytes))
	color.White("Average TPS: %.2f", stats.AvgTxsPerSecond)
	color.White("Average Throughput: %s/sec", formatBytes(int64(stats.AvgBytesPerSecond)))

	if len(stats.PerSecondStats) > 0 {
		color.Green("\n=== Latency Percentiles (Last Second) ===")
		lastSec := stats.PerSecondStats[len(stats.PerSecondStats)-1]
		color.White("P50 (Median): %s", lastSec.LatencyP50.Round(time.Microsecond))
		color.White("P75: %s", lastSec.LatencyP75.Round(time.Microsecond))
		color.White("P90: %s", lastSec.LatencyP90.Round(time.Microsecond))
		color.White("P95: %s", lastSec.LatencyP95.Round(time.Microsecond))
		color.White("P99: %s", lastSec.LatencyP99.Round(time.Microsecond))
	}

	color.Green("\n=== Endpoint Statistics ===")
	for endpoint, endpointStats := range stats.EndpointStats {
		color.White("Endpoint: %s (%s)", endpoint, endpointStats.Protocol)
		color.White("  Transactions: %s", formatNumber(endpointStats.TotalTxs))
		color.White("  Bytes: %s", formatBytes(endpointStats.TotalBytes))
		color.White("  Avg Latency: %s", endpointStats.AvgLatency.Round(time.Microsecond))
		color.White("  Connections: %d", endpointStats.ConnectionCount)
		if endpointStats.ErrorCount > 0 {
			color.Red("  Errors: %d", endpointStats.ErrorCount)
		}
	}

	color.Green("\n=== Configuration Used ===")
	color.White("Client Factory: %s", stats.ClientFactoryUsed)
	color.White("Connections: %d per endpoint", stats.ConfigurationUsed.Connections)
	color.White("Broadcast Method: %s", stats.ConfigurationUsed.BroadcastTxMethod)

	return nil
}

func displayJSONResults(stats *Stats) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(stats)
}

func displayCSVResults(stats *Stats) error {
	// Display summary in CSV format
	fmt.Println("metric,value")
	fmt.Printf("total_txs,%d\n", stats.TotalTxs)
	fmt.Printf("total_time_seconds,%.3f\n", stats.TotalTime.Seconds())
	fmt.Printf("total_bytes,%d\n", stats.TotalBytes)
	fmt.Printf("avg_txs_per_second,%.2f\n", stats.AvgTxsPerSecond)
	fmt.Printf("avg_bytes_per_second,%.2f\n", stats.AvgBytesPerSecond)
	fmt.Printf("client_factory,%s\n", stats.ClientFactoryUsed)

	// Per-second statistics
	fmt.Println("\nsecond,txs_per_second,bytes_per_second,latency_p50_us,latency_p75_us,latency_p90_us,latency_p95_us,latency_p99_us")
	for _, ps := range stats.PerSecondStats {
		fmt.Printf("%d,%.2f,%.2f,%d,%d,%d,%d,%d\n",
			ps.Second,
			ps.TxsPerSecond,
			ps.BytesPerSecond,
			ps.LatencyP50.Nanoseconds()/1000,
			ps.LatencyP75.Nanoseconds()/1000,
			ps.LatencyP90.Nanoseconds()/1000,
			ps.LatencyP95.Nanoseconds()/1000,
			ps.LatencyP99.Nanoseconds()/1000,
		)
	}

	return nil
}

func displaySummaryResults(stats *Stats) error {
	fmt.Printf("TOTAL_TXS=%d\n", stats.TotalTxs)
	fmt.Printf("TOTAL_TIME=%.3f\n", stats.TotalTime.Seconds())
	fmt.Printf("TOTAL_BYTES=%d\n", stats.TotalBytes)
	fmt.Printf("AVG_TPS=%.2f\n", stats.AvgTxsPerSecond)
	fmt.Printf("AVG_THROUGHPUT=%.2f\n", stats.AvgBytesPerSecond)
	fmt.Printf("CLIENT_FACTORY=%s\n", stats.ClientFactoryUsed)

	if len(stats.PerSecondStats) > 0 {
		lastSec := stats.PerSecondStats[len(stats.PerSecondStats)-1]
		fmt.Printf("LATENCY_P50=%d\n", lastSec.LatencyP50.Nanoseconds()/1000)
		fmt.Printf("LATENCY_P75=%d\n", lastSec.LatencyP75.Nanoseconds()/1000)
		fmt.Printf("LATENCY_P90=%d\n", lastSec.LatencyP90.Nanoseconds()/1000)
		fmt.Printf("LATENCY_P95=%d\n", lastSec.LatencyP95.Nanoseconds()/1000)
		fmt.Printf("LATENCY_P99=%d\n", lastSec.LatencyP99.Nanoseconds()/1000)
	}

	return nil
}

// Utility functions
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
} 