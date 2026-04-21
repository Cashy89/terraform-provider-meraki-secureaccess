# merakisecureaccess_network_tunnel_group Resource

Creates and manages a Cisco Secure Access Network Tunnel Group.

## Example Usage

```hcl
resource "merakisecureaccess_network_tunnel_group" "example" {
  name           = "Branch Tunnel Group"
  region         = "us-east-1"
  device_type    = "ASA"
  auth_id_prefix = "branch-tunnels"
  passphrase     = var.network_tunnel_passphrase

  routing {
    type          = "static"
    network_cidrs = ["10.10.0.0/16", "10.20.0.0/16"]
  }
}
```

## Argument Reference

- `name` - (Required) Name of the Network Tunnel Group.
- `region` - (Required) Region for the Network Tunnel Group.
- `device_type` - (Optional) Device type for the tunnel group. Defaults to `other`.
  Changing this value forces a new resource.
- `auth_id_prefix` - (Optional) String-based authentication ID prefix.
- `auth_id_ips` - (Optional) List-based authentication ID prefix using IP addresses. Minimum 2 items.
- `passphrase` - (Required, Sensitive) Passphrase for the primary and secondary tunnels.
- `routing` - (Optional) Routing configuration block.

Exactly one of `auth_id_prefix` or `auth_id_ips` must be set.

### `routing` block

- `type` - (Required) Routing type. One of `static`, `bgp`, or `nat`.
- `network_cidrs` - (Optional) CIDRs for static routing. Required when `type = "static"`.
- `as_number` - (Optional) BGP AS number. Required when `type = "bgp"`.
- `bgp_hop_count` - (Optional) BGP hop count (1-64).
- `bgp_neighbor_cidrs` - (Optional) BGP neighbor CIDRs.
- `bgp_server_subnets` - (Optional) BGP server subnets.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Network Tunnel Group ID.
- `organization_id` - Organization ID returned by Secure Access.
- `status` - Current tunnel group status.
- `created_at` - API timestamp when the tunnel group was created.
- `modified_at` - API timestamp when the tunnel group was last modified.

## Import

Import using the Network Tunnel Group ID:

```bash
terraform import merakisecureaccess_network_tunnel_group.example <network_tunnel_group_id>
```
