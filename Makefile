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

example-dynamic-plan:
	set -a; source .env; set +a && cd examples/dynamic && terraform plan

example-dynamic-apply:
	set -a; source .env; set +a && cd examples/dynamic && terraform apply -auto-approve

example-dynamic-destroy:
	set -a; source .env; set +a && cd examples/dynamic && terraform destroy -auto-approve

fmt:
	go fmt ./...

test:
	go test ./... -v

test-coverage:
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

test-all:
	./run-tests.sh

coverage-html:
	go tool cover -html=coverage.out

test-watch:
	find . -name '*.go' | entr -c go test ./... -v

clean-test:
	rm -f coverage.out coverage.html

.PHONY: test test-coverage test-all test-watch coverage-html clean-test

