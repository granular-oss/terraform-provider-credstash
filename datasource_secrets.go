package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecretRead,

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

func dataSourceSecretRead(d *schema.ResourceData, meta interface{}) error {
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

	log.Printf("[DEBUG] Getting secret for name=%s table=%s version=%v context=%+v", name, table, version, context)

	if version == 0 {
		value, err = client.GetHighestVersionSecret(table, name, context)

	} else {
		value, err = client.GetSecret(name, table, client.PaddedInt(version), context)

	}
	if err != nil {
		return err
	}
	d.Set("value", value.Secret)
	d.SetId(hash(value.Secret))

	return nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
