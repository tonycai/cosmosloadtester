#!/bin/bash
# CLI Examples for Cosmos Load Tester
# These examples demonstrate various usage patterns

set -e

CLI_BINARY="./bin/cosmosloadtester-cli"
TESTNET_ENDPOINT="ws://localhost:26657/websocket"

echo "🚀 Cosmos Load Tester CLI Examples"
echo "=================================="

# Ensure CLI is built
if [ ! -f "$CLI_BINARY" ]; then
    echo "Building CLI tool..."
    make cli
fi

echo ""
echo "1. 📋 List Available Client Factories"
echo "------------------------------------"
$CLI_BINARY --list-factories

echo ""
echo "2. 🎯 Basic Load Test (30 seconds, 100 TPS)"
echo "-------------------------------------------"
read -p "Press Enter to run basic load test (or Ctrl+C to skip)..."
$CLI_BINARY \
    --endpoints="$TESTNET_ENDPOINT" \
    --duration=30s \
    --rate=100 \
    --connections=2 \
    --quiet

echo ""
echo "3. 💾 Create and Use Configuration Profile"
echo "------------------------------------------"
echo "Creating 'example-profile'..."
$CLI_BINARY \
    --endpoints="$TESTNET_ENDPOINT" \
    --duration=20s \
    --rate=50 \
    --connections=1 \
    --broadcast-method=sync \
    --save-profile=example-profile

echo "Using saved profile..."
$CLI_BINARY --profile=example-profile --quiet

echo ""
echo "4. 📊 JSON Output Example"
echo "-------------------------"
echo "Running test with JSON output..."
$CLI_BINARY \
    --endpoints="$TESTNET_ENDPOINT" \
    --duration=10s \
    --rate=50 \
    --output-format=json \
    --quiet > example-results.json

echo "JSON results saved to example-results.json"
echo "Sample output:"
cat example-results.json | jq '{total_txs, avg_txs_per_second, total_time}' 2>/dev/null || cat example-results.json | head -10

echo ""
echo "5. 🏭 Generate Template Profiles"
echo "--------------------------------"
echo "Generating local-testnet template..."
$CLI_BINARY --generate-template=local-testnet

echo ""
echo "6. 📈 Validation and Dry Run"
echo "----------------------------"
echo "Validating configuration..."
$CLI_BINARY --validate --profile=example-profile

echo "Dry run preview..."
$CLI_BINARY --dry-run --profile=example-profile

echo ""
echo "7. 🔍 Endpoint Health Check"
echo "---------------------------"
echo "Checking endpoint connectivity..."
$CLI_BINARY --check-endpoints --endpoints="$TESTNET_ENDPOINT"

echo ""
echo "8. 📋 Profile Management"
echo "------------------------"
echo "Listing all profiles..."
$CLI_BINARY --list-profiles

echo "Showing profile details..."
$CLI_BINARY --show-profile=example-profile

echo ""
echo "9. 📊 CSV Export Example"  
echo "------------------------"
echo "Running test with CSV output..."
$CLI_BINARY \
    --profile=example-profile \
    --output-format=csv \
    --quiet > example-results.csv

echo "CSV results saved to example-results.csv"
echo "Sample CSV output:"
head -5 example-results.csv

echo ""
echo "10. 🧪 AIW3 DeFi Testing (if available)"
echo "---------------------------------------"
if $CLI_BINARY --list-factories | grep -q "aiw3defi-bank-send"; then
    echo "Running AIW3 DeFi test..."
    $CLI_BINARY \
        --client-factory=aiw3defi-bank-send \
        --endpoints="$TESTNET_ENDPOINT" \
        --duration=15s \
        --rate=100 \
        --size=512 \
        --quiet
else
    echo "AIW3 DeFi client factory not available"
fi

echo ""
echo "11. 📦 Profile Import/Export"
echo "----------------------------"
echo "Exporting profiles..."
$CLI_BINARY --export-profiles=exported-profiles.yaml --export-format=yaml

echo "Sample exported profiles:"
head -20 exported-profiles.yaml

echo ""
echo "12. 🎯 Multi-Endpoint Testing"
echo "-----------------------------"
echo "Testing with multiple endpoints (if available)..."
MULTI_ENDPOINTS="$TESTNET_ENDPOINT,http://localhost:26657"
$CLI_BINARY \
    --endpoints="$MULTI_ENDPOINTS" \
    --endpoint-select-method=any \
    --duration=15s \
    --rate=100 \
    --connections=1 \
    --quiet || echo "Multi-endpoint test skipped (endpoints not available)"

echo ""
echo "13. 🏃‍♂️ Quick Benchmark Suite"
echo "------------------------------"
read -p "Run quick benchmark? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    $CLI_BINARY --benchmark=quick --endpoints="$TESTNET_ENDPOINT"
fi

echo ""
echo "14. 🔧 Interactive Mode Demo"
echo "----------------------------"
echo "To try interactive mode, run:"
echo "  $CLI_BINARY --interactive"

echo ""
echo "15. 📊 Summary Output for Scripting"
echo "-----------------------------------"
echo "Getting summary metrics..."
SUMMARY_OUTPUT=$($CLI_BINARY \
    --profile=example-profile \
    --output-format=summary \
    --quiet)

echo "Extracting specific metrics:"
echo "$SUMMARY_OUTPUT" | grep -E "(TOTAL_TXS|AVG_TPS|LATENCY_P95)"

# Extract TPS value for conditional logic
TOTAL_TPS=$(echo "$SUMMARY_OUTPUT" | grep AVG_TPS | cut -d'=' -f2)
echo "Achieved TPS: $TOTAL_TPS"

if (( $(echo "$TOTAL_TPS > 50" | bc -l 2>/dev/null || echo "1") )); then
    echo "✅ Performance target met!"
else
    echo "⚠️  Performance below target"
fi

echo ""
echo "🧹 Cleanup"
echo "----------"
echo "Cleaning up example files..."
rm -f example-results.json example-results.csv exported-profiles.yaml

echo "Deleting example profile..."
$CLI_BINARY --delete-profile=example-profile || echo "Profile already deleted"

echo ""
echo "✅ CLI Examples Complete!"
echo "========================"
echo ""
echo "📚 For more information:"
echo "  • Read CLI_README.md for comprehensive documentation"
echo "  • Run '$CLI_BINARY --help' for all available options"
echo "  • Try '$CLI_BINARY --interactive' for guided setup"
echo ""
echo "🎯 Common Next Steps:"
echo "  1. Create profiles for your specific test scenarios"
echo "  2. Integrate with CI/CD pipeline using JSON/CSV output"
echo "  3. Set up monitoring scripts using summary output"
echo "  4. Explore benchmark suites for performance testing"
echo ""
echo "Happy load testing! 🚀" 