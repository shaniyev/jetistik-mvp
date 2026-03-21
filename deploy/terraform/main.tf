# ── Image ──
data "openstack_images_image_v2" "almalinux" {
  name        = var.image_name
  most_recent = true
}

# ── Security group ──
resource "openstack_networking_secgroup_v2" "jetistik" {
  name        = "jetistik"
  description = "Jetistik MVP security group"
}

resource "openstack_networking_secgroup_rule_v2" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.jetistik.id
}

resource "openstack_networking_secgroup_rule_v2" "http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.jetistik.id
}

resource "openstack_networking_secgroup_rule_v2" "https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.jetistik.id
}

resource "openstack_networking_secgroup_rule_v2" "icmp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.jetistik.id
}

# ── Boot volume ──
resource "openstack_blockstorage_volume_v3" "jetistik_boot" {
  name        = "jetistik-boot"
  size        = var.boot_volume_size
  volume_type = var.volume_type
  image_id    = data.openstack_images_image_v2.almalinux.id
}

# ── Compute instance ──
resource "openstack_compute_instance_v2" "jetistik" {
  name        = "jetistik"
  flavor_name = var.flavor_name
  key_pair    = var.key_pair_name

  security_groups = [
    openstack_networking_secgroup_v2.jetistik.name,
  ]

  block_device {
    uuid                  = openstack_blockstorage_volume_v3.jetistik_boot.id
    source_type           = "volume"
    destination_type      = "volume"
    boot_index            = 0
    delete_on_termination = true
  }

  network {
    name = var.network_name
  }

  metadata = {
    role = "jetistik"
  }
}

# ── Floating IP ──
resource "openstack_networking_floatingip_v2" "jetistik_fip" {
  pool = var.floating_ip_pool
}

resource "openstack_compute_floatingip_associate_v2" "jetistik_fip_assoc" {
  floating_ip = openstack_networking_floatingip_v2.jetistik_fip.address
  instance_id = openstack_compute_instance_v2.jetistik.id
}
