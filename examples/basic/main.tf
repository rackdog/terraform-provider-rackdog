terraform {
  required_providers {
    rackdog = {
      source  = "rackdog/rackdog"
      #version = "0.0.1"
    }
  }
}

provider "rackdog" {
  endpoint = var.endpoint
  api_key  = var.api_key
}

variable "endpoint" { 
    type = string 
}
variable "api_key"  { 
    type = string
    sensitive = true 
}

resource "rackdog_server" "web" {
  plan_id     = 101   # <- int
  location_id = 3     # <- int
  os_id       = 62    # <- int
  hostname    = "web-01"
}

output "server_id"      { value = rackdog_server.web.id }
output "server_ip"      { value = rackdog_server.web.ip_address }
output "server_status"  { value = rackdog_server.web.status }

