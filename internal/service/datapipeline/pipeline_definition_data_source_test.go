package datapipeline_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/datapipeline"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataPipelinePipelineDefinitionDataSource_basic(t *testing.T) {
	dataSourceName := "aws_datapipeline_pipeline_definition.test"
	resourceName := "aws_datapipeline_pipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataPipelinePipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDataPipelinePipelineDefinitionDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "pipeline_id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "pipeline_object.#", "1"),
				),
			},
		},
	})
}

func testAccDataPipelinePipelineDefinitionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_pipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id
  pipeline_object {
    id   = "Default"
    name = "Default"
    field {
      key          = "workerGroup"
      string_value = "workerGroup"
    }
  }
}

data "aws_datapipeline_pipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline_definition.test.pipeline_id
}
`, name)
}