
package main

import (
	"context"
	"flag"
	"log"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackdog/terraform-provider-rackdog/internal/provider"
) 

var version = "dev"

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	ctx := context.Background()
	if *debug {
		tflog.SetLogger(log.Default())
	}

	providerserver.Serve(ctx, provider.New(version), providerserver.ServeOpts {
		Address: "registry.terraform.io/rackdog/rackdog"
	})
	
}
