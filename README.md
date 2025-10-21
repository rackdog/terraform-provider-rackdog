# Terraform Provider for Rackdog

The **Rackdog Terraform Provider** allows you to provision and manage infrastructure resources on the Rackdog platform using declarative Terraform configuration.

Rackdog combines high-performance bare-metal automation with a modern API and Terraform integration — making it simple to deploy and manage servers at global scale.

---

## Example

Starting with Terraform 0.13+, providers are automatically installed from the Terraform Registry.

```hcl
terraform {
  required_providers {
    rackdog = { source = "rackdog/rackdog", version = "0.0.5" }
  }
}
provider "rackdog" {
  recreate_on_missing = true
}

data "rackdog_operating_systems" "all" {}

data "rackdog_plans" "ny" { location = "ny" }

locals {
  chosen_plan = one([for p in data.rackdog_plans.ny.plans : p if p.name == "test"])
  chosen_os   = data.rackdog_operating_systems.all.operating_systems[0]
}

resource "rackdog_server" "web" {
  plan_id     = local.chosen_plan.id
  location_id = 1
  os_id       = local.chosen_os.id                    
  hostname    = "web-01"
}


