package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsCECostCategoryDefinition_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCECostCategoryDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCECostCategoryDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
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

func TestAccAwsCECostCategoryDefinition_complete(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCECostCategoryDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCECostCategoryDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccAwsCECostCategoryDefinitionOperandAndConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
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

func TestAccAwsCECostCategoryDefinition_splitCharge(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCECostCategoryDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCECostCategoryDefinitionSplitChargesConfig(rName, "PROPORTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccAwsCECostCategoryDefinitionSplitChargesConfig(rName, "EVEN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
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

func TestAccAwsCECostCategoryDefinition_disappears(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCECostCategoryDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCECostCategoryDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCECostCategoryDefinitionExists(resourceName, &output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCECostCategoryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsCECostCategoryDefinitionExists(resourceName string, output *costexplorer.CostCategory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).costexplorerconn
		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if err != nil {
			return fmt.Errorf("problem checking for CE Cost Category Definition existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("CE Cost Category Definition %q does not exist", rs.Primary.ID)
		}

		*output = *resp.CostCategory

		return nil
	}
}

func testAccCheckAwsCECostCategoryDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).costexplorerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_cost_category_definition" {
			continue
		}

		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking CE Cost Category Definition was destroyed: %w", err)
		}

		if resp != nil && resp.CostCategory != nil {
			return fmt.Errorf("CE Cost Category Definition %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsCECostCategoryDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category_definition" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "staging"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-stg"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "testing"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-dev"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
}
`, name)
}

func testAccAwsCECostCategoryDefinitionOperandAndConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category_definition" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-prod"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-stg"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-dev"]
          match_options = ["ENDS_WITH"]
        }
      }
    }
    type = "REGULAR"
  }
}
`, name)
}

func testAccAwsCECostCategoryDefinitionSplitChargesConfig(name, method string) string {
	return fmt.Sprintf(`

resource "aws_ce_cost_category_definition" "test1" {
  name         = "%[1]s-1"
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "staging"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-stg"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "testing"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-dev"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
}

resource "aws_ce_cost_category_definition" "test2" {
  name         = "%[1]s-2"
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-prod"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-stg"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-dev"]
          match_options = ["ENDS_WITH"]
        }
      }
    }
    type = "REGULAR"
  }
}

resource "aws_ce_cost_category_definition" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  split_charge_rule {
    method  = %[2]q
    source  = aws_ce_cost_category_definition.test1.id
    targets = [aws_ce_cost_category_definition.test2.id]
  }
}
`, name, method)
}
