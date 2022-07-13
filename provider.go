package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/granular-oss/terraform-provider-credstash/credstash"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var _ *schema.Provider = provider()

const defaultAWSProfile = "default"

func provider() *schema.Provider {
	// rovider that enables reading and creating of secrets with credstash
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
		ConfigureFunc: providerConfig,
	}
}

func providerConfig(d *schema.ResourceData) (interface{}, error) {
	region := d.Get("region").(string)
	table := d.Get("table").(string)
	profile := d.Get("profile").(string)

	var sess *session.Session
	var err error
	if profile != defaultAWSProfile {
		log.Printf("[DEBUG] creating a session for profile: %s", profile)
		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            aws.Config{Region: aws.String(region)},
			Profile:           profile,
			SharedConfigState: session.SharedConfigEnable,
		})
	} else {
		sess, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	}
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] configured credstash for table %s", table)
	return credstash.New(table, sess), nil
}
