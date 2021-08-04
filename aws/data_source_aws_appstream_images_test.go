package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccAwsAppStreamImageDataSourcesImage(t *testing.T) {
	dataSourceName := "data.aws_appstream_images.images"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcesAwsAppStreamConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageDataSourceID(dataSourceName),
				),
			},
		},
	})
}

const testAccDataSourcesAwsAppStreamConfig = `data "aws_appstream_images" "images" {}`
