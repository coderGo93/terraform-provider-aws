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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSRoute53TrafficPolicyImportStateIdFunc(resourceName),
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
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53TrafficPolicy_complete(t *testing.T) {
	var output route53.TrafficPolicySummary
	resourceName := "aws_route53_traffic_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
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
				ImportStateIdFunc: testAccAWSRoute53TrafficPolicyImportStateIdFunc(resourceName),
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

		resp, err := getTrafficPolicyById(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("problem checking for traffic policy existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("traffic policy %q does not exist", rs.Primary.ID)
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

		resp, err := getTrafficPolicyById(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) || resp == nil {
			continue
		}

		if err != nil {
			return fmt.Errorf("error during check if traffic policy still exists, %#v", err)
		}
		if resp != nil {
			return fmt.Errorf("traffic Policy still exists")
		}
	}
	return nil
}

func testAccAWSRoute53TrafficPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["id"], rs.Primary.Attributes["version"]), nil
	}
}

func testAccRoute53TrafficPolicyConfig(name string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_route53_traffic_policy_document" "test" {
  record_type = "A"
  start_rule  = "site_switch"

  endpoint {
    id    = "my_elb"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_region.current.name}.elb.amazonaws.com"
  }
  endpoint {
    id     = "site_down_banner"
    type   = "s3-website"
    region = data.aws_region.current.name
    value  = "www.example.com"
  }

  rule {
    id   = "site_switch"
    type = "failover"

    primary {
      endpoint_reference = "my_elb"
    }
    secondary {
      endpoint_reference = "site_down_banner"
    }
  }
}

resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  document = data.aws_route53_traffic_policy_document.test.json
}
`, name)
}

func testAccRoute53TrafficPolicyConfigComplete(name, comment string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_route53_traffic_policy_document" "test" {
  record_type = "A"
  start_rule  = "site_switch"

  endpoint {
    id    = "my_elb"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_region.current.name}.elb.amazonaws.com"
  }
  endpoint {
    id     = "site_down_banner"
    type   = "s3-website"
    region = data.aws_region.current.name
    value  = "www.example.com"
  }

  rule {
    id   = "site_switch"
    type = "failover"

    primary {
      endpoint_reference = "my_elb"
    }
    secondary {
      endpoint_reference = "site_down_banner"
    }
  }
}

resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  comment  = %[2]q
  document = data.aws_route53_traffic_policy_document.test.json
}
`, name, comment)
}
