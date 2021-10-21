package opensearch

import (
	"context"
	"terraform-provider-opensearch/api"

	"terraform-provider-opensearch/signing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	endpointConfig          = "endpoint"
	awsRequestSigningConfig = "aws_request_signing"
	awsRegionConfig         = "region"
	awsRoleConfig           = "role"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			endpointConfig: {
				Type:        schema.TypeString,
				Description: "Server's URL",
				Required:    true,
			},
			awsRequestSigningConfig: {
				Type: schema.TypeList,
				Description: "Sign requests according to AWS requirements. " +
					"It requires AWS credentials to be accessible through default provider chain.",
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						awsRegionConfig: {
							Type:        schema.TypeString,
							Description: "AWS Region",
							Required:    true,
						},
						awsRoleConfig: {
							Type:        schema.TypeString,
							Description: "ARN of the role to assume when getting AWS credentials",
							Optional:    true,
						},
					},
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"opensearch_cluster_settings": resourceClusterSettings(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	signer, diags := createSigner(d)
	if diags.HasError() {
		return nil, diags
	}

	endpoint, diags := getEndpoint(d)
	if diags.HasError() {
		return nil, diags
	}

	return api.NewClient(endpoint, signer), nil
}

func createSigner(d *schema.ResourceData) (signing.Signer, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := d.GetOk(awsRequestSigningConfig); ok {
		if len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			awsConfig := aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
			}

			signingConfig := v.([]interface{})[0].(map[string]interface{})

			region, ok := signingConfig[awsRegionConfig].(string)
			if !ok || region == "" {
				return nil, diag.Errorf("unable to parse config %s", awsRegionConfig)
			} else {
				awsConfig.Region = &region
			}

			sess, err := session.NewSession(&awsConfig)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			creds := sess.Config.Credentials
			role, ok := signingConfig[awsRoleConfig].(string)
			if ok && role != "" {
				creds = stscreds.NewCredentials(sess, role)
			}

			return signing.NewAwsSigner(region, creds), diags
		} else {
			return nil, diag.Errorf("unable to parse config %s", awsRequestSigningConfig)
		}
	} else {
		return signing.NewNoOpSigner(), diags
	}
}

func getEndpoint(d *schema.ResourceData) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if i, ok := d.GetOk(endpointConfig); ok {
		v, ok := i.(string)
		if ok {
			return v, diags
		} else {
			return "", diag.Errorf("expected to be string")
		}
	} else {
		return "", diag.Errorf("endpoint must be configured")
	}
}
