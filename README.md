## First step
go build -o bin/terraform-provider-rackdog .


## Setup env
VER=0.0.1
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
PLUGDIR="$HOME/.local/share/terraform/plugins/registry.terraform.io/rackdog/rackdog/$VER/${OS}_${ARCH}"
mkdir -p "$PLUGDIR"
cp bin/terraform-provider-rackdog "$PLUGDIR/"

## could also just use makefile

