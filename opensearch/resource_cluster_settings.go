package opensearch

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"terraform-provider-opensearch/api"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	persistentConfig      = "persistent"
	autoCreateIndexConfig = "auto_create_index"
)

func resourceClusterSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterSettingsCreate,
		ReadContext:   resourceClusterSettingsRead,
		UpdateContext: resourceClusterSettingsUpdate,
		DeleteContext: resourceClusterSettingsDelete,
		Schema: map[string]*schema.Schema{
			persistentConfig: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						autoCreateIndexConfig: {
							Type: schema.TypeString,
							Description: "Automatically create indexes when a request is received. " +
								"The operation automatically creates the index and applies any matching index templates. " +
								"If no mapping exists, the index operation creates a dynamic mapping. " +
								"Accepted values are: true, false or comma-separated list of patterns you want to allow," +
								" or each pattern prefixed with + or - to indicate whether it should be allowed or blocked",
							Optional:         true,
							ValidateDiagFunc: isValidAutoCreateIndexValue,
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

func resourceClusterSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*api.Client)

	persistentSettings, diags := extractPersistentSettings(d)
	if diags != nil {
		return diags
	}

	err := client.Update(persistentSettings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.PrefixedUniqueId("opensearch-"))

	resourceClusterSettingsRead(ctx, d, m)

	return diag.Diagnostics{}
}

func extractPersistentSettings(d *schema.ResourceData) (*api.PersistentSettings, diag.Diagnostics) {
	var settings = api.PersistentSettings{}

	if parentList, ok := d.Get(persistentConfig).([]interface{}); ok {
		if len(parentList) > 0 && parentList[0] != nil {
			if persistent, ok := parentList[0].(map[string]interface{}); ok {
				if autoIndexCreation, ok := persistent[autoCreateIndexConfig].(string); ok {
					settings.AutoCreateIndex = &autoIndexCreation
				} else {
					settings.AutoCreateIndex = nil
				}
			} else {
				return nil, diag.Errorf("unable to parse %v as map", persistentConfig)
			}
		} else {
			return nil, diag.Errorf(
				"unable to parse invalid %v data, got zero or more than one element",
				persistentConfig,
			)
		}
	} else {
		return nil, diag.Errorf("unable to find %v data", persistentConfig)
	}

	return &settings, nil
}

func resourceClusterSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*api.Client)

	settings, err := client.Fetch()
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(persistentConfig, flattenPersistentSettings(settings)); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func flattenPersistentSettings(settings *api.PersistentSettings) []interface{} {
	s := make(map[string]interface{})
	s[autoCreateIndexConfig] = settings.AutoCreateIndex

	return []interface{}{s}
}

func resourceClusterSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange(persistentConfig) {
		return resourceClusterSettingsCreate(ctx, d, m)
	} else {
		return resourceClusterSettingsRead(ctx, d, m)
	}

}

func resourceClusterSettingsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*api.Client)

	err := client.Update(&api.PersistentSettings{
		AutoCreateIndex: nil,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diag.Diagnostics{}
}

func isValidAutoCreateIndexValue(i interface{}, _ cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected to be string")
	}

	matched, err := regexp.MatchString(
		"^(true|false|([-+]?[a-z0-9][a-z0-9_-]*\\*?,?)+([-+]?[a-z0-9][a-z0-9_-]*\\*?))$",
		v,
	)
	if err != nil {
		return diag.FromErr(err)
	} else if !matched {
		return diag.Errorf("expected to follow value format: true, false or comma-separated list but got %v", v)
	}

	return diag.Diagnostics{}
}
