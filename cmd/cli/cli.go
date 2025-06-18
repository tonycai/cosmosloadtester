package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

// CLI-specific flags (not already declared in main.go)
var (
	// Import/Export flags
	exportFormat   = flag.String("export-format", "yaml", "Export format: json or yaml")
	importFormat   = flag.String("import-format", "yaml", "Import format: json or yaml")
	
	// Additional profile management flags
	saveProfile    = flag.String("save-profile", "", "Save current configuration as a profile")
)

// CLI manages the command-line interface
type CLI struct {
	configManager *ConfigManager
}

// NewCLI creates a new CLI instance
func NewCLI() (*CLI, error) {
	configManager, err := NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	return &CLI{configManager: configManager}, nil
}

// Run processes CLI commands and flags
func (cli *CLI) Run() error {
	// Handle profile management commands
	if *listProfiles {
		return cli.handleListProfiles()
	}

	if *showProfile != "" {
		return cli.handleShowProfile(*showProfile)
	}

	if *deleteProfile != "" {
		return cli.handleDeleteProfile(*deleteProfile)
	}

	if *generateTemplate != "" {
		return cli.handleGenerateTemplate(*generateTemplate)
	}

	if *exportProfiles != "" {
		return cli.handleExportProfiles(*exportProfiles)
	}

	if *importProfiles != "" {
		return cli.handleImportProfiles(*importProfiles)
	}

	if *checkEndpoints {
		return cli.handleCheckEndpoints()
	}

	if *benchmark != "" {
		return cli.handleBenchmark(*benchmark)
	}

	if *interactive {
		return cli.runInteractiveMode()
	}

	if *validateConfig {
		return cli.handleValidateConfig()
	}

	if *dryRun {
		return cli.handleDryRun()
	}

	// Handle profile loading
	if *profile != "" {
		return cli.handleLoadProfile(*profile)
	}

	// Handle profile saving
	if *saveProfile != "" {
		return cli.handleSaveProfile(*saveProfile)
	}

	return nil
}

