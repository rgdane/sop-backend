#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== JK-API Unit Tests ===${NC}\n"

# Check if running specific layer
if [ "$1" == "services" ]; then
    echo -e "${YELLOW}Running Service Layer Tests...${NC}"
    go test ./test/services -v
elif [ "$1" == "handlers" ]; then
    echo -e "${YELLOW}Running Handler Layer Tests...${NC}"
    go test ./test/handlers -v
elif [ "$1" == "controllers" ]; then
    echo -e "${YELLOW}Running Controller Layer Tests...${NC}"
    go test ./test/controllers -v
elif [ "$1" == "coverage" ]; then
    echo -e "${YELLOW}Running All Tests with Coverage...${NC}"
    go test ./test/... -coverprofile=coverage.out
    echo -e "\n${GREEN}Coverage Report:${NC}"
    go tool cover -func=coverage.out
    echo -e "\n${YELLOW}Opening HTML coverage report...${NC}"
    go tool cover -html=coverage.out
else
    echo -e "${YELLOW}Running All Tests...${NC}\n"
    go test ./test/... -v
    
    if [ $? -eq 0 ]; then
        echo -e "\n${GREEN} All tests passed!${NC}"
    else
        echo -e "\n${RED} Some tests failed!${NC}"
        exit 1
    fi
fi

echo -e "\n${YELLOW}Usage:${NC}"
echo "  ./run-tests.sh          - Run all tests"
echo "  ./run-tests.sh services - Run service layer tests only"
echo "  ./run-tests.sh handlers - Run handler layer tests only"
echo "  ./run-tests.sh coverage - Run tests with coverage report"
