variable "udm_admin_password" {
    type = string
    description = "Password for a local account with administrator privileges on the UDM device"
}

variable "udm_admin_username" {
    type = string
    description = "Username for a local account with administrator privileges on the UDM device"
}

variable "udm_hostname" {
    type = string
    description = "The hostname or IP address of the UDM device"
}

variable "udm_uses_untrusted_ssl_cert" {
    type = bool
    description = "Whether or not the UDM device uses a self-signed or untrusted SSL certificate for its interface"
    default = true
}