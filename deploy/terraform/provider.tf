terraform {
  required_version = ">= 1.0"

  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.53"
    }
  }
}

provider "openstack" {
  auth_url    = var.openstack_auth_url
  user_name   = var.openstack_user_name
  password    = var.openstack_password
  tenant_name = var.openstack_tenant_name
  tenant_id   = var.openstack_tenant_id

  user_domain_name = var.openstack_user_domain_name
  project_domain_id = var.openstack_project_domain_id

  region = var.openstack_region
}
