PLUGIN_ADDR=registry.terraform.io/rackdog/rackdog
VERSION?=0.0.1
OS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
INSTALL_DIR=$(HOME)/.local/share/terraform/plugins/$(PLUGIN_ADDR)/$(VERSION)/$(OS)_$(ARCH)
RACKDOG_API_KEY=$(cat .env | grep "KEY" | awk -F'=' '{print $2}')
RACKDOG_ENDPOINT=$(cat .env | grep "RACKDOG_ENDPOINT" | awk -F'=' '{print $2}')

build:
	go build -o bin/terraform-provider-rackdog .

install: build
	mkdir -p $(INSTALL_DIR)
	cp bin/terraform-provider-rackdog $(INSTALL_DIR)/

example-init:
	cd examples/basic && terraform init

example-plan:
	set -a; source .env; set +a && cd examples/basic && terraform plan

example-apply:
	set -a; source .env; set +a && cd examples/basic && terraform apply -auto-approve

example-destroy:
	set -a; source .env; set +a && cd examples/basic && terraform destroy -auto-approve

fmt:
	go fmt ./...

# Run unit tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

# Run full test suite with formatting, linting, and coverage
test-all:
	./run-tests.sh

# Run acceptance tests (requires real API credentials)
acc:
	TF_ACC=1 RACKDOG_API_KEY=$(RACKDOG_API_KEY) RACKDOG_ENDPOINT=$(RACKDOG_ENDPOINT) go test ./... -v -timeout=30m

# View coverage report in browser
coverage-html:
	go tool cover -html=coverage.out

# Run tests in watch mode (requires entr)
test-watch:
	find . -name '*.go' | entr -c go test ./... -v

# Clean test artifacts
clean-test:
	rm -f coverage.out coverage.html

.PHONY: test test-coverage test-all test-watch coverage-html clean-test

