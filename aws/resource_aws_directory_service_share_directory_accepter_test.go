package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSDirectoryServiceShareDirectoryAccepter_basic(t *testing.T) {
	var output directoryservice.SharedDirectory
	var providers []*schema.Provider
	resourceName := "aws_directory_service_share_directory_accepter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		ErrorCheck:        testAccErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDirectoryServiceShareDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceShareDirectoryAccepterConfig(),
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

func testAccDirectoryServiceShareDirectoryAccepterConfig() string {
	return testAccDirectoryServiceDirectoryConfigBaseAlternate + `
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
  share_method = "HANDSHAKE"
  share_notes  = "Terraform testing"

  share_target {
    id   = data.aws_caller_identity.member.account_id
    type = "ACCOUNT"
  }
}

resource "aws_directory_service_share_directory_accepter" "test" {
  shared_directory_id = aws_directory_service_share_directory.test.id
}
`
}
