package aws

import (
	"context"
	"fmt"
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
			SharedDirectoryIds: []*string{aws.String(rs.Primary.ID)},
		}
		out, err := conn.DescribeSharedDirectoriesWithContext(context.Background(), input)

		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
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
			SharedDirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(out.SharedDirectories) < 1 {
			return fmt.Errorf("No DS directory found")
		}

		if *out.SharedDirectories[0].SharedDirectoryId != rs.Primary.ID {
			return fmt.Errorf("DS share directory ID mismatch - existing: %q, state: %q",
				*out.SharedDirectories[0].SharedDirectoryId, rs.Primary.ID)
		}

		*output = *out.SharedDirectories[0]

		return nil
	}
}

var testAccDirectoryServiceDirectoryConfigBaseAlternate = testAccAlternateAccountProviderConfig() + `
data "aws_caller_identity" "admin" {
  provider = "awsalternate"
}

data "aws_availability_zones" "available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-directory-service-directory-tags"
  }

  depends_on = [data.aws_caller_identity.admin]
}

resource "aws_subnet" "test1" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}
`

func testAccDirectoryServiceShareDirectoryConfig() string {
	return  testAccDirectoryServiceDirectoryConfigBaseAlternate + `
data "aws_caller_identity" "member" {}

resource "aws_directory_service_directory" "test" {
  provider = "awsalternate"

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
  provider = "awsalternate"

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
