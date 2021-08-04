package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsAppStreamImageDataSourceImage_nameBasic(t *testing.T) {
	dataSourceName := "data.aws_appstream_image.name_image"
	imageName := "AppStream-Graphics-Design-WinServer2012R2-07-19-2021"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAppStreamNameConfig(imageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageDataSourceID(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "name", imageName),
					resource.TestCheckResourceAttr(dataSourceName, "visibility", appstream.VisibilityTypePublic),
					testAccCheckResourceAttrRfc3339(dataSourceName, "created_time"),
				),
			},
		},
	})
}

func testAccAwsAppStreamImageDataSourceImage_arnBasic(t *testing.T) {
	dataSourceName := "data.aws_appstream_image.arn_image"
	arnName := "arn:aws:appstream:us-east-1::image/AppStream-Graphics-Design-WinServer2012R2-07-19-2021"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAppStreamArnConfig(arnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageDataSourceID(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "arn", arnName),
					resource.TestCheckResourceAttr(dataSourceName, "visibility", appstream.VisibilityTypePublic),
					testAccCheckResourceAttrRfc3339(dataSourceName, "created_time"),
				),
			},
		},
	})
}

func testAccCheckAwsAppStreamImageDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find appstream image data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("appstream image data source ID not set")
		}
		return nil
	}
}

func testAccDataSourceAwsAppStreamNameConfig(name string) string {
	return fmt.Sprintf(`
data "aws_appstream_image" "name_image" {
  name = %[1]q
}`, name)
}

func testAccDataSourceAwsAppStreamArnConfig(arn string) string {
	return fmt.Sprintf(`
data "aws_appstream_image" "arn_image" {
  arn  = %[1]q
}`, arn)
}
