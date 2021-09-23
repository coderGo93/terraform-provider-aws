package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directoryservice/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func DirectoryStage(conn *directoryservice.DirectoryService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DirectoryByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Stage), nil
	}
}

func ShareDirectoryStatus(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, sharedDirectoryID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ShareDirectoryByID(ctx, conn, directoryID, sharedDirectoryID)

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return output, directoryservice.ShareStatusDeleted, nil
		}

		return output, aws.StringValue(output.ShareStatus), nil
	}
}
