terraform {
  required_providers {
    rackdog = {
      source  = "rackdog/rackdog"
      #version = "0.0.1"
    }
  }
}

provider "rackdog" {}

resource "rackdog_server" "web" {
  plan_id     = 8
  location_id = 1
  os_id       = 48
  hostname    = "web-01"
}

output "server_id"      { value = rackdog_server.web.id }
output "server_ip"      { value = rackdog_server.web.ip_address }

