output "jetistik_instance_id" {
  value = openstack_compute_instance_v2.jetistik.id
}

output "jetistik_private_ip" {
  value = openstack_compute_instance_v2.jetistik.access_ip_v4
}

output "jetistik_floating_ip" {
  value = openstack_networking_floatingip_v2.jetistik_fip.address
}

output "ssh_command" {
  value = "ssh almalinux@${openstack_networking_floatingip_v2.jetistik_fip.address}"
}
