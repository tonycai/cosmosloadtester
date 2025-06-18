# Cosmos Load Tester CLI

A powerful terminal-based load testing tool for Cosmos blockchain applications. This CLI version preserves all the core features of the web-based cosmosloadtester while providing a streamlined command-line interface.

## 🚀 Quick Start

### Installation

1. **Build the CLI tool:**
   ```bash
   make cli
   ```

2. **Install globally (optional):**
   ```bash
   make install-cli
   ```

3. **Run your first load test:**
   ```bash
   ./bin/cosmosloadtester-cli --endpoints="ws://localhost:26657/websocket" --duration=30s --rate=100
   ```

## 📋 Command Reference

### Basic Usage

```bash
cosmosloadtester-cli [flags]
```

### Core Load Testing Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--client-factory` | Client factory for transaction generation | `test-cosmos-client-factory` | `--client-factory=aiw3defi-bank-send` |
| `--endpoints` | Comma-separated RPC endpoints | Required | `--endpoints="ws://localhost:26657/websocket,http://localhost:26657"` |
| `--duration` | Load test duration | `60s` | `--duration=2m30s` |
| `--rate` | Transactions per second per connection | `1000` | `--rate=5000` |
| `--connections` | Connections per endpoint | `1` | `--connections=4` |
| `--size` | Transaction size in bytes | `250` | `--size=512` |
| `--count` | Max transactions (-1 = unlimited) | `-1` | `--count=10000` |
| `--broadcast-method` | Broadcast method | `sync` | `--broadcast-method=async` |

### Profile Management

| Flag | Description | Example |
|------|-------------|---------|
| `--profile` | Use saved configuration profile | `--profile=local-testnet` |
| `--save-profile` | Save current config as profile | `--save-profile=my-config` |
| `--list-profiles` | List all saved profiles | `--list-profiles` |
| `--show-profile` | Show profile details | `--show-profile=high-throughput` |
| `--delete-profile` | Delete a profile | `--delete-profile=old-config` |
| `--generate-template` | Generate template profile | `--generate-template=local-testnet` |

### Output Formats

| Flag | Description | Output |
|------|-------------|--------|
| `--output-format=live` | Interactive live output (default) | Colored terminal output with progress |
| `--output-format=json` | JSON formatted results | Machine-readable JSON |
| `--output-format=csv` | CSV formatted results | Spreadsheet-compatible CSV |
| `--output-format=summary` | Key-value summary | Shell script friendly format |

### Utility Commands

| Flag | Description | Example |
|------|-------------|---------|
| `--interactive` | Run in interactive mode | `--interactive` |
| `--validate` | Validate configuration only | `--validate --profile=test` |
| `--dry-run` | Show config without running | `--dry-run --rate=5000` |
| `--check-endpoints` | Test endpoint connectivity | `--check-endpoints --profile=prod` |
| `--benchmark` | Run predefined benchmarks | `--benchmark=stress` |
| `--list-factories` | List available client factories | `--list-factories` |

## 🎯 Example Usage Scenarios

### 1. Basic Load Test

```bash
# Test local testnet with 100 TPS for 30 seconds
cosmosloadtester-cli \
  --endpoints="ws://localhost:26657/websocket" \
  --duration=30s \
  --rate=100 \
  --connections=2
```

### 2. High-Throughput Stress Test

```bash
# Stress test with async broadcasts
cosmosloadtester-cli \
  --endpoints="ws://localhost:26657/websocket" \
  --duration=2m \
  --rate=5000 \
  --connections=10 \
  --broadcast-method=async \
  --size=40
```

### 3. Latency Measurement

```bash
# Low-rate test with commit broadcasts for latency measurement
cosmosloadtester-cli \
  --endpoints="http://localhost:26657" \
  --duration=1m \
  --rate=10 \
  --connections=1 \
  --broadcast-method=commit
```

### 4. Multi-Endpoint Load Balancing

