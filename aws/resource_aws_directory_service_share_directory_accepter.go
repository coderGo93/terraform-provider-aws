package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDirectoryServiceShareDirectoryAccepter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsDirectoryServiceShareDirectoryAccepterCreate,
		ReadContext:   schema.NoopContext,
		DeleteContext: schema.NoopContext,
		Schema: map[string]*schema.Schema{
			"shared_directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDirectoryServiceShareDirectoryAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn
	input := &directoryservice.AcceptSharedDirectoryInput{
		SharedDirectoryId: aws.String(d.Get("shared_directory_id").(string)),
	}

	var err error
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = conn.AcceptSharedDirectoryWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryDoesNotExistException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.AcceptSharedDirectoryWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Directory Service Share Directory Accepter (%s): %w", d.Id(), err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}
