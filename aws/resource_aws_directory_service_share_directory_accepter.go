package aws

import (
	"context"
	"fmt"
	"log"

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
		ReadContext:   resourceAwsDirectoryServiceShareDirectoryAccepterRead,
		DeleteContext: resourceAwsDirectoryServiceShareDirectoryAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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

	d.SetId(d.Get("shared_directory_id").(string))

	return resourceAwsDirectoryServiceShareDirectoryAccepterRead(ctx, d, meta)
}

func resourceAwsDirectoryServiceShareDirectoryAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn
	_, err := conn.DescribeSharedDirectoriesWithContext(ctx, &directoryservice.DescribeSharedDirectoriesInput{SharedDirectoryIds: []*string{aws.String(d.Id())}})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryNotSharedException) ||
		tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryNotSharedException) {
		log.Printf("[WARN] Directory Service Share Directory Accepter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Directory Service Share Directory Accepter (%s): %w", d.Id(), err))
	}

	return nil
}

func resourceAwsDirectoryServiceShareDirectoryAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dsconn

	input := &directoryservice.RejectSharedDirectoryInput{
		SharedDirectoryId: aws.String(d.Id()),
	}

	_, err := conn.RejectSharedDirectoryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Directory Service Share Directory Accepter (%s): %w", d.Id(), err))
	}

	return nil
}