```bash
# Test across multiple endpoints
cosmosloadtester-cli \
  --endpoints="ws://node1:26657/websocket,ws://node2:26657/websocket,http://node3:26657" \
  --endpoint-select-method=any \
  --duration=5m \
  --rate=1000 \
  --connections=3
```

### 5. AIW3 DeFi Testing

```bash
# Test with AIW3 DeFi bank send transactions
cosmosloadtester-cli \
  --client-factory=aiw3defi-bank-send \
  --endpoints="ws://localhost:26657/websocket" \
  --duration=1m \
  --rate=500 \
  --size=512
```

## 📊 Configuration Profiles

### Creating Profiles

Save frequently used configurations as profiles:

```bash
# Save current configuration
cosmosloadtester-cli \
  --endpoints="ws://localhost:26657/websocket" \
  --rate=1000 \
  --duration=1m \
  --save-profile=local-dev

# Use saved profile
cosmosloadtester-cli --profile=local-dev
```

### Built-in Templates

Generate common configuration templates:

```bash
# Available templates
cosmosloadtester-cli --generate-template=local-testnet
cosmosloadtester-cli --generate-template=high-throughput
cosmosloadtester-cli --generate-template=latency-test
cosmosloadtester-cli --generate-template=multi-endpoint
cosmosloadtester-cli --generate-template=aiw3defi-test
```

### Profile Management

```bash
# List all profiles
cosmosloadtester-cli --list-profiles

# Show profile details
cosmosloadtester-cli --show-profile=local-testnet

# Delete profile
cosmosloadtester-cli --delete-profile=old-config

# Export profiles to file
cosmosloadtester-cli --export-profiles=my-profiles.yaml

# Import profiles from file
cosmosloadtester-cli --import-profiles=shared-profiles.yaml
```

## 🔧 Interactive Mode

Launch interactive mode for guided configuration:

```bash
cosmosloadtester-cli --interactive
```

Available interactive commands:
1. **Run load test** - Configure and run test interactively
2. **List profiles** - Browse saved profiles
3. **Create profile** - Step-by-step profile creation
4. **Load profile** - Select and run existing profile
5. **Check endpoints** - Test endpoint connectivity
6. **Generate template** - Create template profiles

## 📈 Benchmark Suites

Run predefined benchmark suites:

### Quick Benchmark (10 seconds)
```bash
cosmosloadtester-cli --benchmark=quick --endpoints="ws://localhost:26657/websocket"
```

### Standard Benchmark (sync + async tests)
```bash
cosmosloadtester-cli --benchmark=standard --endpoints="ws://localhost:26657/websocket"
```

### Stress Benchmark (high throughput)
```bash
cosmosloadtester-cli --benchmark=stress --endpoints="ws://localhost:26657/websocket"
```

## 📊 Output Formats

### Live Output (Default)
- Real-time progress bar
- Colored terminal output
- Detailed statistics display
- Latency percentiles

### JSON Output
```bash
cosmosloadtester-cli --profile=test --output-format=json > results.json
```

### CSV Output
```bash
cosmosloadtester-cli --profile=test --output-format=csv > results.csv
```

### Summary Output (Shell-friendly)
```bash
# Capture specific metrics
TOTAL_TPS=$(cosmosloadtester-cli --profile=test --output-format=summary | grep AVG_TPS | cut -d'=' -f2)
echo "Achieved TPS: $TOTAL_TPS"
```

## 🔍 Validation and Debugging

### Configuration Validation
```bash
# Validate without running
cosmosloadtester-cli --validate --profile=production
```

### Dry Run
```bash
# Preview configuration
cosmosloadtester-cli --dry-run --rate=5000 --duration=5m
```

### Endpoint Health Check
```bash
# Test connectivity before load testing
cosmosloadtester-cli --check-endpoints --profile=production
```

### Debug Logging
```bash
# Enable debug output
cosmosloadtester-cli --log-level=debug --profile=test
```

## 🎨 Client Factories

### Available Factories

1. **test-cosmos-client-factory** - Default empty transactions
2. **aiw3defi-bank-send** - AIW3 DeFi bank send transactions

List all available factories:
```bash
cosmosloadtester-cli --list-factories
```

