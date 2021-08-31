package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53TrafficPolicy_basic(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := acctest.RandomWithPrefix("")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSRoute53TrafficPolicy_disappears(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53TrafficPolicy(), resourceName),
				),
			},
		},
	})
}

func TestAccAWSRoute53TrafficPolicy_complete(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := acctest.RandomWithPrefix("")
	comment := `comment`
	commentUpdated := `comment updated`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfigComplete(rName, comment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", comment),
				),
			},
			{
				Config: testAccRoute53TrafficPolicyConfigComplete(rName, commentUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", commentUpdated),
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

func testAccCheckAwsRoute53TrafficPolicyExists(resourceName string, trafficPolicy *route53.TrafficPolicySummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).r53conn

		idTraffic, version, err := decodeTrafficPolicyID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding Route53 Traffic Policy %s : %w", rs.Primary.Attributes["name"], err)
		}

		resp, err := getTrafficPolicyById(context.Background(), conn, idTraffic, version)

		if err != nil {
			return fmt.Errorf("problem checking for traffic policy existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("traffic policy %q does not exist", idTraffic)
		}

		*trafficPolicy = *resp

		return nil
	}
}

func testAccCheckRoute53TrafficPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}

		idTraffic, version, err := decodeTrafficPolicyID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding Route53 Traffic Policy %s : %w", rs.Primary.Attributes["name"], err)
		}

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
			continue
		}

		resp, err := getTrafficPolicyById(context.Background(), conn, idTraffic, version)

		if err != nil {
			return fmt.Errorf("error during check if traffic policy still exists, %#v", err)
		}
		if resp != nil {
			return fmt.Errorf("traffic Policy still exists")
		}
	}
	return nil
}

func testAccRoute53TrafficPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  document = <<-EOT
{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}
EOT
}
`, name)
}

func testAccRoute53TrafficPolicyConfigComplete(name, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  comment  = %[2]q
  document = <<-EOT
{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}
EOT
}
`, name, comment)
}
