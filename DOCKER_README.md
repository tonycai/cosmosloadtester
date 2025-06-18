# CosmosLoadTester CLI - Docker Guide

This guide explains how to use the containerized version of the CosmosLoadTester CLI tool using Docker and Docker Compose.

## 🚀 Quick Start

### Prerequisites

- Docker (version 20.10 or later)
- Docker Compose (version 2.0 or later)

### Basic Usage

1. **Build and run the container:**
   ```bash
   docker-compose up -d cosmosloadtester-cli
   ```

2. **Execute load test commands:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --help
   ```

3. **Run a basic load test:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
     --endpoints="ws://host.docker.internal:26657/websocket" \
     --duration=30s \
     --rate=100
   ```

## 📁 Directory Structure

After running the container, the following directories will be created:

```
./
├── config/           # Configuration profiles (mounted volume)
├── results/          # Test results and output files (mounted volume)
├── testnet-data/     # Local testnet data (if using local testnet)
└── monitoring/       # Monitoring configuration files
```

## 🛠️ Container Usage Patterns

### Interactive Mode

Start an interactive session:
```bash
docker-compose exec cosmosloadtester-cli /bin/sh
```

Run interactive load testing:
```bash
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --interactive
```

### One-off Commands

Run single commands without entering the container:
```bash
# List available client factories
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --list-factories

# Generate template profiles
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --generate-template=local-testnet

# Run load test with specific configuration
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --client-factory=test-cosmos-client-factory \
  --endpoints="ws://your-node:26657/websocket" \
  --duration=60s \
  --rate=1000 \
  --output-format=json \
  > ./results/test-results.json
```

### Using Configuration Profiles

1. **Create a profile inside the container:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
     --endpoints="ws://your-node:26657/websocket" \
     --duration=60s \
     --rate=500 \
     --save-profile=docker-test
   ```

2. **Use the saved profile:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --profile=docker-test
   ```

3. **List profiles:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --list-profiles
   ```

## 🌐 Network Configuration

### Connecting to External Nodes

When connecting to blockchain nodes running outside Docker:

- **On macOS/Windows:** Use `host.docker.internal` instead of `localhost`
  ```bash
  --endpoints="ws://host.docker.internal:26657/websocket"
  ```

- **On Linux:** Use the host's IP address or `--network=host` mode
  ```bash
  --endpoints="ws://192.168.1.100:26657/websocket"
  ```

### Using with Local Testnet

1. **Uncomment the cosmos-testnet service in docker-compose.yml**

2. **Start both services:**
   ```bash
   docker-compose up -d cosmos-testnet cosmosloadtester-cli
   ```

3. **Wait for testnet to be ready:**
   ```bash
   docker-compose logs -f cosmos-testnet
   ```

4. **Run load test against local testnet:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
     --endpoints="ws://cosmos-testnet:26657/websocket" \
     --duration=30s \
     --rate=100
   ```

## 📊 Output and Results

### File Output

Results are saved to the mounted `./results` directory:

```bash
# Save JSON results
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --profile=test \
  --output-format=json \
  --quiet > ./results/$(date +%Y%m%d_%H%M%S)_results.json

# Save CSV results
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --profile=test \
  --output-format=csv \
  --quiet > ./results/$(date +%Y%m%d_%H%M%S)_results.csv
```

### Real-time Monitoring

With the monitoring stack enabled:

```bash
# Start with monitoring
docker-compose --profile monitoring up -d

# Access Grafana dashboard
open http://localhost:3000  # admin/admin
```

## 🔧 Configuration Options

### Environment Variables

Set these in the docker-compose.yml or pass via command line:

```yaml
environment:
  - LOG_LEVEL=debug          # Log level: debug, info, warn, error
  - TZ=UTC                   # Timezone
  - NO_COLOR=false          # Disable colored output
```

### Volume Mounts

- `./config:/home/cosmosload/.cosmosloadtester` - Configuration profiles
- `./results:/app/results` - Output files and results
- `./examples:/app/examples:ro` - Example scripts (read-only)

## 🚀 Advanced Usage

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Load Test
on: [push]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start Cosmos LoadTester
        run: |
          docker-compose up -d cosmosloadtester-cli
          
      - name: Run Load Test
        run: |
          docker-compose exec -T cosmosloadtester-cli cosmosloadtester-cli \
            --endpoints="ws://testnet.cosmos.network:26657/websocket" \
            --duration=60s \
            --rate=500 \
            --output-format=json \
            --quiet > results.json
            
      - name: Process Results
        run: |
          TPS=$(cat results.json | jq '.avg_txs_per_second')
          echo "Achieved TPS: $TPS"
          if (( $(echo "$TPS < 400" | bc -l) )); then
            echo "Performance regression detected!"
            exit 1
          fi
