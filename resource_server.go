package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/secrethub/secrethub-go/pkg/randchar"
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "version of the secrets",
				Default:     "",
			},
			"autoversion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Automatically increment the version of the credential to be stored. This option causes the version arguement to be ignored. (This option will fail if the currently stored version is not numeric.)",
				Default:     false,
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
							Deprecated:  "use the charsets attribute instead",
							Optional:    true,
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
	}
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*credstash.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	version := d.Get("version").(string)
	table := d.Get("table").(string)
	value := d.Get("value").(string)
	generateList := d.Get("generate").([]interface{})
	if value == "" && len(generateList) == 0 {
		return diag.FromErr(fmt.Errorf("either 'value' or 'generate' must be specified"))
	}

	context := make(map[string]*string)
	for k, v := range d.Get("context").(map[string]interface{}) {
		stringValue := fmt.Sprintf("%v", v)
		context[k] = &stringValue
	}

	if len(generateList) > 0 {
		settings := generateList[0].(map[string]interface{})
		useSymbols := settings["use_symbols"].(bool)
		length := settings["length"].(int)
		charsetSet := settings["charsets"].(*schema.Set)
		charsets := charsetSet.List()
		charset := randchar.Charset{}
		if len(charsets) == 0 {
			charset = randchar.Alphanumeric
		}
		if useSymbols {
			charset = charset.Add(randchar.Symbols)
		}
		for _, charsetName := range charsets {
			set, found := randchar.CharsetByName(charsetName.(string))
			if !found {
				return diag.FromErr(fmt.Errorf("unknown charset: %s", charsetName))
			}
			charset = charset.Add(set)
		}

		minRuleMap := settings["min"].(map[string]interface{})
		var minRules []randchar.Option
		for charset, min := range minRuleMap {
			n := min.(int)
			set, found := randchar.CharsetByName(charset)
			if !found {
				return diag.FromErr(fmt.Errorf("unknown charset: %s", charset))
			}
			minRules = append(minRules, randchar.Min(n, set))
		}

		var err error
		rand, err := randchar.NewRand(charset, minRules...)
		if err != nil {
			return diag.FromErr(err)
		}
		byteValue, err := rand.Generate(length)
		if err != nil {
			return diag.FromErr(err)
		}
		value = string(byteValue)
	}

	err := client.PutSecret(table, name, value, version, context)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(hash(value))

	return diags
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*credstash.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	version := d.Get("version").(string)
	table := d.Get("table").(string)

	context := make(map[string]string)
	for k, v := range d.Get("context").(map[string]interface{}) {
		context[k] = fmt.Sprintf("%v", v)
	}

	log.Printf("[DEBUG] Getting secret for name=%q table=%q version=%q context=%+v", name, table, version, context)
	value, err := client.GetSecret(name, table, version, context)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(hash(value))

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

	if d.HasChange("value") || d.HasChange("generate") {

		name := d.Get("name").(string)
		table := d.Get("table").(string)
		value := d.Get("value").(string)

		context := make(map[string]*string)
		for k, v := range d.Get("context").(map[string]interface{}) {
			stringValue := fmt.Sprintf("%v", v)
			context[k] = &stringValue
		}

		version, err := c.ResolveVersion(table, name, 0)

		if err != nil {
			return diag.FromErr(err)
		}
		err = c.PutSecret(table, name, value, version, context)

		if err != nil {
			return diag.FromErr(err)
		}

		//Update the secret

		d.Set("last_updated", time.Now().Format(time.RFC850))

	}

	return resourceSecretRead(ctx, d, m)
}
