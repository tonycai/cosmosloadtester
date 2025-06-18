# AIW3 Devnet Endpoints Reference

## 🌐 Available Endpoints

| Service | URL | Protocol | Purpose |
|---------|-----|----------|---------|
| **RPC Server** | `https://devnet-rpc.aiw3.io` | Tendermint RPC | Transaction submission, consensus queries |
| **API Server** | `https://devnet-api.aiw3.io` | REST API | Account queries, balance checks, chain data |
| **Faucet Server** | `https://devnet-faucet.aiw3.io` | HTTP API | Test token distribution |

## 🔧 Usage Examples

### Load Testing Scenarios

#### Single RPC Endpoint Testing
```bash
# High-throughput transaction testing
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=60s \
  --rate=1000 \
  --connections=5 \
  --broadcast-method=async

# Latency measurement
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=10 \
  --connections=1 \
  --broadcast-method=commit
```

#### Multi-Endpoint Load Balancing
```bash
# Load balancing across RPC and API endpoints
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=60s \
  --rate=500 \
  --connections=3 \
  --broadcast-method=sync

# Stress testing with multiple endpoints
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=120s \
  --rate=2000 \
  --connections=8 \
  --broadcast-method=async
```

### Profile-Based Testing
```bash
# Create AIW3 devnet profile (includes both endpoints)
./bin/cosmosloadtester-cli --generate-template=aiw3defi-test

# Use the profile for testing
./bin/cosmosloadtester-cli --profile=aiw3defi-test

# Override profile settings
./bin/cosmosloadtester-cli --profile=aiw3defi-test --rate=2000 --duration=30s
```

## 💰 Faucet Integration

### Request Test Tokens
```bash
# Basic faucet request
curl -X POST https://devnet-faucet.aiw3.io/request \
  -H 'Content-Type: application/json' \
  -d '{"address": "aiw3...", "amount": "1000000"}'

# Request with specific denomination
curl -X POST https://devnet-faucet.aiw3.io/request \
  -H 'Content-Type: application/json' \
  -d '{"address": "aiw3...", "amount": "1000000", "denom": "uaiw3"}'
```

### Check Account Balance
```bash
# Check balance via API server
curl https://devnet-api.aiw3.io/cosmos/bank/v1beta1/balances/aiw3...

# Check specific denomination balance
curl https://devnet-api.aiw3.io/cosmos/bank/v1beta1/balances/aiw3.../by_denom?denom=uaiw3
```

## 🐳 Docker Usage

### Single Command Testing
```bash
# Test single endpoint in Docker
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100

# Test multiple endpoints in Docker
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100
```

### Automated Setup in Docker
```bash
# Run the full AIW3 setup script in Docker
docker-compose exec cosmosloadtester-cli /app/examples/aiw3-devnet-example.sh

# Run specific steps
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --list-factories
docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --list-profiles
```

## 🔍 Endpoint Verification

### Manual Connectivity Testing
```bash
# Test RPC endpoint
curl -s --connect-timeout 10 https://devnet-rpc.aiw3.io

# Test API endpoint
curl -s --connect-timeout 10 https://devnet-api.aiw3.io

# Test faucet endpoint
curl -s --connect-timeout 10 https://devnet-faucet.aiw3.io
```

### CLI Connectivity Testing
```bash
# Test all endpoints
./bin/cosmosloadtester-cli --check-endpoints \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io"

# Test individual endpoints
./bin/cosmosloadtester-cli --check-endpoints --endpoints="https://devnet-rpc.aiw3.io"
./bin/cosmosloadtester-cli --check-endpoints --endpoints="https://devnet-api.aiw3.io"
```

## 📊 Monitoring and Metrics

### Output Formats
```bash
# JSON output for automation
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100 \
  --output-format=json > aiw3-results.json

# CSV output for analysis
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100 \
  --output-format=csv > aiw3-results.csv

# Summary for CI/CD
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100 \
  --output-format=summary
```

### Debug Logging
```bash
# Enable debug logging for troubleshooting
./bin/cosmosloadtester-cli \
  --endpoints="https://devnet-rpc.aiw3.io,https://devnet-api.aiw3.io" \
  --client-factory="aiw3defi-bank-send" \
  --duration=30s \
  --rate=100 \
  --log-level=debug
```

## 🚨 Troubleshooting

### Common Issues

1. **Endpoint Unreachable**
   ```bash
   # Check network connectivity
   ping devnet-rpc.aiw3.io
   ping devnet-api.aiw3.io
   
   # Test with curl
   curl -v https://devnet-rpc.aiw3.io
   curl -v https://devnet-api.aiw3.io
   ```

2. **Load Test Failures**
   ```bash
   # Start with minimal load
   ./bin/cosmosloadtester-cli \
     --endpoints="https://devnet-rpc.aiw3.io" \
     --rate=1 --duration=10s --dry-run
   
   # Check client factory
   ./bin/cosmosloadtester-cli --list-factories
   ```

3. **Authentication Issues**
   ```bash
   # Verify faucet access
   curl -X POST https://devnet-faucet.aiw3.io/request \
     -H 'Content-Type: application/json' \
     -d '{"address": "test", "amount": "1"}'
   ```

### Best Practices

1. **Start Small**: Begin with low rates and short durations
2. **Monitor Resources**: Watch for rate limiting and resource constraints
3. **Use Profiles**: Save configurations for repeatable testing
4. **Log Everything**: Enable debug logging for troubleshooting
5. **Test Connectivity**: Always verify endpoints before load testing

## 📚 Related Documentation

- [AIW3 Devnet Example Script](./aiw3-devnet-example.sh)
- [CLI Usage Examples](./cli-examples.sh)
- [Main README](../README.md#-aiw3-devnet-integration)
- [Integration Summary](../INTEGRATION_SUMMARY.md) 