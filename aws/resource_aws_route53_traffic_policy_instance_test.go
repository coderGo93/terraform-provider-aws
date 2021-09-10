package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53TrafficPolicyInstance_basic(t *testing.T) {
	var output route53.GetTrafficPolicyInstanceOutput
	resourceName := "aws_route53_traffic_policy_instance.test"

	zoneName := testAccRandomDomainName()
	rName := fmt.Sprintf("%s_%s", acctest.RandomWithPrefix("tf-acc-test"), zoneName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyInstanceDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyInstanceConfig(zoneName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyInstanceExists(resourceName, &output),
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

func TestAccAWSRoute53TrafficPolicyInstance_disappears(t *testing.T) {
	var output route53.GetTrafficPolicyInstanceOutput
	resourceName := "aws_route53_traffic_policy_instance.test"

	zoneName := testAccRandomDomainName()
	rName := fmt.Sprintf("%s_%s", acctest.RandomWithPrefix("tf-acc-test"), zoneName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyInstanceDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyInstanceConfig(zoneName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyInstanceExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53TrafficPolicyInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53TrafficPolicyInstance_complete(t *testing.T) {
	var output route53.GetTrafficPolicyInstanceOutput
	resourceName := "aws_route53_traffic_policy_instance.test"

	zoneName := testAccRandomDomainName()
	rName := fmt.Sprintf("%s_%s", acctest.RandomWithPrefix("tf-acc-test"), zoneName)
	rNameUpdated := fmt.Sprintf("%s_%s", acctest.RandomWithPrefix("tf-acc-test"), zoneName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyInstanceDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyInstanceConfig(zoneName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyInstanceExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccRoute53TrafficPolicyInstanceConfig(zoneName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53TrafficPolicyInstanceExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func testAccCheckAwsRoute53TrafficPolicyInstanceExists(resourceName string, output *route53.GetTrafficPolicyInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).r53conn

		input := &route53.GetTrafficPolicyInstanceInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetTrafficPolicyInstanceWithContext(context.Background(), input)

		if err != nil {
			return fmt.Errorf("problem checking for traffic policy instance existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("traffic policy instance %q does not exist", rs.Primary.ID)
		}

		output = resp

		return nil
	}
}

func testAccCheckRoute53TrafficPolicyInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy_instance" {
			continue
		}

		input := &route53.GetTrafficPolicyInstanceInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetTrafficPolicyInstanceWithContext(context.Background(), input)
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Route53 Traffic Policy Instance %s : %w", rs.Primary.Attributes["name"], err)
		}

		if err != nil {
			return fmt.Errorf("error during check if traffic policy instance still exists, %#v", err)
		}
		if resp != nil {
			return fmt.Errorf("traffic Policy instance still exists")
		}
	}
	return nil
}

func testAccRoute53TrafficPolicyInstanceConfig(zoneName, instanceName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

resource "aws_route53_traffic_policy" "test" {
  name     = aws_route53_zone.test.name
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

resource "aws_route53_traffic_policy_instance" "test" {
  hosted_zone_id         = aws_route53_zone.test.zone_id
  name                   = %[2]q
  traffic_policy_id      = aws_route53_traffic_policy.test.id
  traffic_policy_version = aws_route53_traffic_policy.test.version
  ttl                    = 360
}
`, zoneName, instanceName)
}
