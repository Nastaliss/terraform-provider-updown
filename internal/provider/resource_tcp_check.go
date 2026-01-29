// Package provider implements the Terraform provider for updown.io.
package provider

import (
	"fmt"
	"strings"

	"github.com/antoineaugusti/updown"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tcpCheckResource() *schema.Resource {
	return &schema.Resource{
		Description: "`updown_tcp_check` defines a TCP or TCPS (TLS) check for monitoring TCP endpoints",

		Create: tcpCheckCreate,
		Read:   tcpCheckRead,
		Delete: tcpCheckDelete,
		Update: tcpCheckUpdate,
		Exists: tcpCheckExists,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The TCP endpoint to monitor (e.g., tcp://host:port or tcps://host:port for TLS).",
			},
			"period": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Interval in seconds (15, 30, 60, 120, 300, 600, 1800 or 3600).",
				Default:     60,
			},
			"apdex_t": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "APDEX threshold in seconds (0.125, 0.25, 0.5, 1.0 or 2.0).",
				Default:     0.5,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Is the check enabled (true or false).",
				Default:     true,
			},
			"published": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Shall the status page be public (true or false).",
				Default:     false,
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Human readable name.",
			},
			"mute_until": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Mute notifications until given time, accepts a time, 'recovery' or 'forever'.",
			},
			"disabled_locations": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Disabled monitoring locations. It's a list of abbreviated location names.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"recipients": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Selected alert recipients. It's an array of recipient IDs you can get from the recipients API.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func constructTCPCheckPayload(d *schema.ResourceData) updown.CheckItem {
	payload := updown.CheckItem{}

	if v, ok := d.GetOk("url"); ok {
		url := v.(string)
		payload.URL = url
		// Set the type based on the URL scheme
		if strings.HasPrefix(url, "tcps://") {
			payload.Type = "tcps"
		} else {
			payload.Type = "tcp"
		}
	}

	if v, ok := d.GetOk("period"); ok {
		payload.Period = v.(int)
	}

	if v, ok := d.GetOk("apdex_t"); ok {
		payload.Apdex = v.(float64)
	}

	if v, ok := d.GetOk("enabled"); ok {
		payload.Enabled = v.(bool)
	}

	if v, ok := d.GetOk("published"); ok {
		payload.Published = v.(bool)
	}

	if v, ok := d.GetOk("alias"); ok {
		payload.Alias = v.(string)
	}

	if v, ok := d.GetOk("mute_until"); ok {
		payload.MuteUntil = v.(string)
	}

	if v, ok := d.GetOk("disabled_locations"); ok {
		interfaceSlice := v.(*schema.Set).List()
		var stringSlice []string
		for s := range interfaceSlice {
			stringSlice = append(stringSlice, interfaceSlice[s].(string))
		}
		payload.DisabledLocations = stringSlice
	}

	if v, ok := d.GetOk("recipients"); ok {
		interfaceSlice := v.(*schema.Set).List()
		var stringSlice []string
		for s := range interfaceSlice {
			stringSlice = append(stringSlice, interfaceSlice[s].(string))
		}
		payload.RecipientIDs = stringSlice
	}

	return payload
}

func tcpCheckCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)

	check, _, err := client.Check.Add(constructTCPCheckPayload(d))
	if err != nil {
		return fmt.Errorf("creating TCP check with the API: %w", err)
	}

	d.SetId(check.Token)

	return tcpCheckRead(d, meta)
}

func tcpCheckRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)
	check, _, err := client.Check.Get(d.Id())

	if err != nil {
		return fmt.Errorf("reading TCP check from the API: %w", err)
	}

	// Verify this is actually a TCP or TCPS check
	if check.Type != "tcp" && check.Type != "tcps" {
		return fmt.Errorf("check %s is not a TCP check (type: %s)", d.Id(), check.Type)
	}

	for k, v := range map[string]interface{}{
		"url":                check.URL,
		"period":             check.Period,
		"apdex_t":            check.Apdex,
		"enabled":            check.Enabled,
		"published":          check.Published,
		"alias":              check.Alias,
		"mute_until":         check.MuteUntil,
		"disabled_locations": check.DisabledLocations,
		"recipients":         check.RecipientIDs,
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

func tcpCheckUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)

	_, _, err := client.Check.Update(d.Id(), constructTCPCheckPayload(d))
	if err != nil {
		return fmt.Errorf("updating TCP check with the API: %w", err)
	}

	return tcpCheckRead(d, meta)
}

func tcpCheckDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)
	checkDeleted, _, err := client.Check.Remove(d.Id())

	if err != nil {
		return fmt.Errorf("removing TCP check from the API: %w", err)
	}

	if !checkDeleted {
		return fmt.Errorf("TCP check couldn't be deleted")
	}

	return nil
}

func tcpCheckExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := tcpCheckRead(d, meta)
	return err == nil, err
}
