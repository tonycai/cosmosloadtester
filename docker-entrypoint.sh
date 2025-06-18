#!/bin/sh
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print banner
echo -e "${BLUE}"
echo "   ____                              _                    _ _____         _            "
echo "  / ___|___  ___ _ __ ___   ___  ___| |    ___   __ _  __| |_   _|__  ___| |_ ___ _ __ "
echo " | |   / _ \/ __| '_ \` _ \ / _ \/ __| |   / _ \ / _\` |/ _\` | | |/ _ \/ __| __/ _ \ '__|"
echo " | |__| (_) \__ \ | | | | | (_) \__ \ |__| (_) | (_| | (_| | | |  __/\__ \ ||  __/ |   "
echo "  \____\___/|___/_| |_| |_|\___/|___/____|\___/ \__,_|\__,_| |_|\___||___/\__\___|_|   "
echo "                                                                                       "
echo "                       Terminal-based Cosmos Load Testing Tool"
echo "                                   Docker Edition"
echo -e "${NC}"

# Create necessary directories
mkdir -p /home/cosmosload/.cosmosloadtester
mkdir -p /app/results

# Check if cosmosloadtester-cli is available
if ! command -v cosmosloadtester-cli >/dev/null 2>&1; then
    echo -e "${RED}Error: cosmosloadtester-cli not found in PATH${NC}"
    exit 1
fi

# Display help information
echo -e "${GREEN}Container is ready!${NC}"
echo -e "${YELLOW}Available commands:${NC}"
echo "  cosmosloadtester-cli --help                    # Show help"
echo "  cosmosloadtester-cli --list-factories          # List client factories"
echo "  cosmosloadtester-cli --interactive             # Interactive mode"
echo "  cosmosloadtester-cli --generate-template=TYPE  # Generate templates"
echo ""
echo -e "${YELLOW}Quick examples:${NC}"
echo "  # Basic load test"
echo "  cosmosloadtester-cli --endpoints=\"ws://host.docker.internal:26657/websocket\" --duration=30s --rate=100"
echo ""
echo "  # Use a profile"
echo "  cosmosloadtester-cli --profile=my-profile"
echo ""
echo "  # Interactive mode"
echo "  cosmosloadtester-cli --interactive"
echo ""
echo -e "${YELLOW}Volume mounts:${NC}"
echo "  /home/cosmosload/.cosmosloadtester  # Configuration profiles"
echo "  /app/results                        # Output files"
echo ""

# If no arguments provided, start interactive shell
if [ $# -eq 0 ]; then
    echo -e "${GREEN}Starting interactive shell...${NC}"
    echo "Type 'cosmosloadtester-cli --help' to get started"
    exec /bin/sh
fi

# If first argument is a flag or cosmosloadtester-cli command, run it
if [ "${1#-}" != "$1" ] || [ "$1" = "cosmosloadtester-cli" ]; then
    exec cosmosloadtester-cli "$@"
fi

# Otherwise, execute the command as-is
exec "$@" 