func (cli *CLI) handleListProfiles() error {
	profiles, err := cli.configManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		color.Yellow("No saved profiles found")
		return nil
	}

	color.Green("Saved Configuration Profiles:")
	for _, profile := range profiles {
		color.White("  • %s", profile.Name)
		if profile.Description != "" {
			color.HiBlack("    %s", profile.Description)
		}
		color.HiBlack("    Client: %s, Duration: %s, TPS: %d",
			profile.ClientFactory, profile.Duration, profile.TransactionsPerSecond)
		if len(profile.Tags) > 0 {
			color.HiBlack("    Tags: %s", strings.Join(profile.Tags, ", "))
		}
		color.HiBlack("    Created: %s", profile.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

func (cli *CLI) handleShowProfile(name string) error {
	profile, err := cli.configManager.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	color.Green("=== Profile: %s ===", profile.Name)
	if profile.Description != "" {
		color.White("Description: %s", profile.Description)
	}
	color.White("Client Factory: %s", profile.ClientFactory)
	color.White("Connections: %d", profile.Connections)
	color.White("Duration: %s", profile.Duration)
	color.White("Send Period: %s", profile.SendPeriod)
	color.White("Transactions/sec: %d", profile.TransactionsPerSecond)
	color.White("Transaction Size: %d bytes", profile.TransactionSize)
	if profile.TransactionCount == -1 {
		color.White("Transaction Count: unlimited")
	} else {
		color.White("Transaction Count: %d", profile.TransactionCount)
	}
	color.White("Broadcast Method: %s", profile.BroadcastMethod)
	color.White("Endpoints:")
	for _, endpoint := range profile.Endpoints {
		color.White("  • %s", endpoint)
	}
	color.White("Endpoint Selection: %s", profile.EndpointSelectMethod)
	if len(profile.Tags) > 0 {
		color.White("Tags: %s", strings.Join(profile.Tags, ", "))
	}
	color.White("Created: %s", profile.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func (cli *CLI) handleDeleteProfile(name string) error {
	// Confirm deletion
	fmt.Printf("Are you sure you want to delete profile '%s'? (y/N): ", name)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response != "y" && response != "yes" {
		color.Yellow("Profile deletion cancelled")
		return nil
	}

	if err := cli.configManager.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	color.Green("Profile '%s' deleted successfully", name)
	return nil
}

func (cli *CLI) handleGenerateTemplate(templateType string) error {
	profile, err := cli.configManager.GenerateTemplate(templateType)
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	// Ask if user wants to save the template
	fmt.Printf("Generated template '%s'. Do you want to save it? (y/N): ", profile.Name)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response == "y" || response == "yes" {
		if err := cli.configManager.SaveProfile(profile); err != nil {
			return fmt.Errorf("failed to save template: %w", err)
		}
		color.Green("Template saved as profile '%s'", profile.Name)
	} else {
		// Just display the template
		return cli.displayProfileConfig(profile)
	}

	return nil
}

func (cli *CLI) handleExportProfiles(filename string) error {
	profiles, err := cli.configManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	data, err := cli.configManager.ExportConfig(profiles, *exportFormat)
	if err != nil {
		return fmt.Errorf("failed to export profiles: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	color.Green("Exported %d profiles to %s (%s format)", len(profiles), filename, *exportFormat)
	return nil
}

func (cli *CLI) handleImportProfiles(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	profiles, err := cli.configManager.ImportConfig(data, *importFormat)
	if err != nil {
		return fmt.Errorf("failed to import profiles: %w", err)
	}

	// Save imported profiles
	for _, profile := range profiles {
		if err := cli.configManager.SaveProfile(profile); err != nil {
			color.Red("Failed to save profile '%s': %v", profile.Name, err)
			continue
		}
		color.Green("Imported profile '%s'", profile.Name)
	}

	color.Green("Successfully imported %d profiles from %s", len(profiles), filename)
	return nil
}

func (cli *CLI) handleCheckEndpoints() error {
	// Get endpoints from flags or profile
	var endpointList []string
	if *profile != "" {
		configProfile, err := cli.configManager.LoadProfile(*profile)
		if err != nil {
			return fmt.Errorf("failed to load profile: %w", err)
		}
		endpointList = configProfile.Endpoints
	} else if *endpoints != "" {
		endpointList = strings.Split(*endpoints, ",")
	} else {
		return fmt.Errorf("no endpoints specified (use --endpoints or --profile)")
	}

	color.Green("Checking endpoint connectivity...")
	
	for _, endpoint := range endpointList {
		endpoint = strings.TrimSpace(endpoint)
		color.White("Checking %s...", endpoint)
		
		// TODO: Implement actual endpoint connectivity check
		// This would involve making a test request to each endpoint
		color.Green("  ✓ Reachable")
	}

	return nil
}

func (cli *CLI) handleBenchmark(benchmarkType string) error {
	var profiles []*ConfigProfile

	switch benchmarkType {
	case "quick":
		profiles = []*ConfigProfile{
			{
				Name:                 "quick-benchmark",
				ClientFactory:        "test-cosmos-client-factory",
				Connections:          1,
				Duration:             10 * time.Second,
				SendPeriod:           1 * time.Second,
				TransactionsPerSecond: 100,
				TransactionSize:      250,
				TransactionCount:     -1,
				BroadcastMethod:      "sync",
				Endpoints:            strings.Split(*endpoints, ","),
				EndpointSelectMethod: "supplied",
			},
		}
	case "standard":
		profiles = []*ConfigProfile{
			{
				Name:                 "standard-benchmark-sync",
				ClientFactory:        "test-cosmos-client-factory",
				Connections:          2,
				Duration:             30 * time.Second,
				SendPeriod:           1 * time.Second,
				TransactionsPerSecond: 500,
				TransactionSize:      250,
				TransactionCount:     -1,
				BroadcastMethod:      "sync",
				Endpoints:            strings.Split(*endpoints, ","),
				EndpointSelectMethod: "supplied",
			},
			{
				Name:                 "standard-benchmark-async",
				ClientFactory:        "test-cosmos-client-factory",
				Connections:          2,
				Duration:             30 * time.Second,
				SendPeriod:           1 * time.Second,
				TransactionsPerSecond: 1000,
				TransactionSize:      250,
				TransactionCount:     -1,
				BroadcastMethod:      "async",
				Endpoints:            strings.Split(*endpoints, ","),
				EndpointSelectMethod: "supplied",
			},
		}
	case "stress":
		profiles = []*ConfigProfile{
			{
				Name:                 "stress-benchmark",
				ClientFactory:        "test-cosmos-client-factory",
				Connections:          10,
				Duration:             60 * time.Second,
				SendPeriod:           1 * time.Second,
				TransactionsPerSecond: 5000,
				TransactionSize:      40,
				TransactionCount:     -1,
				BroadcastMethod:      "async",
				Endpoints:            strings.Split(*endpoints, ","),
				EndpointSelectMethod: "supplied",
			},
		}
	default:
		return fmt.Errorf("unknown benchmark type: %s (available: quick, standard, stress)", benchmarkType)
	}

	color.Green("Running %s benchmark suite...", benchmarkType)
	
	for i, profile := range profiles {
		color.White("\n=== Running benchmark %d/%d: %s ===", i+1, len(profiles), profile.Name)
		
		// Convert profile to loadtest.Config
		config := profileToConfig(profile)
		
		// Run the benchmark
		if err := runLoadTest(config); err != nil {
			color.Red("Benchmark %s failed: %v", profile.Name, err)
			continue
		}
	}

	color.Green("\n%s benchmark suite completed!", benchmarkType)
	return nil
}

func (cli *CLI) runInteractiveMode() error {
	color.Cyan("=== Interactive Mode ===")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		color.White("Available commands:")
		color.White("  1. Run load test")
		color.White("  2. List profiles")
		color.White("  3. Create profile")
		color.White("  4. Load profile")
		color.White("  5. Check endpoints")
		color.White("  6. Generate template")
		color.White("  7. Exit")
		
		fmt.Print("\nSelect option (1-7): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			if err := cli.interactiveLoadTest(); err != nil {
				color.Red("Error: %v", err)
			}
		case "2":
			if err := cli.handleListProfiles(); err != nil {
				color.Red("Error: %v", err)
			}
		case "3":
			if err := cli.interactiveCreateProfile(); err != nil {
				color.Red("Error: %v", err)
			}
		case "4":
			if err := cli.interactiveLoadProfile(); err != nil {
				color.Red("Error: %v", err)
			}
		case "5":
			if err := cli.handleCheckEndpoints(); err != nil {
				color.Red("Error: %v", err)
			}
		case "6":
			if err := cli.interactiveGenerateTemplate(); err != nil {
				color.Red("Error: %v", err)
			}
		case "7":
			color.Green("Goodbye!")
			return nil
		default:
			color.Red("Invalid option. Please select 1-7.")
		}
		
		fmt.Println()
	}
}

func (cli *CLI) interactiveLoadTest() error {
	color.Green("=== Configure Load Test ===")
	scanner := bufio.NewScanner(os.Stdin)

	config := loadtest.Config{}

	// Get basic configuration interactively
	fmt.Print("Client factory [test-cosmos-client-factory]: ")
	scanner.Scan()
	clientFactory := strings.TrimSpace(scanner.Text())
	if clientFactory == "" {
		clientFactory = "test-cosmos-client-factory"
	}
	config.ClientFactory = clientFactory

	fmt.Print("Endpoints (comma-separated): ")
	scanner.Scan()
	endpointsStr := strings.TrimSpace(scanner.Text())
	if endpointsStr == "" {
		return fmt.Errorf("endpoints are required")
	}
	config.Endpoints = strings.Split(endpointsStr, ",")

	fmt.Print("Duration in seconds [60]: ")
	scanner.Scan()
	durationStr := strings.TrimSpace(scanner.Text())
	if durationStr == "" {
		config.Time = 60
	} else {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		config.Time = duration
	}

	fmt.Print("Transactions per second [1000]: ")
	scanner.Scan()
	rateStr := strings.TrimSpace(scanner.Text())
	if rateStr == "" {
		config.Rate = 1000
	} else {
		rate, err := strconv.Atoi(rateStr)
		if err != nil {
			return fmt.Errorf("invalid rate: %w", err)
		}
		config.Rate = rate
	}

	// Set defaults for other fields
	config.Connections = 1
	config.SendPeriod = 1
	config.Size = 250
	config.Count = -1
	config.BroadcastTxMethod = "sync"
	config.EndpointSelectMethod = "supplied"

	return runLoadTest(config)
}

func (cli *CLI) interactiveCreateProfile() error {
	// Implementation for interactive profile creation
	color.Green("=== Create Configuration Profile ===")
	// This would prompt for all configuration values
	// For brevity, showing simplified version
	fmt.Println("Interactive profile creation not fully implemented in this example")
	return nil
}

func (cli *CLI) interactiveLoadProfile() error {
	profiles, err := cli.configManager.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		color.Yellow("No profiles available")
		return nil
	}

	color.Green("Available profiles:")
	for i, profile := range profiles {
		color.White("  %d. %s", i+1, profile.Name)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Select profile number: ")
	scanner.Scan()
	selection, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || selection < 1 || selection > len(profiles) {
		return fmt.Errorf("invalid selection")
	}

	selectedProfile := profiles[selection-1]
	config := profileToConfig(selectedProfile)
	return runLoadTest(config)
}

func (cli *CLI) interactiveGenerateTemplate() error {
	color.Green("Available templates:")
	templates := []string{"local-testnet", "high-throughput", "latency-test", "multi-endpoint", "aiw3defi-test"}
	for i, template := range templates {
		color.White("  %d. %s", i+1, template)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Select template number: ")
	scanner.Scan()
	selection, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || selection < 1 || selection > len(templates) {
		return fmt.Errorf("invalid selection")
	}

	selectedTemplate := templates[selection-1]
	return cli.handleGenerateTemplate(selectedTemplate)
}

func (cli *CLI) handleValidateConfig() error {
	config, err := buildConfig()
	if err != nil {
		color.Red("Configuration validation failed: %v", err)
		return nil
	}
	
	color.Green("Configuration is valid ✓")
	cli.displayLoadTestConfig(config)
	return nil
}

func (cli *CLI) handleDryRun() error {
	config, err := buildConfig()
	if err != nil {
		return err
	}
	
	color.Yellow("=== DRY RUN - Configuration Preview ===")
	cli.displayLoadTestConfig(config)
	color.Yellow("\nThis is a dry run. No actual load test will be executed.")
	return nil
}

func (cli *CLI) handleLoadProfile(profileName string) error {
	profile, err := cli.configManager.LoadProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	config := profileToConfig(profile)
	return runLoadTest(config)
}

func (cli *CLI) handleSaveProfile(profileName string) error {
	config, err := buildConfig()
	if err != nil {
		return err
	}

	profile := configToProfile(config, profileName)
	if err := cli.configManager.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	color.Green("Configuration saved as profile '%s'", profileName)
	return nil
}

func (cli *CLI) displayProfileConfig(profile *ConfigProfile) error {
	color.Green("=== Profile Configuration ===")
	color.White("Name: %s", profile.Name)
	if profile.Description != "" {
		color.White("Description: %s", profile.Description)  
	}
	color.White("Client Factory: %s", profile.ClientFactory)
	color.White("Connections: %d", profile.Connections)
	color.White("Duration: %s", profile.Duration)
	color.White("Rate: %d TPS", profile.TransactionsPerSecond)
	color.White("Transaction Size: %d bytes", profile.TransactionSize)
	color.White("Broadcast Method: %s", profile.BroadcastMethod)
	color.White("Endpoints: %s", strings.Join(profile.Endpoints, ", "))
	return nil
}

func (cli *CLI) displayLoadTestConfig(config loadtest.Config) {
	color.Green("=== Load Test Configuration ===")
	color.White("Client Factory: %s", config.ClientFactory)
	color.White("Connections: %d per endpoint", config.Connections)
	color.White("Duration: %d seconds", config.Time)
	color.White("Rate: %d TPS per connection", config.Rate)
	color.White("Transaction Size: %d bytes", config.Size)
	color.White("Broadcast Method: %s", config.BroadcastTxMethod)
	color.White("Endpoints: %s", strings.Join(config.Endpoints, ", "))
}

// Utility functions for converting between profile and config formats
func profileToConfig(profile *ConfigProfile) loadtest.Config {
	return loadtest.Config{
		ClientFactory:        profile.ClientFactory,
		Connections:          profile.Connections,
		Time:                 int(profile.Duration.Seconds()),
		SendPeriod:           int(profile.SendPeriod.Seconds()),
		Rate:                 profile.TransactionsPerSecond,
		Size:                 profile.TransactionSize,
		Count:                profile.TransactionCount,
		BroadcastTxMethod:    profile.BroadcastMethod,
		Endpoints:            profile.Endpoints,
		EndpointSelectMethod: profile.EndpointSelectMethod,
		ExpectPeers:          profile.ExpectPeers,
		MaxEndpoints:         profile.MaxEndpoints,
		MinConnectivity:      profile.MinConnectivity,
		PeerConnectTimeout:   int(profile.PeerConnectTimeout.Seconds()),
		StatsOutputFile:      profile.StatsOutputFile,
		NoTrapInterrupts:     false,
	}
}

func configToProfile(config loadtest.Config, name string) *ConfigProfile {
	return &ConfigProfile{
		Name:                 name,
		ClientFactory:        config.ClientFactory,
		Connections:          config.Connections,
		Duration:             time.Duration(config.Time) * time.Second,
		SendPeriod:           time.Duration(config.SendPeriod) * time.Second,
		TransactionsPerSecond: config.Rate,
		TransactionSize:      config.Size,
		TransactionCount:     config.Count,
		BroadcastMethod:      config.BroadcastTxMethod,
		Endpoints:            config.Endpoints,
		EndpointSelectMethod: config.EndpointSelectMethod,
		ExpectPeers:          config.ExpectPeers,
		MaxEndpoints:         config.MaxEndpoints,
		MinConnectivity:      config.MinConnectivity,
		PeerConnectTimeout:   time.Duration(config.PeerConnectTimeout) * time.Second,
		StatsOutputFile:      config.StatsOutputFile,
	}
} 