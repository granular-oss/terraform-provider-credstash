package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of the secret",
			},
			"version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "version of the secrets",
				Default:     0,
			},
			"table": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "name of DynamoDB table where the secrets are stored",
				Default:     "",
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "encryption context for the secret",
			},
			"value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "value of the secret",
				Sensitive:   true,
			},
		},
	}
}

func dataSourceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := meta.(*credstash.Client)

	name := d.Get("name").(string)
	version := d.Get("version").(int)
	table := d.Get("table").(string)

	context := credstash.NewEncryptionContextValue()
	for k, v := range d.Get("context").(map[string]interface{}) {
		stringValue := fmt.Sprintf("%v", v)
		(*context)[k] = &stringValue
	}

	var value *credstash.DecryptedCredential
	var err error

	tflog.Debug(ctx, "dataSourceSecretRead getting secret", map[string]interface{}{
		"name":    name,
		"version": version,
		"table":   table,
		"context": context,
	})

	if version == 0 {
		value, err = client.GetHighestVersionSecret(table, name, context)

	} else {
		value, err = client.GetSecret(name, table, client.PaddedInt(version), context)

	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("value", value.Secret)
	d.SetId(hash(value.Secret))

	return diags
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
