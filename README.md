# terraform-provider-meraki-secureaccess

Terraform provider for Cisco Secure Access Network Tunnel Groups.

This provider manages the lifecycle of Network Tunnel Groups using the Secure Access Deployments API:
- Create: `POST /networktunnelgroups`
- Read: `GET /networktunnelgroups/{id}`
- Update: `PATCH /networktunnelgroups/{id}`
- Delete: `DELETE /networktunnelgroups/{id}`

## Requirements

- Terraform >= 1.0
- Go >= 1.22 (for local builds)
- Secure Access API credentials (key ID and key secret)

## Provider Configuration

```hcl
terraform {
  required_providers {
    secureaccessntg = {
      source  = "Cashy89/meraki-secureaccess"
      version = "~> 0.0"
    }
  }
}

provider "secureaccessntg" {
  key_id        = var.secure_access_key_id
  key_secret    = var.secure_access_key_secret

  # Optional
  # organization_id = "123456"
  # base_url        = "https://api.sse.cisco.com/deployments/v2"
  # token_url       = "https://api.sse.cisco.com/auth/v2/token"
  # scope           = "deployments.networktunnelgroups:read deployments.networktunnelgroups:write"
}
```

Environment variables are also supported:
- `CISCOSECUREACCESS_KEY_ID`
- `CISCOSECUREACCESS_KEY_SECRET`
- `SECURE_ACCESS_ORGANIZATION_ID`

Legacy aliases are still accepted for backward compatibility:
- `client_id` / `client_secret` provider arguments
- `SECURE_ACCESS_CLIENT_ID` / `SECURE_ACCESS_CLIENT_SECRET` environment variables

## Resource Example

```hcl
resource "secureaccessntg_network_tunnel_group" "example" {
  name         = "Branch Tunnel Group"
  region       = "us-east-1"
  device_type  = "ASA"
  auth_id_prefix = "branch-tunnels"
  passphrase   = var.network_tunnel_passphrase

  routing {
    type          = "static"
    network_cidrs = ["10.10.0.0/16", "10.20.0.0/16"]
  }
}
```

## Routing Options

- `nat`: no additional fields required.
- `static`: set `network_cidrs`.
- `bgp`: set `as_number`, optionally `bgp_hop_count`, `bgp_neighbor_cidrs`, and `bgp_server_subnets`.

## Import

```bash
terraform import secureaccessntg_network_tunnel_group.example <network_tunnel_group_id>
```