```

### Multi-Stage Testing

Run different test scenarios:

```bash
#!/bin/bash
# multi-stage-test.sh

# Quick smoke test
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --benchmark=quick \
  --endpoints="ws://host.docker.internal:26657/websocket"

# Standard performance test
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --benchmark=standard \
  --endpoints="ws://host.docker.internal:26657/websocket"

# Stress test
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --benchmark=stress \
  --endpoints="ws://host.docker.internal:26657/websocket"
```

### Custom Client Factories

To use custom client factories:

1. **Add your client factory code to the project**
2. **Rebuild the container:**
   ```bash
   docker-compose build cosmosloadtester-cli
   ```
3. **Use the new factory:**
   ```bash
   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
     --client-factory=your-custom-factory \
     --endpoints="ws://your-node:26657/websocket"
   ```

## 🐛 Troubleshooting

### Common Issues

1. **Connection refused to localhost:**
   ```bash
   # Use host.docker.internal instead of localhost on macOS/Windows
   --endpoints="ws://host.docker.internal:26657/websocket"
   ```

2. **Permission denied on config directory:**
   ```bash
   # Fix permissions
   sudo chown -R $(id -u):$(id -g) ./config ./results
   ```

3. **Container doesn't start:**
   ```bash
   # Check logs
   docker-compose logs cosmosloadtester-cli
   
   # Rebuild container
   docker-compose build --no-cache cosmosloadtester-cli
   ```

4. **Out of memory errors:**
   ```bash
   # Increase Docker memory limits or reduce test parameters
   --rate=100 --connections=1
   ```

### Debug Mode

Enable debug logging:
```bash
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --log-level=debug \
  --profile=test
```

### Health Checks

Check container health:
```bash
docker-compose ps
docker inspect cosmosloadtester-cli | jq '.[0].State.Health'
```

## 📊 Monitoring Integration

### Prometheus Metrics

If using the monitoring profile:

```bash
# Start with monitoring
docker-compose --profile monitoring up -d

# Access Prometheus
open http://localhost:9090

# View Grafana dashboards
open http://localhost:3000  # admin/admin
```

### Custom Metrics Collection

Export metrics to external systems:

```bash
#!/bin/bash
# collect-metrics.sh

RESULTS=$(docker-compose exec -T cosmosloadtester-cli cosmosloadtester-cli \
  --profile=monitoring \
  --output-format=summary \
  --quiet)

# Extract metrics
TPS=$(echo "$RESULTS" | grep AVG_TPS | cut -d'=' -f2)
LATENCY=$(echo "$RESULTS" | grep LATENCY_P95 | cut -d'=' -f2)

# Send to monitoring system
curl -X POST http://monitoring-system/metrics \
  -d "service=cosmosloadtester&tps=$TPS&latency_p95=$LATENCY"
```

## 🔄 Scaling and Performance

### Multiple Containers

Run multiple load generators:

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  cosmosloadtester-cli-1:
    extends:
      service: cosmosloadtester-cli
    container_name: cosmosloadtester-cli-1
    
  cosmosloadtester-cli-2:
    extends:
      service: cosmosloadtester-cli
    container_name: cosmosloadtester-cli-2
```

### Resource Limits

Configure resource limits:

```yaml
services:
  cosmosloadtester-cli:
    # ... other config
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
        reservations:
          cpus: '1.0'
          memory: 512M
```

## 📚 Examples

### Basic Load Test

```bash
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --endpoints="ws://host.docker.internal:26657/websocket" \
  --duration=30s \
  --rate=100 \
  --connections=2
```

### High-Throughput Test

```bash
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --endpoints="ws://host.docker.internal:26657/websocket" \
  --duration=2m \
  --rate=5000 \
  --connections=10 \
  --broadcast-method=async \
  --size=40
```

### Latency Measurement

```bash
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --endpoints="http://host.docker.internal:26657" \
  --duration=1m \
  --rate=10 \
  --connections=1 \
  --broadcast-method=commit
```

## 🎯 Best Practices

1. **Use profiles for reproducible tests**
2. **Mount volumes for persistent configuration**
3. **Use appropriate resource limits**
4. **Monitor container health and logs**
5. **Clean up results regularly**
6. **Use JSON/CSV output for automation**
7. **Test with different client factories**
8. **Monitor blockchain node performance**

## 📄 License

Same license as the main cosmosloadtester project.

---

*Happy containerized load testing! 🚀🐳* 