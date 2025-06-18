#!/bin/bash

# AIW3 Devnet Load Testing Example
# This script demonstrates how to use cosmosloadtester with AIW3 devnet
# including faucet integration for getting test tokens

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AIW3_RPC_ENDPOINT="https://devnet-rpc.aiw3.io"
AIW3_API_ENDPOINT="https://devnet-api.aiw3.io"
AIW3_FAUCET_ENDPOINT="https://devnet-faucet.aiw3.io"
CLI_BINARY="./bin/cosmosloadtester-cli"

echo -e "${BLUE}=== AIW3 Devnet Load Testing Example ===${NC}"

# Check if CLI binary exists
if [ ! -f "$CLI_BINARY" ]; then
    echo -e "${RED}Error: CLI binary not found at $CLI_BINARY${NC}"
    echo -e "${YELLOW}Please build the CLI first: make cli${NC}"
    exit 1
fi

# Function to print step headers
print_step() {
    echo -e "\n${GREEN}=== $1 ===${NC}"
}

# Function to print info
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_step "Step 1: Check AIW3 Devnet Connectivity"
print_info "Testing connection to AIW3 devnet endpoints..."

# Test RPC endpoint connectivity
print_info "Testing RPC endpoint..."
if curl -s --connect-timeout 10 "$AIW3_RPC_ENDPOINT" > /dev/null; then
    print_success "AIW3 RPC endpoint is reachable"
else
    print_error "Cannot reach AIW3 RPC endpoint: $AIW3_RPC_ENDPOINT"
    print_warning "Please check your internet connection and try again"
    exit 1
fi

# Test API endpoint connectivity
print_info "Testing API endpoint..."
if curl -s --connect-timeout 10 "$AIW3_API_ENDPOINT" > /dev/null; then
    print_success "AIW3 API endpoint is reachable"
else
    print_warning "Cannot reach AIW3 API endpoint: $AIW3_API_ENDPOINT"
    print_warning "API endpoint may not be required for load testing"
fi

# Test faucet endpoint connectivity
print_info "Testing faucet endpoint..."
if curl -s --connect-timeout 10 "$AIW3_FAUCET_ENDPOINT" > /dev/null; then
    print_success "AIW3 faucet endpoint is reachable"
else
    print_warning "Cannot reach AIW3 faucet endpoint: $AIW3_FAUCET_ENDPOINT"
    print_warning "You may need to manually fund test accounts"
fi

print_step "Step 2: List Available Client Factories"
print_info "Checking available client factories..."
$CLI_BINARY --list-factories

print_step "Step 3: Create AIW3 Devnet Profile"
print_info "Creating a load test profile for AIW3 devnet..."

# Generate AIW3 devnet profile
$CLI_BINARY --generate-template=aiw3defi-test --quiet > /dev/null 2>&1 || true

# Create custom AIW3 devnet profile
cat > /tmp/aiw3-devnet-profile.yaml << EOF
name: aiw3-devnet
description: AIW3 Devnet load testing configuration with faucet integration
client_factory: aiw3defi-bank-send
connections: 3
duration: 60s
send_period: 1s
transactions_per_second: 100
transaction_size: 512
transaction_count: -1
broadcast_method: sync
endpoints:
  - $AIW3_RPC_ENDPOINT
  - $AIW3_API_ENDPOINT
endpoint_select_method: supplied
expect_peers: 0
max_endpoints: 0
min_connectivity: 0
peer_connect_timeout: 5s
tags:
  - aiw3
  - devnet
  - bank-send
  - multi-endpoint
created_at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
updated_at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

# Import the profile
if $CLI_BINARY --import-profiles=/tmp/aiw3-devnet-profile.yaml --import-format=yaml --quiet; then
    print_success "AIW3 devnet profile created successfully"
else
    print_error "Failed to create AIW3 devnet profile"
    exit 1
fi

# Clean up temp file
rm -f /tmp/aiw3-devnet-profile.yaml

print_step "Step 4: Display Profile Details"
$CLI_BINARY --show-profile=aiw3-devnet

print_step "Step 5: Faucet Integration Guide"
print_info "AIW3 Devnet endpoints and services:"
echo -e "${BLUE}RPC Endpoint: $AIW3_RPC_ENDPOINT${NC}"
echo -e "${BLUE}API Endpoint: $AIW3_API_ENDPOINT${NC}"
echo -e "${BLUE}Faucet URL: $AIW3_FAUCET_ENDPOINT${NC}"
echo ""
print_info "To get test tokens for AIW3 devnet, you can use the faucet:"
echo -e "${YELLOW}curl -X POST $AIW3_FAUCET_ENDPOINT/request \\"
echo -e "  -H 'Content-Type: application/json' \\"
echo -e "  -d '{\"address\": \"aiw3...\", \"amount\": \"1000000\"}'"
echo -e "${NC}"
echo ""
print_info "You can also check account balances via the API:"
echo -e "${YELLOW}curl $AIW3_API_ENDPOINT/cosmos/bank/v1beta1/balances/aiw3...${NC}"

print_step "Step 6: Validate Configuration"
print_info "Validating the AIW3 devnet configuration..."
if $CLI_BINARY --profile=aiw3-devnet --validate-config; then
    print_success "Configuration is valid"
else
    print_error "Configuration validation failed"
    exit 1
fi

print_step "Step 7: Dry Run Test"
print_info "Performing a dry run to preview the test configuration..."
$CLI_BINARY --profile=aiw3-devnet --dry-run

