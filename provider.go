package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"CISCOSECUREACCESS_KEY_ID", "SECURE_ACCESS_CLIENT_ID"}, nil),
				Description: "Secure Access API key ID. Can also be set with CISCOSECUREACCESS_KEY_ID.",
			},
			"key_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"CISCOSECUREACCESS_KEY_SECRET", "SECURE_ACCESS_CLIENT_SECRET"}, nil),
				Description: "Secure Access API key secret. Can also be set with CISCOSECUREACCESS_KEY_SECRET.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SECURE_ACCESS_CLIENT_ID", nil),
				Description: "Deprecated. Use key_id instead.",
				Deprecated:  "Use key_id instead of client_id.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SECURE_ACCESS_CLIENT_SECRET", nil),
				Description: "Deprecated. Use key_secret instead.",
				Deprecated:  "Use key_secret instead of client_secret.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SECURE_ACCESS_ORGANIZATION_ID", nil),
				Description: "Optional child organization ID for multi-org environments. When set, this value is sent as X-Umbrella-OrgId during token generation.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://api.sse.cisco.com/deployments/v2",
				Description: "Base URL for Secure Access Deployments API.",
			},
			"token_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://api.sse.cisco.com/auth/v2/token",
				Description: "OAuth token URL for Secure Access API.",
			},
			"scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Optional OAuth scope override for client credentials flow.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"merakisecureaccess_network_tunnel_group": resourceNetworkTunnelGroup(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	keyID := d.Get("key_id").(string)
	keySecret := d.Get("key_secret").(string)
	if keyID == "" {
		keyID = d.Get("client_id").(string)
	}
	if keySecret == "" {
		keySecret = d.Get("client_secret").(string)
	}
	orgID := d.Get("organization_id").(string)
	baseURL := d.Get("base_url").(string)
	tokenURL := d.Get("token_url").(string)
	scope := d.Get("scope").(string)

	if keyID == "" || keySecret == "" {
		return nil, diag.Errorf("provider requires key_id and key_secret (or CISCOSECUREACCESS_KEY_ID and CISCOSECUREACCESS_KEY_SECRET environment variables)")
	}

	client := NewClient(keyID, keySecret, orgID, baseURL, tokenURL, scope)
	return client, diags
}
