package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var networkTunnelDeviceTypes = []string{
	"ASA",
	"AWS S2S VPN",
	"AZURE S2S VPN",
	"FTD",
	"ISR",
	"Meraki MX",
	"Viptela cEdge",
	"Viptela vEdge",
	"other",
}

func resourceNetworkTunnelGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkTunnelGroupCreate,
		ReadContext:   resourceNetworkTunnelGroupRead,
		UpdateContext: resourceNetworkTunnelGroupUpdate,
		DeleteContext: resourceNetworkTunnelGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Network Tunnel Group.",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region for the Network Tunnel Group.",
			},
			"device_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "other",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkTunnelDeviceTypes, false),
				Description:  "The type of device that establishes the network tunnel.",
			},
			"auth_id_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"auth_id_prefix", "auth_id_ips"},
				Description:  "A string value used as authIdPrefix.",
			},
			"auth_id_ips": {
				Type:         schema.TypeList,
				Optional:     true,
				AtLeastOneOf: []string{"auth_id_prefix", "auth_id_ips"},
				MinItems:     2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of IPs used as authIdPrefix. Cisco requires at least two IPs when this format is used.",
			},
			"passphrase": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Passphrase for primary and secondary tunnels.",
			},
			"routing": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"static", "bgp", "nat"}, false),
						},
						"network_cidrs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"as_number": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bgp_hop_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"bgp_neighbor_cidrs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"bgp_server_subnets": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Description: "Optional routing configuration.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID returned by Secure Access.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the Network Tunnel Group.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp from API.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last modification timestamp from API.",
			},
		},
	}
}

func resourceNetworkTunnelGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SecureAccessClient)

	reqBody, err := buildCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	created, err := client.CreateNetworkTunnelGroup(ctx, reqBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(formatID(created.ID))
	return resourceNetworkTunnelGroupRead(ctx, d, m)
}

func resourceNetworkTunnelGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SecureAccessClient)

	group, err := client.GetNetworkTunnelGroup(ctx, d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := setStateFromNetworkTunnelGroup(d, group); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkTunnelGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SecureAccessClient)
	operations := make([]patchOperation, 0)

	if d.HasChange("name") {
		operations = append(operations, patchOperation{Op: "replace", Path: "/name", Value: d.Get("name").(string)})
	}
	if d.HasChange("region") {
		operations = append(operations, patchOperation{Op: "replace", Path: "/region", Value: d.Get("region").(string)})
	}
	if d.HasChange("passphrase") {
		operations = append(operations, patchOperation{Op: "replace", Path: "/passphrase", Value: d.Get("passphrase").(string)})
	}
	if d.HasChange("auth_id_prefix") || d.HasChange("auth_id_ips") {
		authIDPrefix, err := expandAuthIDPrefix(d)
		if err != nil {
			return diag.FromErr(err)
		}
		operations = append(operations, patchOperation{Op: "replace", Path: "/authIdPrefix", Value: authIDPrefix})
	}
	if d.HasChange("routing") {
		routing, err := expandRouting(d.Get("routing").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		if routing == nil {
			routing = &routingPayload{Type: "nat", Data: ""}
		}
		operations = append(operations, patchOperation{Op: "replace", Path: "/routing", Value: routing})
	}

	if len(operations) == 0 {
		return resourceNetworkTunnelGroupRead(ctx, d, m)
	}

	updated, err := client.UpdateNetworkTunnelGroup(ctx, d.Id(), operations)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setStateFromNetworkTunnelGroup(d, updated); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkTunnelGroupRead(ctx, d, m)
}

func resourceNetworkTunnelGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SecureAccessClient)

	err := client.DeleteNetworkTunnelGroup(ctx, d.Id())
	if err != nil && !IsNotFoundError(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func buildCreateRequest(d *schema.ResourceData) (networkTunnelGroupRequest, error) {
	authIDPrefix, err := expandAuthIDPrefix(d)
	if err != nil {
		return networkTunnelGroupRequest{}, err
	}

	routing, err := expandRouting(d.Get("routing").([]interface{}))
	if err != nil {
		return networkTunnelGroupRequest{}, err
	}

	return networkTunnelGroupRequest{
		Name:         d.Get("name").(string),
		Region:       d.Get("region").(string),
		DeviceType:   d.Get("device_type").(string),
		AuthIDPrefix: authIDPrefix,
		Passphrase:   d.Get("passphrase").(string),
		Routing:      routing,
	}, nil
}

func expandAuthIDPrefix(d *schema.ResourceData) (interface{}, error) {
	if v := strings.TrimSpace(d.Get("auth_id_prefix").(string)); v != "" {
		return v, nil
	}

	ips := toStringSlice(d.Get("auth_id_ips").([]interface{}))
	if len(ips) > 0 {
		return ips, nil
	}

	return nil, fmt.Errorf("one of auth_id_prefix or auth_id_ips must be set")
}

func expandRouting(routingInput []interface{}) (*routingPayload, error) {
	if len(routingInput) == 0 || routingInput[0] == nil {
		return nil, nil
	}

	cfg := routingInput[0].(map[string]interface{})
	routingType := cfg["type"].(string)

	var data interface{}
	switch routingType {
	case "nat":
		data = ""
	case "static":
		networkCIDRs := toStringSlice(cfg["network_cidrs"].([]interface{}))
		if len(networkCIDRs) == 0 {
			return nil, fmt.Errorf("routing.network_cidrs must be set when routing.type is static")
		}
		data = map[string]interface{}{"networkCIDRs": networkCIDRs}
	case "bgp":
		asNumber := strings.TrimSpace(cfg["as_number"].(string))
		if asNumber == "" {
			return nil, fmt.Errorf("routing.as_number must be set when routing.type is bgp")
		}
		bgpData := map[string]interface{}{"asNumber": asNumber}

		if hopCount := cfg["bgp_hop_count"].(int); hopCount > 0 {
			bgpData["bgpHopCount"] = hopCount
		}
		if neighborCIDRs := toStringSlice(cfg["bgp_neighbor_cidrs"].([]interface{})); len(neighborCIDRs) > 0 {
			bgpData["bgpNeighborCIDRs"] = neighborCIDRs
		}
		if serverSubnets := toStringSlice(cfg["bgp_server_subnets"].([]interface{})); len(serverSubnets) > 0 {
			bgpData["bgpServerSubnets"] = serverSubnets
		}

		data = bgpData
	default:
		return nil, fmt.Errorf("unsupported routing.type %q", routingType)
	}

	return &routingPayload{Type: routingType, Data: data}, nil
}

func setStateFromNetworkTunnelGroup(d *schema.ResourceData, group *networkTunnelGroup) error {
	if err := d.Set("name", group.Name); err != nil {
		return err
	}
	if err := d.Set("region", group.Region); err != nil {
		return err
	}
	if err := d.Set("device_type", group.DeviceType); err != nil {
		return err
	}
	if err := d.Set("organization_id", fmt.Sprintf("%d", group.OrganizationID)); err != nil {
		return err
	}
	if err := d.Set("status", group.Status); err != nil {
		return err
	}
	if err := d.Set("created_at", group.CreatedAt); err != nil {
		return err
	}
	if err := d.Set("modified_at", group.ModifiedAt); err != nil {
		return err
	}
	if err := d.Set("routing", flattenRouting(group.Routing)); err != nil {
		return err
	}

	return nil
}

func flattenRouting(r *routingResponse) []interface{} {
	if r == nil {
		return nil
	}

	routing := map[string]interface{}{
		"type": r.Type,
	}

	switch r.Type {
	case "static":
		var payload struct {
			NetworkCIDRs []string `json:"networkCIDRs"`
		}
		if err := json.Unmarshal(r.Data, &payload); err == nil {
			routing["network_cidrs"] = payload.NetworkCIDRs
		}
	case "bgp":
		var payload struct {
			AsNumber         string   `json:"asNumber"`
			BGPHopCount      int      `json:"bgpHopCount"`
			BGPNeighborCIDRs []string `json:"bgpNeighborCIDRs"`
			BGPServerSubnets []string `json:"bgpServerSubnets"`
		}
		if err := json.Unmarshal(r.Data, &payload); err == nil {
			routing["as_number"] = payload.AsNumber
			if payload.BGPHopCount > 0 {
				routing["bgp_hop_count"] = payload.BGPHopCount
			}
			if len(payload.BGPNeighborCIDRs) > 0 {
				routing["bgp_neighbor_cidrs"] = payload.BGPNeighborCIDRs
			}
			if len(payload.BGPServerSubnets) > 0 {
				routing["bgp_server_subnets"] = payload.BGPServerSubnets
			}
		}
	}

	return []interface{}{routing}
}

func toStringSlice(values []interface{}) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, v := range values {
		s := strings.TrimSpace(v.(string))
		if s != "" {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
