package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsCEAnomalySubscription_basic(t *testing.T) {
	var output costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCEAnomalySubscriptionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCEAnomalySubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCEAnomalySubscriptionExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "subscription_name", rName),
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

func TestAccAwsCEAnomalySubscription_disappears(t *testing.T) {
	var output costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCEAnomalySubscriptionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCEAnomalySubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCEAnomalySubscriptionExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCEAnomalySubscription(), resourceName),
				),
			},
		},
	})
}

func testAccCheckAwsCEAnomalySubscriptionExists(resourceName string, output *costexplorer.AnomalySubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).costexplorerconn
		resp, err := conn.GetAnomalySubscriptionsWithContext(context.Background(), &costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return fmt.Errorf("problem checking for CE Anomaly Subscription existence: %w", err)
		}

		if resp == nil && len(resp.AnomalySubscriptions) == 0 {
			return fmt.Errorf("CE Anomaly Subscription %q does not exist", rs.Primary.ID)
		}

		*output = *resp.AnomalySubscriptions[0]

		return nil
	}
}

func testAccCheckAwsCEAnomalySubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).costexplorerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_subscription" {
			continue
		}

		resp, err := conn.GetAnomalySubscriptionsWithContext(context.Background(), &costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: []*string{aws.String(rs.Primary.ID)}})

		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, costexplorer.ErrCodeUnknownSubscriptionException, "No anomaly subscription") {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking CE Anomaly Subscription was destroyed: %w", err)
		}

		if resp != nil && len(resp.AnomalySubscriptions) > 0 {
			return fmt.Errorf("CE Anomaly Subscription %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsCEAnomalySubscriptionConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  monitor_dimension = "SERVICE"
  monitor_name      = %[1]q
  monitor_type      = "DIMENSIONAL"
}

resource "aws_ce_anomaly_subscription" "test" {
  subscription_name = %[1]q
  threshold         = 0
  frequency         = "DAILY"
  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.id,
  ]
  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
    status  = "CONFIRMED"
  }
}
`, name)
}
