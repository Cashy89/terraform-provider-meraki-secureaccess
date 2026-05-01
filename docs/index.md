# merakisecureaccess Provider

The `merakisecureaccess` provider manages Cisco Secure Access resources through the Deployments API.

## Example Usage

```hcl
terraform {
  required_providers {
    merakisecureaccess = {
      source  = "Cashy89/meraki-secureaccess"
      version = "0.0.6"
    }
  }
}

provider "merakisecureaccess" {
  client_id     = var.secure_access_client_id
  client_secret = var.secure_access_client_secret

  # Optional
  # organization_id = "123456"
  # base_url        = "https://api.sse.cisco.com/deployments/v2"
  # token_url       = "https://api.sse.cisco.com/auth/v2/token"
  # scope           = "deployments.networktunnelgroups:read deployments.networktunnelgroups:write"
}
```

## Argument Reference

- `client_id` - (Optional, Sensitive) Secure Access API client ID. Can also be set with `SECURE_ACCESS_CLIENT_ID`.
- `client_secret` - (Optional, Sensitive) Secure Access API client secret. Can also be set with `SECURE_ACCESS_CLIENT_SECRET`.
- `organization_id` - (Optional) Child organization ID. Can also be set with `SECURE_ACCESS_ORGANIZATION_ID`.
- `base_url` - (Optional) Base URL for the Secure Access Deployments API. Defaults to `https://api.sse.cisco.com/deployments/v2`.
- `token_url` - (Optional) OAuth token URL. Defaults to `https://api.sse.cisco.com/auth/v2/token`.
- `scope` - (Optional) OAuth scope override for client credentials flow.

## Environment Variables

- `CISCOSECUREACCESS_KEY_ID`
- `CISCOSECUREACCESS_KEY_SECRET`
- `SECURE_ACCESS_ORGANIZATION_ID`

Legacy variables are also supported:
- `SECURE_ACCESS_CLIENT_ID` (deprecated alias for `key_id`)
- `SECURE_ACCESS_CLIENT_SECRET` (deprecated alias for `key_secret`)
