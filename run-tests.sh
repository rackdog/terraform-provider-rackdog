#!/bin/bash

# Rackdog Terraform Provider - Test Suite Runner
# This script runs all tests and generates a coverage report

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "==========================================  "
echo " Rackdog Terraform Provider - Test Suite"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Go version:${NC}"
go version
echo ""

# Run formatting check
echo -e "${YELLOW}Checking code formatting...${NC}"
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo -e "${RED}The following files are not formatted:${NC}"
    echo "$UNFORMATTED"
    echo ""
    echo "Run 'make fmt' to format them"
    exit 1
fi
echo -e "${GREEN}✓ All files are properly formatted${NC}"
echo ""

# Run linting if staticcheck is available
if command -v staticcheck &> /dev/null; then
    echo -e "${YELLOW}Running static analysis...${NC}"
    staticcheck ./...
    echo -e "${GREEN}✓ Static analysis passed${NC}"
    echo ""
else
    echo -e "${YELLOW}staticcheck not found, skipping static analysis${NC}"
    echo "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
    echo ""
fi

# Run unit tests
echo -e "${YELLOW}Running unit tests...${NC}"
go test ./... -v -race -coverprofile=coverage.out -covermode=atomic

TEST_EXIT=$?

if [ $TEST_EXIT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""

    # Generate coverage report
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -func=coverage.out | tail -1

    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
    echo ""

    # Show coverage summary
    echo -e "${YELLOW}Coverage by package:${NC}"
    go tool cover -func=coverage.out | grep -v "total:" | column -t
    echo ""

    exit 0
else
    echo ""
    echo -e "${RED}✗ Tests failed${NC}"
    echo ""
    exit $TEST_EXIT
fi
