package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigProfile represents a saved configuration profile
type ConfigProfile struct {
	Name                 string        `yaml:"name" json:"name"`
	Description          string        `yaml:"description,omitempty" json:"description,omitempty"`
	ClientFactory        string        `yaml:"client_factory" json:"client_factory"`
	Connections          int           `yaml:"connections" json:"connections"`
	Duration             time.Duration `yaml:"duration" json:"duration"`
	SendPeriod           time.Duration `yaml:"send_period" json:"send_period"`
	TransactionsPerSecond int           `yaml:"transactions_per_second" json:"transactions_per_second"`
	TransactionSize      int           `yaml:"transaction_size" json:"transaction_size"`
	TransactionCount     int           `yaml:"transaction_count" json:"transaction_count"`
	BroadcastMethod      string        `yaml:"broadcast_method" json:"broadcast_method"`
	Endpoints            []string      `yaml:"endpoints" json:"endpoints"`
	EndpointSelectMethod string        `yaml:"endpoint_select_method" json:"endpoint_select_method"`
	ExpectPeers          int           `yaml:"expect_peers" json:"expect_peers"`
	MaxEndpoints         int           `yaml:"max_endpoints" json:"max_endpoints"`
	MinConnectivity      int           `yaml:"min_connectivity" json:"min_connectivity"`
	PeerConnectTimeout   time.Duration `yaml:"peer_connect_timeout" json:"peer_connect_timeout"`
	StatsOutputFile      string        `yaml:"stats_output_file,omitempty" json:"stats_output_file,omitempty"`
	Tags                 []string      `yaml:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt            time.Time     `yaml:"created_at" json:"created_at"`
}

// ConfigManager handles configuration profiles
type ConfigManager struct {
	configDir string
}

// NewConfigManager creates a new configuration  manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".cosmosloadtester")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &ConfigManager{configDir: configDir}, nil
}

// SaveProfile saves a configuration profile
func (cm *ConfigManager) SaveProfile(profile *ConfigProfile) error {
	profile.CreatedAt = time.Now()
	filename := filepath.Join(cm.configDir, profile.Name+".yaml")
	
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// LoadProfile loads a configuration profile
func (cm *ConfigManager) LoadProfile(name string) (*ConfigProfile, error) {
	filename := filepath.Join(cm.configDir, name+".yaml")
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile ConfigProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles lists all available configuration profiles
func (cm *ConfigManager) ListProfiles() ([]*ConfigProfile, error) {
	files, err := filepath.Glob(filepath.Join(cm.configDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list profile files: %w", err)
	}

	var profiles []*ConfigProfile
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var profile ConfigProfile
		if err := yaml.Unmarshal(data, &profile); err != nil {
			continue // Skip files that can't be parsed
		}

		profiles = append(profiles, &profile)
	}

	return profiles, nil
}

// DeleteProfile deletes a configuration profile
func (cm *ConfigManager) DeleteProfile(name string) error {
	filename := filepath.Join(cm.configDir, name+".yaml")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// GenerateTemplate generates common configuration templates
func (cm *ConfigManager) GenerateTemplate(templateType string) (*ConfigProfile, error) {
	switch templateType {
	case "local-testnet":
		return &ConfigProfile{
			Name:                 "local-testnet",
			Description:          "Local testnet configuration for development",
			ClientFactory:        "test-cosmos-client-factory",
			Connections:          4,
			Duration:             30 * time.Second,
			SendPeriod:           1 * time.Second,
			TransactionsPerSecond: 100,
			TransactionSize:      250,
			TransactionCount:     -1,
			BroadcastMethod:      "sync",
			Endpoints:            []string{"ws://localhost:26657/websocket", "http://localhost:26657"},
			EndpointSelectMethod: "supplied",
			ExpectPeers:          0,
			MaxEndpoints:         0,
			MinConnectivity:      0,
			PeerConnectTimeout:   5 * time.Second,
			Tags:                 []string{"local", "development"},
		}, nil

	case "high-throughput":
		return &ConfigProfile{
			Name:                 "high-throughput",
			Description:          "High throughput stress testing configuration",
			ClientFactory:        "test-cosmos-client-factory",
			Connections:          10,
			Duration:             120 * time.Second,
			SendPeriod:           1 * time.Second,
			TransactionsPerSecond: 5000,
			TransactionSize:      40,
			TransactionCount:     -1,
			BroadcastMethod:      "async",
			Endpoints:            []string{"ws://localhost:26657/websocket"},
			EndpointSelectMethod: "supplied",
			ExpectPeers:          0,
			MaxEndpoints:         0,
			MinConnectivity:      0,
			PeerConnectTimeout:   10 * time.Second,
			Tags:                 []string{"stress", "high-throughput"},
		}, nil

	case "latency-test":
		return &ConfigProfile{
			Name:                 "latency-test",
			Description:          "Latency measurement configuration",
			ClientFactory:        "test-cosmos-client-factory",
			Connections:          1,
			Duration:             60 * time.Second,
			SendPeriod:           1 * time.Second,
			TransactionsPerSecond: 10,
			TransactionSize:      250,
			TransactionCount:     -1,
			BroadcastMethod:      "commit",
			Endpoints:            []string{"http://localhost:26657"},
			EndpointSelectMethod: "supplied",
			ExpectPeers:          0,
			MaxEndpoints:         0,
			MinConnectivity:      0,
			PeerConnectTimeout:   5 * time.Second,
			Tags:                 []string{"latency", "measurement"},
		}, nil

	case "multi-endpoint":
		return &ConfigProfile{
			Name:                 "multi-endpoint",
			Description:          "Multi-endpoint load balancing test",
			ClientFactory:        "test-cosmos-client-factory",
			Connections:          2,
			Duration:             90 * time.Second,
			SendPeriod:           1 * time.Second,
			TransactionsPerSecond: 1000,
			TransactionSize:      250,
			TransactionCount:     -1,
			BroadcastMethod:      "sync",
			Endpoints: []string{
				"ws://node1.example.com:26657/websocket",
				"ws://node2.example.com:26657/websocket",
				"http://node3.example.com:26657",
			},
			EndpointSelectMethod: "any",
			ExpectPeers:          0,
			MaxEndpoints:         0,
			MinConnectivity:      0,
			PeerConnectTimeout:   5 * time.Second,
			Tags:                 []string{"multi-endpoint", "load-balancing"},
		}, nil

	case "aiw3defi-test":
		return &ConfigProfile{
			Name:                 "aiw3defi-test",
			Description:          "AIW3 DeFi bank send transactions test",
			ClientFactory:        "aiw3defi-bank-send",
			Connections:          5,
			Duration:             60 * time.Second,
			SendPeriod:           1 * time.Second,
			TransactionsPerSecond: 500,
			TransactionSize:      512,
			TransactionCount:     -1,
			BroadcastMethod:      "sync",
			Endpoints:            []string{"ws://localhost:26657/websocket"},
			EndpointSelectMethod: "supplied",
			ExpectPeers:          0,
			MaxEndpoints:         0,
			MinConnectivity:      0,
			PeerConnectTimeout:   5 * time.Second,
			Tags:                 []string{"aiw3", "defi", "bank-send"},
		}, nil

	default:
		return nil, fmt.Errorf("unknown template type: %s", templateType)
	}
}

// ValidateConfig validates a configuration profile
func ValidateConfig(profile *ConfigProfile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.ClientFactory == "" {
		return fmt.Errorf("client factory is required")
	}

	if profile.Connections <= 0 {
		return fmt.Errorf("connections must be greater than 0")
	}

	if profile.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if profile.SendPeriod <= 0 {
		return fmt.Errorf("send period must be greater than 0")
	}

	if profile.TransactionsPerSecond <= 0 {
		return fmt.Errorf("transactions per second must be greater than 0")
	}

	if profile.TransactionSize < 40 {
		return fmt.Errorf("transaction size must be at least 40 bytes")
	}

	if len(profile.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}

	// Validate endpoints
	for _, endpoint := range profile.Endpoints {
		if !strings.HasPrefix(endpoint, "ws://") && 
		   !strings.HasPrefix(endpoint, "wss://") &&
		   !strings.HasPrefix(endpoint, "http://") &&
		   !strings.HasPrefix(endpoint, "https://") {
			return fmt.Errorf("invalid endpoint protocol: %s (must start with ws://, wss://, http://, or https://)", endpoint)
		}
	}

	// Validate broadcast method
	validBroadcastMethods := map[string]bool{
		"sync":   true,
		"async":  true,
		"commit": true,
	}
	if !validBroadcastMethods[profile.BroadcastMethod] {
		return fmt.Errorf("invalid broadcast method: %s (valid: sync, async, commit)", profile.BroadcastMethod)
	}

	// Validate endpoint select method
	validEndpointSelectMethods := map[string]bool{
		"supplied":   true,
		"discovered": true,
		"any":        true,
	}
	if !validEndpointSelectMethods[profile.EndpointSelectMethod] {
		return fmt.Errorf("invalid endpoint select method: %s (valid: supplied, discovered, any)", profile.EndpointSelectMethod)
	}

	return nil
}

// ExportConfig exports configuration profiles to various formats
func (cm *ConfigManager) ExportConfig(profiles []*ConfigProfile, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(profiles, "", "  ")
	case "yaml":
		return yaml.Marshal(profiles)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ImportConfig imports configuration profiles from various formats
func (cm *ConfigManager) ImportConfig(data []byte, format string) ([]*ConfigProfile, error) {
	var profiles []*ConfigProfile

	switch format {
	case "json":
		if err := json.Unmarshal(data, &profiles); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &profiles); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}

	// Validate imported profiles
	for _, profile := range profiles {
		if err := ValidateConfig(profile); err != nil {
			return nil, fmt.Errorf("invalid profile %s: %w", profile.Name, err)
		}
	}

	return profiles, nil
} 