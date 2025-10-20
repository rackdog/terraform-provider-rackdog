terraform {
  required_providers {
    rackdog = {
      source  = "rackdog/rackdog"
      #version = "0.0.1"
    }
  }
}

provider "rackdog" {
  recreate_on_missing    = true
}

resource "rackdog_server" "web" {
  plan_id     = 8
  location_id = 1
  os_id       = 48
  hostname    = "web-01"
}

resource "rackdog_server" "web-2" {
  plan_id     = 8
  location_id = 1
  os_id       = 48
  hostname    = "web-02"
}

output "web-01_id"      { value = rackdog_server.web.id }
output "web-01_ip"      { value = rackdog_server.web.ip_address }
output "web-02_id"      { value = rackdog_server.web-2.id }
output "web-02_ip"      { value = rackdog_server.web-2.ip_address }

