# Terraform Provider for Ubiquiti Dream Machine

This repository is a custom Terraform provider for working with Ubiquiti Dream Machine (UDM) devices.

Please note that this project is in a very early stage and subject to frequent changes.  As of today, the API for UDM devices is unpublished by Ubiquiti, so much of the work in this provider was done through reverse engineering by making changes through the UI and inspecting calls via Chrome Developer Tools.

This provider is currently in active development and is **NOT** ready for production use.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

1. You'll need to update your `$HOME/.terraformrc` file with the following:

   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/josh-hogle/ubiquiti-udm" = "/your/$GOPATH/bin/folder"
     }
     direct {}
   }
   ```

1. In your `main.tf` file, include the following:

   ```hcl
   terraform {
     required_providers {
       udm = {
         source = "registry.terraform.io/josh-hogle/ubiquiti-udm"
       }
     }
   }

   provider "udm" {
     hostname = "your.udm.ip.address"
     username = "udm_local_admin_username"
     password = "udm_local_admin_password"
     ignore_untrusted_ssl_certificate = true
   }
   ```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.