print_step "Step 8: Run Load Test Options"
echo -e "${BLUE}You can now run load tests using any of these methods:${NC}"
echo ""
echo -e "${YELLOW}1. Using the profile (multi-endpoint):${NC}"
echo -e "   $CLI_BINARY --profile=aiw3-devnet"
echo ""
echo -e "${YELLOW}2. Quick benchmark (single endpoint):${NC}"
echo -e "   $CLI_BINARY --benchmark=quick --endpoints=$AIW3_RPC_ENDPOINT"
echo ""
echo -e "${YELLOW}3. Multi-endpoint testing:${NC}"
echo -e "   $CLI_BINARY \\"
echo -e "     --endpoints=\"$AIW3_RPC_ENDPOINT,$AIW3_API_ENDPOINT\" \\"
echo -e "     --client-factory=aiw3defi-bank-send \\"
echo -e "     --duration=30s \\"
echo -e "     --rate=50 \\"
echo -e "     --connections=2"
echo ""
echo -e "${YELLOW}4. Single RPC endpoint testing:${NC}"
echo -e "   $CLI_BINARY \\"
echo -e "     --endpoints=$AIW3_RPC_ENDPOINT \\"
echo -e "     --client-factory=aiw3defi-bank-send \\"
echo -e "     --duration=30s \\"
echo -e "     --rate=100 \\"
echo -e "     --connections=3"
echo ""
echo -e "${YELLOW}5. Interactive mode:${NC}"
echo -e "   $CLI_BINARY --interactive"

print_step "Step 9: Advanced Testing Scenarios"
echo -e "${BLUE}Advanced testing scenarios for AIW3 devnet:${NC}"
echo ""
echo -e "${YELLOW}High-throughput test:${NC}"
echo -e "   $CLI_BINARY --profile=aiw3-devnet --rate=1000 --connections=5 --duration=120s"
echo ""
echo -e "${YELLOW}Latency measurement:${NC}"
echo -e "   $CLI_BINARY --profile=aiw3-devnet --rate=10 --connections=1 --broadcast-method=commit"
echo ""
echo -e "${YELLOW}Stress test:${NC}"
echo -e "   $CLI_BINARY --benchmark=stress --endpoints=$AIW3_RPC_ENDPOINT --client-factory=aiw3defi-bank-send"

print_step "Step 10: Monitoring and Results"
print_info "Output formats available:"
echo -e "${BLUE}• Live output (default): Real-time progress and results${NC}"
echo -e "${BLUE}• JSON format: Machine-readable results for automation${NC}"
echo -e "${BLUE}• CSV format: Spreadsheet-compatible data${NC}"
echo -e "${BLUE}• Summary format: Key metrics for CI/CD${NC}"
echo ""
print_info "Example with JSON output:"
echo -e "${YELLOW}$CLI_BINARY --profile=aiw3-devnet --output-format=json > results.json${NC}"

print_step "Step 11: Troubleshooting"
echo -e "${BLUE}Common troubleshooting steps:${NC}"
echo ""
echo -e "${YELLOW}1. Check endpoint connectivity:${NC}"
echo -e "   $CLI_BINARY --check-endpoints --endpoints=\"$AIW3_RPC_ENDPOINT,$AIW3_API_ENDPOINT\""
echo ""
echo -e "${YELLOW}2. Test individual endpoints:${NC}"
echo -e "   $CLI_BINARY --check-endpoints --endpoints=$AIW3_RPC_ENDPOINT"
echo -e "   $CLI_BINARY --check-endpoints --endpoints=$AIW3_API_ENDPOINT"
echo ""
echo -e "${YELLOW}3. Enable debug logging:${NC}"
echo -e "   $CLI_BINARY --profile=aiw3-devnet --log-level=debug"
echo ""
echo -e "${YELLOW}4. Test with minimal load:${NC}"
echo -e "   $CLI_BINARY --endpoints=$AIW3_RPC_ENDPOINT --rate=1 --duration=10s"
echo ""
echo -e "${YELLOW}5. Verify client factory:${NC}"
echo -e "   $CLI_BINARY --list-factories"

print_step "Docker Usage"
echo -e "${BLUE}To run in Docker container:${NC}"
echo ""
echo -e "${YELLOW}1. Build and start container:${NC}"
echo -e "   docker-compose up -d cosmosloadtester-cli"
echo ""
echo -e "${YELLOW}2. Run multi-endpoint load test in container:${NC}"
echo -e "   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \\"
echo -e "     --endpoints=\"$AIW3_RPC_ENDPOINT,$AIW3_API_ENDPOINT\" \\"
echo -e "     --client-factory=aiw3defi-bank-send \\"
echo -e "     --duration=30s \\"
echo -e "     --rate=100"
echo ""
echo -e "${YELLOW}3. Run single endpoint test in container:${NC}"
echo -e "   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli \\"
echo -e "     --endpoints=$AIW3_RPC_ENDPOINT \\"
echo -e "     --client-factory=aiw3defi-bank-send \\"
echo -e "     --duration=30s \\"
echo -e "     --rate=100"
echo ""
echo -e "${YELLOW}4. Use persistent profiles:${NC}"
echo -e "   docker-compose exec cosmosloadtester-cli cosmosloadtester-cli --profile=aiw3-devnet"

print_success "AIW3 devnet setup complete!"
print_info "The 'aiw3-devnet' profile is now available for load testing"
print_info "Run '$CLI_BINARY --profile=aiw3-devnet' to start testing"

echo -e "\n${GREEN}🚀 Happy load testing with AIW3 devnet! 🚀${NC}" 