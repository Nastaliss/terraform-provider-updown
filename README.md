# terraform-provider-updown

[![PkgGoDev](https://pkg.go.dev/badge/github.com/Nastaliss/terraform-provider-updown)](https://pkg.go.dev/mod/github.com/Nastaliss/terraform-provider-updown)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nastaliss/terraform-provider-updown)](https://goreportcard.com/report/github.com/Nastaliss/terraform-provider-updown)
[![test](https://github.com/Nastaliss/terraform-provider-updown/actions/workflows/test.yml/badge.svg)](https://github.com/Nastaliss/terraform-provider-updown/actions/workflows/test.yml)
[![release](https://github.com/Nastaliss/terraform-provider-updown/actions/workflows/release.yml/badge.svg)](https://github.com/Nastaliss/terraform-provider-updown/actions/workflows/release.yml)

Terraform provider for [updown.io](https://updown.io)

## Docs

https://registry.terraform.io/providers/Nastaliss/updown/latest/docs

## Resources

| TYPE | NAME | DESCRIPTION |
|---|---|---|
| **data** |`updown_nodes`| Returns the list of testing nodes ipv4 and ipv6 addresses |
| **resource** |`updown_check`| Creates a check |
| **resource** |`updown_recipient`| Creates a recipient |
| **resource** |`updown_tcp_check`| Creates a TCP/TCPS check |

## Example usage

```hcl
# Import the provider
terraform {
  required_providers {
    updown = {
      source = "Nastaliss/updown"
    }
  }
}

# Configure it
provider "updown" {
  # Can also be set using UPDOWN_API_KEY env variable.
  api_key = "<YOUR_UPDOWN_API_KEY>"
}

# Add a recipient
resource "updown_recipient" "myrecipient" {
  type  = "email"
  value = "foo@bar.baz"
}

# Create a check
resource "updown_check" "mywebsite" {
  alias        = "https://example.com"
  apdex_t      = 1.0
  enabled      = true
  period       = 30
  published    = true
  url          = "https://test.example.com/healthz"
  string_match = "OK"
  mute_until   = "tomorrow"

  recipients = [
    updown_recipient.myrecipient.id,
  ]

  disabled_locations = [
    "mia",
  ]

  custom_headers = {
    "X-GREAT-HEADER" = "yay!"
  }
}

# Create a TCP check
resource "updown_tcp_check" "my_database" {
  alias   = "PostgreSQL Database"
  url     = "tcp://db.example.com:5432"
  period  = 60
  enabled = true
}

# Output ipv4 and ipv6 nodes addresses list
data "updown_nodes" "global" {}

output "updown_nodes_ipv4" {
  value = data.updown_nodes.global.ipv4
}

output "updown_nodes_ipv6" {
  value = data.updown_nodes.global.ipv6
}
```

## Using with OpenTofu OCI mirror

This provider is published as an OCI artifact to GitHub Container Registry. To use it with OpenTofu's `oci_mirror` provider installation, add the following to your `~/.terraformrc` (or `~/.tofurc`):

```hcl
provider_installation {
  oci_mirror {
    repository_template = "ghcr.io/nastaliss/opentofu-providers/${namespace}/${type}"
    include             = ["registry.opentofu.org/nastaliss/*"]
  }
  direct {
    exclude = ["registry.opentofu.org/nastaliss/*"]
  }
}
```

## Building the provider

```bash
~$ export PROVIDER_PATH=${GOPATH}/src/github.com/Nastaliss/terraform-provider-updown
~$ mkdir -p ${PROVIDER_PATH}; cd ${PROVIDER_PATH}
~$ git clone git@github.com:Nastaliss/terraform-provider-updown .
~$ make install
```

## TODO

- Add tests, need to figure out how to get a mocking endpoint