### Creating Custom Client Factories

Follow the same process as the web version. After creating your factory in the `clients/` directory and registering it in `cmd/server/main.go`, rebuild the CLI:

```bash
make cli
```

## 🔗 Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/load-test.yml
- name: Run Load Test
  run: |
    ./bin/cosmosloadtester-cli \
      --profile=ci-test \
      --output-format=json \
      --quiet > load-test-results.json
    
    # Extract metrics for assertions
    TPS=$(jq '.avg_txs_per_second' load-test-results.json)
    if (( $(echo "$TPS < 1000" | bc -l) )); then
      echo "Performance regression detected!"
      exit 1
    fi
```

### Monitoring Integration

```bash
#!/bin/bash
# monitoring-script.sh

# Run load test and extract metrics
RESULTS=$(cosmosloadtester-cli --profile=monitoring --output-format=summary --quiet)

# Extract metrics
TOTAL_TPS=$(echo "$RESULTS" | grep AVG_TPS | cut -d'=' -f2)
LATENCY_P95=$(echo "$RESULTS" | grep LATENCY_P95 | cut -d'=' -f2)

# Send to monitoring system
curl -X POST http://monitoring-system/metrics \
  -d "tps=$TOTAL_TPS&latency_p95=$LATENCY_P95"
```

### Performance Comparison

```bash
#!/bin/bash
# compare-performance.sh

echo "Running baseline test..."
cosmosloadtester-cli --profile=baseline --output-format=summary --quiet > baseline.txt

echo "Running optimized test..."
cosmosloadtester-cli --profile=optimized --output-format=summary --quiet > optimized.txt

# Compare results
BASELINE_TPS=$(grep AVG_TPS baseline.txt | cut -d'=' -f2)
OPTIMIZED_TPS=$(grep AVG_TPS optimized.txt | cut -d'=' -f2)

IMPROVEMENT=$(echo "scale=2; ($OPTIMIZED_TPS - $BASELINE_TPS) / $BASELINE_TPS * 100" | bc)
echo "Performance improvement: ${IMPROVEMENT}%"
```

## 🛠️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/orijtech/cosmosloadtester
cd cosmosloadtester

# Install dependencies
go mod download

# Build CLI
make cli

# Run tests
make test
```

### Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Submit pull request

## 🐛 Troubleshooting

### Common Issues

1. **"Client factory not found"**
   ```bash
   # List available factories
   cosmosloadtester-cli --list-factories
   ```

2. **"Endpoint connection failed"**
   ```bash
   # Test connectivity
   cosmosloadtester-cli --check-endpoints --endpoints="ws://localhost:26657/websocket"
   ```

3. **"Configuration validation failed"**
   ```bash
   # Validate configuration
   cosmosloadtester-cli --validate --profile=your-profile
   ```

4. **Permission denied when saving profiles**
   ```bash
   # Check permissions on config directory
   ls -la ~/.cosmosloadtester/
   ```

### Debug Mode

Enable detailed logging:
```bash
cosmosloadtester-cli --log-level=debug --profile=test
```

### Getting Help

- View all available flags: `cosmosloadtester-cli --help`
- Check version: `cosmosloadtester-cli --version`
- Join discussions: [GitHub Issues](https://github.com/orijtech/cosmosloadtester/issues)

## 🏆 Advanced Features

### Protocol Auto-Detection

The CLI automatically detects protocols based on endpoint URLs:
- `ws://` or `wss://` → WebSocket transactor
- `http://` or `https://` → HTTP transactor

### Graceful Shutdown

Press `Ctrl+C` to gracefully stop load tests and display partial results.

### Resource Management

The CLI automatically:
- Creates temporary stats files
- Cleans up resources on exit
- Manages connection pools efficiently

### Cross-Platform Support

Built binaries support:
- Linux (amd64, arm64)
- macOS (amd64, arm64)  
- Windows (amd64)

## 📄 License

This project is licensed under the same terms as the main cosmosloadtester project.

---

*Happy load testing! 🚀* 