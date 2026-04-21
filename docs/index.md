# secureaccessntg Provider

The `secureaccessntg` provider manages Cisco Secure Access resources through the Deployments API.

Authentication follows the Cisco Secure Access API key pattern (`key_id` and `key_secret`).

## Example Usage

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

## Argument Reference

- `key_id` - (Optional, Sensitive) Secure Access API key ID. Can also be set with `CISCOSECUREACCESS_KEY_ID`.
- `key_secret` - (Optional, Sensitive) Secure Access API key secret. Can also be set with `CISCOSECUREACCESS_KEY_SECRET`.
- `client_id` - (Optional, Sensitive, Deprecated) Legacy alias for `key_id`.
- `client_secret` - (Optional, Sensitive, Deprecated) Legacy alias for `key_secret`.
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
