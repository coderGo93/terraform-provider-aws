package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsAppStreamImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsAppStreamImageRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(appstream.VisibilityType_Values(), false),
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
	}
}

func dataSourceAwsAppStreamImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	params := &appstream.DescribeImagesInput{}

	if v, ok := d.GetOk("name"); ok {
		params.Names = []*string{aws.String(v.(string))}
	}
	if v, ok := d.GetOk("arn"); ok {
		params.Arns = []*string{aws.String(v.(string))}
	}
	if v, ok := d.GetOk("type"); ok {
		params.Type = aws.String(v.(string))
	}

	resp, err := conn.DescribeImages(params)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(resp.Images) == 0 {
		return diag.FromErr(fmt.Errorf("your query returned no results. please change your search criteria and try again"))
	}

	if len(resp.Images) > 1 {
		return diag.FromErr(fmt.Errorf("your query returned more than one result. please change your search criteria and try again"))

	}

	image := resp.Images[0]

	d.Set("applications", flattenAppStreamApplications(image.Applications))
	d.Set("app_stream_agent_version", image.AppstreamAgentVersion)
	d.Set("base_image_arn", image.BaseImageArn)
	d.Set("created_time", image.CreatedTime.Format(time.RFC3339))
	d.Set("description", image.Description)
	d.Set("display_name", image.DisplayName)
	d.Set("image_builder_name", image.ImageBuilderName)
	d.Set("image_builder_supported", image.ImageBuilderSupported)
	d.Set("platform", image.Platform)
	d.Set("public_base_image_released_date", image.PublicBaseImageReleasedDate.Format(time.RFC3339))
	d.Set("state", image.State)
	d.Set("visibility", image.Visibility)

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}

func flattenAppStreamApplications(applications []*appstream.Application) []interface{} {
	if applications == nil {
		return nil
	}

	var listApplications []interface{}

	for _, application := range applications {
		app := map[string]interface{}{}

		app["display_name"] = aws.StringValue(application.DisplayName)
		app["enabled"] = aws.BoolValue(application.Enabled)
		app["icon_url"] = aws.StringValue(application.IconURL)
		app["launch_parameters"] = aws.StringValue(application.LaunchParameters)
		app["launch_path"] = aws.StringValue(application.LaunchPath)
		app["metadata"] = aws.StringValueMap(application.Metadata)
		app["name"] = aws.StringValue(application.Name)

		listApplications = append(listApplications, app)
	}

	return listApplications
}
