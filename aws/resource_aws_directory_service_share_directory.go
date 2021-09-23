package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directoryservice/waiter"
)

func resourceAwsDirectoryServiceShareDirectory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsDirectoryServiceShareDirectoryCreate,
		ReadContext:   resourceAwsDirectoryServiceShareDirectoryRead,
		DeleteContext: resourceAwsDirectoryServiceShareDirectoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"created_date_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"last_updated_date_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.ShareMethod_Values(), false),
			},
			"share_notes": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"shared_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(directoryservice.TargetType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsDirectoryServiceShareDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn

	directoryID := d.Get("directory_id").(string)
	input := &directoryservice.ShareDirectoryInput{
		DirectoryId: aws.String(directoryID),
		ShareMethod: aws.String(d.Get("share_method").(string)),
		ShareTarget: expandShareDirectories(d.Get("share_target").([]interface{})),
	}

	if v, ok := d.GetOk("share_notes"); ok {
		input.ShareNotes = aws.String(v.(string))
	}

	var err error
	var output *directoryservice.ShareDirectoryOutput
	err = resource.RetryContext(ctx, waiter.ShareDirectoryOperationTimeout, func() *resource.RetryError {
		output, err = conn.ShareDirectoryWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryDoesNotExistException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.ShareDirectoryWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Directory Service Share Directory (%s): %w", d.Id(), err))
	}

	_, err = waiter.ShareDirectoryShared(ctx, conn, directoryID, aws.StringValue(output.SharedDirectoryId))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Directory Service Share Directory (%s) to be shared: %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.SharedDirectoryId))

	return resourceAwsDirectoryServiceShareDirectoryRead(ctx, d, meta)
}

func resourceAwsDirectoryServiceShareDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn

	resp, err := conn.DescribeSharedDirectoriesWithContext(ctx, &directoryservice.DescribeSharedDirectoriesInput{
		SharedDirectoryIds: []*string{aws.String(d.Id())},
		OwnerDirectoryId:   aws.String(d.Get("directory_id").(string)),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryNotSharedException) {
		log.Printf("[WARN] Directory Service Share Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Directory Service Share Directory (%s): %w", d.Id(), err))
	}
	for _, v := range resp.SharedDirectories {

		d.Set("created_date_time", aws.TimeValue(v.CreatedDateTime).Format(time.RFC3339))
		d.Set("last_updated_date_time", aws.TimeValue(v.LastUpdatedDateTime).Format(time.RFC3339))
		d.Set("owner_account_id", v.SharedAccountId)
		d.Set("owner_directory_id", v.SharedDirectoryId)
		d.Set("shared_account_id", v.SharedAccountId)
		d.Set("share_method", v.ShareMethod)
		d.Set("share_notes", v.ShareNotes)
		d.Set("share_status", v.ShareStatus)

	}

	return nil
}

func resourceAwsDirectoryServiceShareDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn

	directoryID := d.Get("directory_id").(string)
	input := &directoryservice.UnshareDirectoryInput{
		DirectoryId:   aws.String(directoryID),
		UnshareTarget: expandUnShareDirectories(d.Get("share_target").([]interface{})),
	}

	_, err := conn.UnshareDirectoryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) ||
		tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryNotSharedException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Directory Service Share Directory (%s): %w", d.Id(), err))
	}

	_, err = waiter.ShareDirectoryDeleted(ctx, conn, directoryID, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Directory Service Share Directory (%s) to be deleted: %w", d.Id(), err))
	}

	return nil
}

func expandShareDirectory(tfMap map[string]interface{}) *directoryservice.ShareTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.ShareTarget{
		Id:   aws.String(tfMap["id"].(string)),
		Type: aws.String(tfMap["type"].(string)),
	}

	return apiObject
}

func expandShareDirectories(tfList []interface{}) *directoryservice.ShareTarget {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject *directoryservice.ShareTarget

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject = expandShareDirectory(tfMap)
	}

	return apiObject
}

func expandUnShareDirectory(tfMap map[string]interface{}) *directoryservice.UnshareTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.UnshareTarget{
		Id:   aws.String(tfMap["id"].(string)),
		Type: aws.String(tfMap["type"].(string)),
	}

	return apiObject
}

func expandUnShareDirectories(tfList []interface{}) *directoryservice.UnshareTarget {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject *directoryservice.UnshareTarget

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject = expandUnShareDirectory(tfMap)
	}

	return apiObject
}
