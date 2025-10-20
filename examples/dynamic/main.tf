terraform {
  required_providers {
    rackdog = { source = "rackdog/rackdog", version = "0.0.1" }
  }
}
provider "rackdog" {}

data "rackdog_operating_systems" "all" {}

data "rackdog_plans" "ny" { location = "ny" }

locals {
  chosen_plan = one([for p in data.rackdog_plans.ny.plans : p if p.name == "test"])
  chosen_os   = data.rackdog_operating_systems.all.operating_systems[0]
}

resource "rackdog_server" "web" {
  plan_id     = tonumber(local.chosen_plan.id)         
  location_id = 3
  os_id       = local.chosen_os.id                    
  hostname    = "web-01"
  # raid     = 1
}

output "server" { value = rackdog_server.web }

