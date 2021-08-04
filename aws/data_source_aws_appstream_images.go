package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsAppStreamImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsAppStreamImagesRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(appstream.VisibilityType_Values(), false),
			},
			"results": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"applications": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"icon_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"launch_parameters": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"launch_path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"metadata": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"app_stream_agent_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_image_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"image_builder_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"image_builder_supported": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"platform": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_base_image_released_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"visibility": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAppStreamImagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	params := &appstream.DescribeImagesInput{}

	if v, ok := d.GetOk("names"); ok {
		params.Names = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("arns"); ok {
		params.Arns = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("type"); ok {
		params.Type = aws.String(v.(string))
	}

	resp, err := conn.DescribeImages(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("results", flattenAppStreamImages(resp.Images))

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}

func flattenAppStreamImages(images []*appstream.Image) []map[string]interface{} {

	results := make([]map[string]interface{}, 0)

	for _, image := range images {
		result := map[string]interface{}{
			"name":                            image.Name,
			"arn":                             image.Arn,
			"applications":                    flattenAppStreamApplications(image.Applications),
			"app_stream_agent_version":        image.AppstreamAgentVersion,
			"base_image_arn":                  image.BaseImageArn,
			"created_time":                    image.CreatedTime.Format(time.RFC3339),
			"description":                     image.Description,
			"display_name":                    image.DisplayName,
			"image_builder_name":              image.ImageBuilderName,
			"image_builder_supported":         image.ImageBuilderSupported,
			"platform":                        image.Platform,
			"public_base_image_released_date": image.PublicBaseImageReleasedDate.Format(time.RFC3339),
			"state":                           image.State,
			"visibility":                      image.Visibility,
		}
		results = append(results, result)
	}

	return results
}
