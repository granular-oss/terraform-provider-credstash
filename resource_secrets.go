package main

import (
	"context"
	"fmt"
	"time"

	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretCreate,
		ReadContext:   resourceSecretRead,
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of the secret",
			},
			"table": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "name of DynamoDB table where the secrets are stored",
				Default:     "",
			},
			"version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "version of the secrets",
				Default:     0,
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "encryption context for the secret",
			},
			"value": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"generate"},
				Description:   "The secret contents. Either `value` or `generate` must be defined.",
			},
			"generate": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				Description:   "Settings for autogenerating a secret. Either `value` or `generate` must be defined.",
				ConflictsWith: []string{"value"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"length": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The length of the secret to generate.",
						},
						"use_symbols": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether the secret should contain symbols.",
						},
						"charsets": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "Define the set of characters to randomly generate a password from. Options are all, alphanumeric, numeric, lowercase, uppercase, letters, symbols and human-readable.",
						},
						"min": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "Ensure that the generated secret contains at least n characters from the given character set. Note that adding constraints reduces the strength of the secret.",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*credstash.Client)

	// Warning or errors can be collected in a slice type
	//var diags diag.Diagnostics

	name := d.Get("name").(string)
	version := d.Get("version").(int)
	table := d.Get("table").(string)
	value := d.Get("value").(string)
	generateList := d.Get("generate").([]interface{})
	if value == "" && len(generateList) == 0 {
		return diag.FromErr(fmt.Errorf("either 'value' or 'generate' must be specified"))
	}

	context := credstash.NewEncryptionContextValue()
	for k, v := range d.Get("context").(map[string]interface{}) {
		stringValue := fmt.Sprintf("%v", v)
		(*context)[k] = &stringValue
	}

	if len(generateList) > 0 {
		settings := generateList[0].(map[string]interface{})
		useSymbols := settings["use_symbols"].(bool)
		length := settings["length"].(int)
		charsetSet := settings["charsets"].(*schema.Set)
		minRuleMap := settings["min"].(map[string]interface{})
		charsets := charsetSet.List()
		var err error
		value, err = client.GenerateRandomSecret(length, useSymbols, charsets, minRuleMap)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	var paddedVersion string
	if version == 0 {
		paddedVersion = client.PaddedInt(1)
	} else {
		paddedVersion = client.PaddedInt(version)
	}
	err := client.PutSecret(table, name, value, paddedVersion, context)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("version", version)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("value", string(value))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(hash(value))
	return resourceSecretRead(ctx, d, m)
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*credstash.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	//Make Imports work, by reading ID if name is empty
	if name == "" {
		name = d.Id()
	}
	version := d.Get("version").(int)
	table := d.Get("table").(string)

	context := credstash.NewEncryptionContextValue()
	for k, v := range d.Get("context").(map[string]interface{}) {
		stringValue := fmt.Sprintf("%v", v)
		(*context)[k] = &stringValue
	}

	var value *credstash.DecryptedCredential
	var err error

	tflog.Debug(ctx, "resourceSecretRead getting secret", map[string]interface{}{
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

	d.SetId(hash(value.Secret))
	d.Set("value", value.Secret)
	d.Set("table", table)
	d.Set("version", version)
	d.Set("name", name)
	d.Set("generate", d.Get("generate"))

	return diags
}

func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*credstash.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// secretID := d.Id()
	name := d.Get("name").(string)
	table := d.Get("table").(string)

	err := c.DeleteSecret(table, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*credstash.Client)

	if d.HasChange("value") || d.HasChange("generate") || d.HasChange("version") {

		name := d.Get("name").(string)
		table := d.Get("table").(string)
		value := d.Get("value").(string)
		version := d.Get("version").(int)

		generateList := d.Get("generate").([]interface{})
		if value == "" && len(generateList) == 0 {
			return diag.FromErr(fmt.Errorf("either 'value' or 'generate' must be specified"))
		}

		context := credstash.NewEncryptionContextValue()
		for k, v := range d.Get("context").(map[string]interface{}) {
			stringValue := fmt.Sprintf("%v", v)
			(*context)[k] = &stringValue
		}

		paddedVersion, err := c.ResolveVersion(table, name, version)

		if err != nil {
			return diag.FromErr(err)
		}

		if len(generateList) > 0 {
			settings := generateList[0].(map[string]interface{})
			useSymbols := settings["use_symbols"].(bool)
			length := settings["length"].(int)
			charsetSet := settings["charsets"].(*schema.Set)
			minRuleMap := settings["min"].(map[string]interface{})
			charsets := charsetSet.List()
			var err error
			value, err = c.GenerateRandomSecret(length, useSymbols, charsets, minRuleMap)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		err = c.PutSecret(table, name, value, paddedVersion, context)

		if err != nil {
			return diag.FromErr(err)
		}

		//Update the secret version if we are not storing 0.
		// if version != 0 {
		// 	intVersion, err := strconv.Atoi(paddedVersion)
		// 	if err != nil {
		// 		return diag.FromErr(err)
		// 	}
		// 	d.Set("version", intVersion)
		// }

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceSecretRead(ctx, d, m)
}
