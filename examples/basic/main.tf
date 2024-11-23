terraform {
  required_providers {
    udm = {
      source = "registry.terraform.io/josh-hogle/ubiquiti-udm"
    }
  }
}

provider "udm" {
  hostname = var.udm_hostname
  username = var.udm_admin_username
  password = var.udm_admin_password
  ignore_untrusted_ssl_certificate = var.udm_uses_untrusted_ssl_cert
}

data "udm_static_dns_records" "all" {}

data "udm_static_dns_records" "filtered" {
  filter = {
    enabled = "true"
    record_type = "NS"
    key = "internal.acmelabs.dev"
  }
}

data "udm_client_devices" "all" {}

data "udm_client_devices" "filtered" {
  filter = {
    use_fixed_ip = true
    local_dns_record_enabled = true
  }
}

#resource "udm_static_dns_record" "k8s_dev_cluster" {
#  enabled = true
#  key = "k8s-dev-cluster.internal.acmelabs.dev"
#  record_type = "A"
#  value = "192.168.20.10"
#  ttl = 300
#}

#resource "udm_client_device" "test_device" {
#  mac_address = "12:12:12:12:12:21"
#  use_fixed_ip = true
#  fixed_ip = "192.168.16.221"
#}

output "static_dns_records" {
  value = data.udm_static_dns_records.all
}

output "filtered_static_dns_records" {
  value = data.udm_static_dns_records.filtered
}


output "filtered_client_devices" {
  value = data.udm_client_devices.filtered
}