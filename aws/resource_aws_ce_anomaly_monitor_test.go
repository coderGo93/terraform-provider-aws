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

func TestAccAwsCEAnomalyMonitor_basic(t *testing.T) {
	var output costexplorer.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCEAnomalyMonitorDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCEAnomalyMonitorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCEAnomalyMonitorExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
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

func TestAccAwsCEAnomalyMonitor_disappears(t *testing.T) {
	var output costexplorer.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCEAnomalyMonitorDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCEAnomalyMonitorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCEAnomalyMonitorExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCEAnomalyMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsCEAnomalyMonitorExists(resourceName string, output *costexplorer.AnomalyMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).costexplorerconn
		resp, err := conn.GetAnomalyMonitorsWithContext(context.Background(), &costexplorer.GetAnomalyMonitorsInput{MonitorArnList: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return fmt.Errorf("problem checking for CE Anomaly Monitor existence: %w", err)
		}

		if resp == nil && len(resp.AnomalyMonitors) == 0 {
			return fmt.Errorf("CE Anomaly Monitor %q does not exist", rs.Primary.ID)
		}

		*output = *resp.AnomalyMonitors[0]

		return nil
	}
}

func testAccCheckAwsCEAnomalyMonitorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).costexplorerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_monitor" {
			continue
		}

		resp, err := conn.GetAnomalyMonitorsWithContext(context.Background(), &costexplorer.GetAnomalyMonitorsInput{MonitorArnList: []*string{aws.String(rs.Primary.ID)}})

		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking CE Anomaly Monitor was destroyed: %w", err)
		}

		if resp != nil && len(resp.AnomalyMonitors) > 0 {
			return fmt.Errorf("CE Anomaly Monitor %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsCEAnomalyMonitorConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  monitor_dimension = "SERVICE"
  monitor_name      = %[1]q
  monitor_type      = "DIMENSIONAL"
}
`, name)
}
