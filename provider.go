package main

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultAWSProfile = "default"

func Provider() *schema.Provider {
	// Provider that enables reading and creating of secrets with credstash
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"credstash_secret": dataSourceSecret(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"credstash_secret": resourceSecret(),
		},
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				Description: "The region where AWS operations will take place. Examples\n" +
					"are us-east-1, us-west-2, etc.",
			},
			"table": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The DynamoDB table where the secrets are stored.",
				Default:     "credential-store",
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     defaultAWSProfile,
				Description: "The profile that should be used to connect to AWS",
			},
		},
		ConfigureContextFunc: providerConfig,
	}
}

func providerConfig(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	region := d.Get("region").(string)
	table := d.Get("table").(string)
	profile := d.Get("profile").(string)

	var opts []func(*awsconfig.LoadOptions) error
	opts = append(opts, awsconfig.WithRegion(region))

	if profile != defaultAWSProfile {
		tflog.Debug(ctx, "Creating AWS config", map[string]interface{}{
			"profile": profile,
		})
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	tflog.Debug(ctx, "Creating Credstash Client", map[string]interface{}{
		"table": table,
	})

	dynamoClient := dynamodb.NewFromConfig(cfg)
	kmsClient := kms.NewFromConfig(cfg)

	return credstash.New(table, dynamoClient, kmsClient), nil
}
