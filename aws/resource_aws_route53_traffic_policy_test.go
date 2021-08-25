package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53TrafficPolicy_basic(t *testing.T) {
	resourceName := "aws_route53_traffic_policy.test"
	rName := acctest.RandomWithPrefix("")
	comment := `{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53TrafficPolicyDestroy,
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfig(rName, comment),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", comment),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoute53TrafficPolicyDestroy(s *terraform.State) error {
	return testAccCheckRoute53TrafficPolicyDestroyWithProvider(s, testAccProvider)
}

func testAccCheckRoute53TrafficPolicyDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}
		tp, err := getTrafficPolicyById(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error during check if traffic policy still exists, %#v", err)
		}
		if tp != nil {
			return fmt.Errorf("traffic Policy still exists")
		}
	}
	return nil
}

func testAccRoute53TrafficPolicyConfig(name, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  comment  = "comment"
  document = <<-EOT
%[2]s
EOT
}
`, name, comment)
}
