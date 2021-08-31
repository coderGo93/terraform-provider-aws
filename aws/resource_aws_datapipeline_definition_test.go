package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsDataPipelineDefinition_basic(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDataPipelineDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataPipelineDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDataPipelineDefinitionExists(resourceName, &pipelineOutput),
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

func TestAccAwsDataPipelineDefinition_disappears(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDataPipelineDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataPipelineDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDataPipelineDefinitionExists(resourceName, &pipelineOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDataPipelineDefinition(), resourceName),
				),
			},
		},
	})
}

func TestAccAwsDataPipelineDefinition_complete(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDataPipelineDefinitionDestroy,
		ErrorCheck:        testAccErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataPipelineDefinitionConfigComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDataPipelineDefinitionExists(resourceName, &pipelineOutput),
				),
			},
			{
				Config: testAccAwsDataPipelineDefinitionConfigCompleteUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDataPipelineDefinitionExists(resourceName, &pipelineOutput),
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

func testAccCheckAwsDataPipelineDefinitionExists(resourceName string, datapipelineOutput *datapipeline.GetPipelineDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn
		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})
		if err != nil {
			return fmt.Errorf("problem checking for DataPipeline Definition existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("datapipeline definition %q does not exist", rs.Primary.ID)
		}

		*datapipelineOutput = *resp

		return nil
	}
}

func testAccCheckAwsDataPipelineDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_definition" {
			continue
		}

		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) ||
			tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking DataPipeline Definition was destroyed: %w", err)
		}

		if resp != nil {
			return fmt.Errorf("datapipeline definition %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsDataPipelineDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"
}

resource "aws_datapipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id
  pipeline_objects {
    id   = "Default"
    name = "Default"
    fields {
      key          = "workerGroup"
      string_value = "workerGroup"
    }
  }
}
`, name)
}

func testAccAwsDataPipelineDefinitionConfigComplete(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "datapipeline.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id

  parameter_objects {
    id = "myAWSCLICmd"

    attributes {
      key          = "description"
      string_value = "AWS CLI command"
    }
    attributes {
      key          = "type"
      string_value = "String"
    }
    attributes {
      key          = "watermark"
      string_value = "aws [options] <command> <subcommand> [parameters]"
    }
  }

  parameter_values {
    id           = "myAWSCLICmd"
    string_value = "aws sts get-caller-identity"
  }

  pipeline_objects {
    id   = "CliActivity"
    name = "CliActivity"

    fields {
      key          = "command"
      string_value = "(sudo yum -y update aws-cli) && (#{myAWSCLICmd})"
    }
    fields {
      key       = "runsOn"
      ref_value = "Ec2Instance"
    }
    fields {
      key          = "type"
      string_value = "ShellCommandActivity"
    }
  }
  pipeline_objects {
    id   = "Default"
    name = "Default"

    fields {
      key          = "failureAndRerunMode"
      string_value = "CASCADE"
    }
    fields {
      key          = "resourceRole"
      string_value = aws_iam_role.test.name
    }
    fields {
      key          = "role"
      string_value = aws_iam_role.test.name
    }
    fields {
      key       = "schedule"
      ref_value = "DefaultSchedule"
    }
    fields {
      key          = "scheduleType"
      string_value = "cron"
    }
  }
  pipeline_objects {
    id   = "Ec2Instance"
    name = "Ec2Instance"

    fields {
      key          = "instanceType"
      string_value = "t1.micro"
    }
    fields {
      key          = "terminateAfter"
      string_value = "50 minutes"
    }
    fields {
      key          = "type"
      string_value = "Ec2Resource"
    }
  }
  pipeline_objects {
    id   = "DefaultSchedule"
    name = "Every 2 day"

    fields {
      key          = "period"
      string_value = "1 days"
    }
    fields {
      key          = "startAt"
      string_value = "FIRST_ACTIVATION_DATE_TIME"
    }
    fields {
      key          = "type"
      string_value = "Schedule"
    }
  }
}
`, name)
}

func testAccAwsDataPipelineDefinitionConfigCompleteUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "datapipeline.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id

  parameter_objects {
    id = "myAWSCLICmd"

    attributes {
      key          = "description"
      string_value = "AWS CLI command"
    }
    attributes {
      key          = "type"
      string_value = "String"
    }
    attributes {
      key          = "watermark"
      string_value = "aws [options] <command> <subcommand> [parameters]"
    }
  }

  parameter_values {
    id           = "myAWSCLICmd"
    string_value = "aws sts get-caller-identity"
  }

  pipeline_objects {
    id   = "CliActivity"
    name = "CliActivity"

    fields {
      key          = "command"
      string_value = "(sudo yum -y update aws-cli) && (#{myAWSCLICmd})"
    }
    fields {
      key       = "runsOn"
      ref_value = "Ec2Instance"
    }
    fields {
      key          = "type"
      string_value = "ShellCommandActivity"
    }
  }
  pipeline_objects {
    id   = "Default"
    name = "Default"

    fields {
      key          = "failureAndRerunMode"
      string_value = "CASCADE"
    }
    fields {
      key          = "resourceRole"
      string_value = aws_iam_role.test.name
    }
    fields {
      key          = "role"
      string_value = aws_iam_role.test.name
    }
    fields {
      key       = "schedule"
      ref_value = "DefaultSchedule"
    }
    fields {
      key          = "scheduleType"
      string_value = "cron"
    }
  }
  pipeline_objects {
    id   = "Ec2Instance"
    name = "Ec2Instance"

    fields {
      key          = "instanceType"
      string_value = "t1.micro"
    }
    fields {
      key          = "terminateAfter"
      string_value = "50 minutes"
    }
    fields {
      key          = "type"
      string_value = "Ec2Resource"
    }
  }
  pipeline_objects {
    id   = "DefaultSchedule"
    name = "Every 2 day"

    fields {
      key          = "period"
      string_value = "1 days"
    }
    fields {
      key          = "startAt"
      string_value = "FIRST_ACTIVATION_DATE_TIME"
    }
    fields {
      key          = "type"
      string_value = "Schedule"
    }
  }
}
`, name)
}
