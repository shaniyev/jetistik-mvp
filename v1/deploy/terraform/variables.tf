# ── OpenStack credentials ──
variable "openstack_auth_url" {
  type      = string
  sensitive = true
}

variable "openstack_user_name" {
  type      = string
  sensitive = true
}

variable "openstack_password" {
  type      = string
  sensitive = true
}

variable "openstack_tenant_name" {
  type      = string
  sensitive = true
}

variable "openstack_tenant_id" {
  type      = string
  sensitive = true
}

variable "openstack_user_domain_name" {
  type    = string
  default = "Default"
}

variable "openstack_project_domain_id" {
  type    = string
  default = "default"
}

variable "openstack_region" {
  type    = string
  default = "kz-ast-1"
}

# ── Infrastructure ──
variable "image_name" {
  type    = string
  default = "AlmaLinux-10-x86_64-202510"
}

variable "flavor_name" {
  type    = string
  default = "d1.ram8cpu2"
}

variable "boot_volume_size" {
  type    = number
  default = 20
}

variable "volume_type" {
  type    = string
  default = "ceph-ssd"
}

variable "key_pair_name" {
  type    = string
  default = "rokko"
}

variable "network_name" {
  type = string
}

variable "floating_ip_pool" {
  type    = string
  default = "FloatingIP Net"
}
