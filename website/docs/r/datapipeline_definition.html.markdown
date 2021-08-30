---
layout: "aws"
page_title: "AWS: aws_datapipeline_definition"
sidebar_current: "docs-aws-resource-datapipeline-definition"
description: |-
Provides a AWS DataPipeline Definition.
---

# Resource: aws_datapipeline_definition

Provides a Data Pipeline Definition resource.

## Example Usage

```hcl
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"
}

resource "aws_datapipeline_definition" "example" {
  pipeline_id = aws_datapipeline_pipeline.default.id
  pipeline_objects {
    id   = "Default"
    name = "Default"
    fields {
      key          = "workerGroup"
      string_value = "workerGroup"
    }
  }
  pipeline_objects {
    id   = "Schedule"
    name = "Schedule"
    fields {
      key          = "startDateTime"
      string_value = "2012-12-12T00:00:00"
    }
    fields {
      key          = "type"
      string_value = "Schedule"
    }
    fields {
      key          = "period"
      string_value = "1 hour"
    }
    fields {
      key          = "endDateTime"
      string_value = "2012-12-21T18:00:00"
    }
  }
  pipeline_objects {
    id   = "SayHello"
    name = "SayHello"
    fields {
      key          = "type"
      string_value = "ShellCommandActivity"
    }
    fields {
      key          = "command"
      string_value = "echo hello"
    }
    fields {
      key          = "parent"
      string_value = "Default"
    }
    fields {
      key          = "schedule"
      string_value = "Schedule"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `pipeline_id` - (Required) ID of the pipeline.
* `pipeline_objects` - (Required) Configuration block for the objects that define the pipeline. These objects overwrite the existing pipeline definition. See below

The following arguments are optional:

* `parameter_objects` - (Optional) Configuration block for the objects that define the pipeline. These objects overwrite the existing pipeline definition. See below
* `parameter_values` - (Optional) Configuration block for the objects that define the pipeline. These objects overwrite the existing pipeline definition. See below

### `pipeline_objects`

* `fields` - (Required) Configuration block for Key-value pairs that define the properties of the object. See below
* `id` - (Required) ID of the object.
* `name` - (Required) ARN of the storage connector.

### `fields`

* `key` - (Required) Field identifier.
* `ref_value` - (Optional) Field value, expressed as the identifier of another object
* `string_value` - (Optional) Field value, expressed as a String.

### `parameter_objects_`

* `attributes` - (Required) Configuration block for attributes of the parameter object. See below
* `id` - (Required) ID of the parameter object.

### `attributes`

* `key` - (Required) Field identifier.
* `string_value` - (Required) Field value, expressed as a String.

### `parameter_values`

* `id` - (Required) ID of the parameter value.
* `string_value` - (Required) Field value, expressed as a String.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique ID of the datapipeline definition.


## Import

`aws_datapipeline_pipeline` can be imported using the id, e.g.

```
$ terraform import aws_datapipeline_definition.example pipelineId
```
