// Package provider implements the Terraform provider for updown.io.
package provider

import (
	"fmt"

	"github.com/antoineaugusti/updown"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func pulseResource() *schema.Resource {
	return &schema.Resource{
		Description: "`updown_pulse` defines a pulse (heartbeat) check for monitoring scheduled jobs and cron tasks",

		Create: pulseCreate,
		Read:   pulseRead,
		Delete: pulseDelete,
		Update: pulseUpdate,
		Exists: pulseExists,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Human readable name for the pulse check.",
			},
			"period": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Expected interval in seconds between heartbeats (15 to 2678400, i.e., 15 seconds to 31 days).",
			},
			"apdex_t": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "APDEX threshold in seconds (0.125, 0.25, 0.5, 1.0, 2.0, 4.0 or 8.0).",
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
			"mute_until": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Mute notifications until given time, accepts a time, 'recovery' or 'forever'.",
			},
			"recipients": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Selected alert recipients. It's an array of recipient IDs you can get from the recipients API.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pulse_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to POST heartbeats to. Your scheduled job should POST to this URL on each successful run.",
			},
		},
	}
}

func constructPulsePayload(d *schema.ResourceData) updown.CheckItem {
	payload := updown.CheckItem{
		Type: "pulse",
	}

	if v, ok := d.GetOk("alias"); ok {
		payload.Alias = v.(string)
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

	if v, ok := d.GetOk("mute_until"); ok {
		payload.MuteUntil = v.(string)
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

func pulseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)

	check, _, err := client.Check.Add(constructPulsePayload(d))
	if err != nil {
		return fmt.Errorf("creating pulse check with the API: %s", err.Error())
	}

	d.SetId(check.Token)

	return pulseRead(d, meta)
}

func pulseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)
	check, _, err := client.Check.Get(d.Id())

	if err != nil {
		return fmt.Errorf("reading pulse check from the API: %s", err.Error())
	}

	// Verify this is actually a pulse check
	if check.Type != "pulse" {
		return fmt.Errorf("check %s is not a pulse check (type: %s)", d.Id(), check.Type)
	}

	for k, v := range map[string]interface{}{
		"alias":      check.Alias,
		"period":     check.Period,
		"apdex_t":    check.Apdex,
		"enabled":    check.Enabled,
		"published":  check.Published,
		"mute_until": check.MuteUntil,
		"recipients": check.RecipientIDs,
		"pulse_url":  fmt.Sprintf("https://updown.io/api/checks/%s/pulse", check.Token),
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

func pulseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)

	_, _, err := client.Check.Update(d.Id(), constructPulsePayload(d))
	if err != nil {
		return fmt.Errorf("updating pulse check with the API: %s", err.Error())
	}

	return pulseRead(d, meta)
}

func pulseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)
	checkDeleted, _, err := client.Check.Remove(d.Id())

	if err != nil {
		return fmt.Errorf("removing pulse check from the API: %s", err.Error())
	}

	if !checkDeleted {
		return fmt.Errorf("pulse check couldn't be deleted")
	}

	return nil
}

func pulseExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := pulseRead(d, meta)
	return err == nil, err
}
