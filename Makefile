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

test:
	go test ./... -v

acc:
	TF_ACC=1 RACKDOG_API_KEY=$(RACKDOG_API_KEY) RACKDOG_ENDPOINT=$(RACKDOG_ENDPOINT) go test ./... -v -timeout=30m

