package finder

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func DirectoryByID(conn *directoryservice.DirectoryService, id string) (*directoryservice.DirectoryDescription, error) {
	input := &directoryservice.DescribeDirectoriesInput{
		DirectoryIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeDirectories(input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectoryDescriptions) == 0 || output.DirectoryDescriptions[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	directory := output.DirectoryDescriptions[0]

	if stage := aws.StringValue(directory.Stage); stage == directoryservice.DirectoryStageDeleted {
		return nil, &resource.NotFoundError{
			Message:     stage,
			LastRequest: input,
		}
	}

	return directory, nil
}

func ShareDirectoryByID(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, sharedDirectoryID string) (*directoryservice.SharedDirectory, error) {
	input := &directoryservice.DescribeSharedDirectoriesInput{
		SharedDirectoryIds: aws.StringSlice([]string{sharedDirectoryID}),
		OwnerDirectoryId:   aws.String(directoryID),
	}

	output, err := conn.DescribeSharedDirectoriesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	var directory *directoryservice.SharedDirectory
	if len(output.SharedDirectories) > 1 {
		return nil, fmt.Errorf("[ERROR] got more than one shared directory with the shared id: %s and directory id: %s", sharedDirectoryID, directoryID)
	}

	if len(output.SharedDirectories) == 1 {
		directory = output.SharedDirectories[0]
	}

	return directory, nil
}
