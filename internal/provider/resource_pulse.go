// Package provider implements the Terraform provider for updown.io.
package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nastaliss/terraform-provider-updown/internal/updown"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func pulseResource() *schema.Resource {
	return &schema.Resource{
		Description: "`updown_pulse` defines a pulse (heartbeat) check for monitoring scheduled jobs and cron tasks",

		Create: pulseCreate,
		Read:   pulseRead,
		Delete: pulseDelete,
		Update: pulseUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: pulseImportState,
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
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				Description: "The URL to POST heartbeats to. Your scheduled job should POST to this URL on each successful run. " +
					"Note: the updown.io API redacts the secret key in GET responses. On import, the provider " +
					"toggles the enabled flag to force a real update and recover the full URL.",
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
	if err := d.Set("pulse_url", check.URL); err != nil {
		return fmt.Errorf("setting pulse_url: %s", err.Error())
	}

	return pulseRead(d, meta)
}

func pulseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*updown.Client)
	check, resp, err := client.Check.Get(d.Id())

	if err != nil {
		// The pulse check no longer exists on updown.io: drop it from state so
		// Terraform plans a recreate instead of failing.
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading pulse check from the API: %w", err)
	}

	// Verify this is actually a pulse check
	if check.Type != "pulse" {
		return fmt.Errorf("check %s is not a pulse check (type: %s)", d.Id(), check.Type)
	}

	for k, v := range map[string]interface{}{
		"alias":      check.Alias,
		"period":     check.Period,
		"enabled":    check.Enabled,
		"published":  check.Published,
		"mute_until": check.MuteUntil,
		"recipients": check.RecipientIDs,
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	// Note: pulse_url is intentionally not set here. The API redacts the secret
	// key on GET, so the value stored on create (or recovered on import) is
	// preserved. Read must stay side-effect-free, so URL recovery lives in the
	// importer (pulseImportState), not here.
	return nil
}

// pulseImportState recovers the unredacted pulse URL when importing an existing
// pulse check. The updown.io API only returns the full URL when a field actually
// changes, so we briefly toggle `enabled` to force a real update, capture the
// URL, then restore the original value. This runs only during import — never
// during a plan/refresh — so a plain `terraform plan` never mutates the remote
// resource.
func pulseImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*updown.Client)

	check, resp, err := client.Check.Get(d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("pulse check %s does not exist", d.Id())
		}
		return nil, fmt.Errorf("reading pulse check from the API: %w", err)
	}
	if check.Type != "pulse" {
		return nil, fmt.Errorf("check %s is not a pulse check (type: %s)", d.Id(), check.Type)
	}

	url, err := recoverPulseURL(client, d.Id(), check)
	if err != nil {
		return nil, err
	}
	if err := d.Set("pulse_url", url); err != nil {
		return nil, fmt.Errorf("setting pulse_url: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}

// recoverPulseURL forces the API to return the unredacted pulse URL by toggling
// the `enabled` flag, then always restores the original state via a deferred
// update — even if capturing the URL fails — so the check is never stranded in
// the wrong enabled state.
func recoverPulseURL(client *updown.Client, token string, check updown.Check) (url string, err error) {
	payload := updown.CheckItem{
		Type:         check.Type,
		Period:       check.Period,
		Enabled:      !check.Enabled,
		Published:    check.Published,
		Alias:        check.Alias,
		StringMatch:  check.StringMatch,
		MuteUntil:    check.MuteUntil,
		RecipientIDs: check.RecipientIDs,
	}

	defer func() {
		payload.Enabled = check.Enabled
		if _, _, rerr := client.Check.Update(token, payload); rerr != nil && err == nil {
			err = fmt.Errorf("restoring enabled flag after pulse URL recovery: %w", rerr)
		}
	}()

	updated, _, err := client.Check.Update(token, payload)
	if err != nil {
		return "", fmt.Errorf("recovering full pulse URL via update: %w", err)
	}

	return updated.URL, nil
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
