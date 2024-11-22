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

data "udm_static_dns_entries" "all" {}

data "udm_static_dns_entries" "filtered" {
  filter = {
    enabled = "true"
    record_type = "NS"
    key = "internal.acmelabs.dev"
  }
}

#resource "udm_static_dns" "k8s_dev_cluster" {
#  enabled = true
#  key = "k8s-dev-cluster.internal.acmelabs.dev"
#  record_type = "A"
#  value = "192.168.20.10"
#  ttl = 300
#}

output "static_dns_entries" {
  value = data.udm_static_dns_entries.all
}

output "static_dns_filtered_entries" {
  value = data.udm_static_dns_entries.filtered
}