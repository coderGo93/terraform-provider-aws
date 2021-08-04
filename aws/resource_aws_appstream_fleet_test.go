package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)


func testAccAwsAppStreamFleet_basic(t *testing.T) {
	var providers []*schema.Provider
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"
	dataSourceAlternate := "data.aws_caller_identity.member"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigBasic(testAccDefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					testAccCheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
				),
			},
			{
				Config:            testAccAwsAppStreamFleetConfigBasic(testAccDefaultEmailAddress),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppStreamFleet_disappears(t *testing.T) {
	var providers []*schema.Provider
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigBasic(testAccDefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppstreamFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}


func testAccAwsAppStreamFleet_withTags(t *testing.T) {
	var providers []*schema.Provider
	var fleetOutput appstream.Fleet
	resourceName := "aws_appstream_fleet.fleet"
	dataSourceAlternate := "data.aws_caller_identity.member"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsAppStreamFleetDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamFleetConfigWithTags(testAccDefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamFleetExists(resourceName, &fleetOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					testAccCheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					testAccCheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
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

func testAccCheckAwsAppStreamFleetExists(resourceName string, appStreamFleet *appstream.Fleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn
		resp, err := conn.DescribeFleets(&appstream.DescribeFleetsInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.Fleets) == 0 {
			return fmt.Errorf("appstream fleet %q does not exist", rs.Primary.ID)
		}

		*appStreamFleet = *resp.Fleets[0]

		return nil
	}
}

func testAccCheckAwsAppStreamFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_fleet" {
			continue
		}

		resp, err := conn.DescribeFleets(&appstream.DescribeFleetsInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp != nil && len(resp.Fleets) > 0 {
			return fmt.Errorf("appstream fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamFleetConfigBasic(email string) string {
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test_fleet" {
  name       = %[1]q
  compute_capacity {
    desired_instances = 1
  }
  description                    = %[1]q
  disconnect_timeout             = 15
  display_name                   = %[1]q
  enable_default_internet_access = false
  fleet_type                     = %[2]q
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = %[3]q
  max_user_duration              = 600
  subnet_ids                     = ["subnet-06e9b13400c225127"]
  security_group_ids             = ["sg-0397cdfe509785903", "sg-0bd2dddff01dee52d"]
  tags = {
    TagName = "tag-value"
  }
}
`, email)
}

func testAccAwsAppStreamFleetConfigWithTags(email string) string {
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test_fleet" {
  name       = %[1]q
  compute_capacity {
    desired_instances = 1
  }
  description                    = %[1]q
  disconnect_timeout             = 15
  display_name                   = %[1]q
  enable_default_internet_access = false
  fleet_type                     = %[2]q
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = %[3]q
  max_user_duration              = 600
  subnet_ids                     = ["subnet-06e9b13400c225127"]
  security_group_ids             = ["sg-0397cdfe509785903", "sg-0bd2dddff01dee52d"]
  tags = {
    TagName = "tag-value"
  }
}
`, email)
}
