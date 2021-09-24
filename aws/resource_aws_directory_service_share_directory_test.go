package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDirectoryServiceShareDirectory_basic(t *testing.T) {
	var output directoryservice.SharedDirectory
	var providers []*schema.Provider
	resourceName := "aws_directory_service_share_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		ErrorCheck:        testAccErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDirectoryServiceShareDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceShareDirectoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceShareDirectoryExists(resourceName, &output),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDirectoryServiceShareDirectory_disappears(t *testing.T) {
	var output directoryservice.SharedDirectory
	var providers []*schema.Provider
	resourceName := "aws_directory_service_share_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		ErrorCheck:        testAccErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDirectoryServiceShareDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceShareDirectoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceShareDirectoryExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDirectoryServiceShareDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDirectoryServiceShareDirectoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_share_directory" {
			continue
		}

		input := &directoryservice.DescribeSharedDirectoriesInput{
			SharedDirectoryIds: []*string{aws.String(rs.Primary.Attributes["shared_directory_id"])},
			OwnerDirectoryId:   aws.String(rs.Primary.Attributes["directory_id"]),
		}
		out, err := conn.DescribeSharedDirectoriesWithContext(context.Background(), input)
		log.Printf("[DEBIG] testAccCheckDirectoryServiceShareDirectoryDestroy invoked")
		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) ||
			tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryNotSharedException) {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil && len(out.SharedDirectories) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Share Directory to be gone, but was still found")
		}
	}

	return nil
}

func testAccCheckServiceShareDirectoryExists(name string, output *directoryservice.SharedDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dsconn

		out, err := conn.DescribeSharedDirectoriesWithContext(context.Background(), &directoryservice.DescribeSharedDirectoriesInput{
			SharedDirectoryIds: []*string{aws.String(rs.Primary.Attributes["shared_directory_id"])},
			OwnerDirectoryId:   aws.String(rs.Primary.Attributes["directory_id"]),
		})
		log.Printf("[DEBIG] testAccCheckServiceShareDirectoryExists invoked")

		if err != nil {
			return err
		}

		if out != nil && len(out.SharedDirectories) == 0 {
			return fmt.Errorf("No DS share directory found")
		}

		*output = *out.SharedDirectories[0]

		return nil
	}
}

var testAccDirectoryServiceShareDirectoryConfigBase = testAccAlternateAccountProviderConfig() + `
data "aws_caller_identity" "admin" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-directory-service-directory-tags"
  }

  depends_on = [data.aws_caller_identity.admin]
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}
`

func testAccDirectoryServiceShareDirectoryConfig() string {
	return testAccDirectoryServiceShareDirectoryConfigBase + `
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
  depends_on = [data.aws_caller_identity.admin]
}

resource "aws_directory_service_share_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
  share_method = "ORGANIZATIONS"
  share_notes  = "Terraform testing"

  share_target {
    id   = data.aws_caller_identity.member.account_id
    type = "ACCOUNT"
  }
}
`
}
