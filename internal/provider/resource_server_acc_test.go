package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Server_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	endpoint := os.Getenv("RACKDOG_ENDPOINT")
	apiKey := os.Getenv("RACKDOG_API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("RACKDOG_ENDPOINT/RACKDOG_API_KEY not set")
	}

	// Minimal config: adjust IDs that exist in your staging env
	cfg := `
provider "rackdog" {
  endpoint = "` + endpoint + `"
  api_key  = "` + apiKey + `"
}

resource "rackdog_server" "test" {
  plan_id     = 101
  location_id = 3
  os_id       = 62
  # hostname optional
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"rackdog": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{Config: cfg, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("rackdog_server.test", "id"),
			)},
			{ResourceName: "rackdog_server.test", ImportState: true, ImportStateVerify: true},
		},
	})
}
