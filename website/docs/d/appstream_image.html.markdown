---
subcategory: "appstream"
layout: "aws"
page_title: "AWS: aws_appstream_image"
description: |-
  Get information on an Amazon AppStream Image.
---

# Data Source: aws_appstream_image

## Example Usage

```terraform
data "aws_appstream_image" "name_image" {
  name = "IMAGE NAME"
  type = "PRIVATE"
}
```

## Argument Reference

* `name` - (Optional) The name of the public or private image to describe.

* `arn` - (Optional) The ARN of the public, private, and shared images to describe.

* `type` - (Optional) The type of image (public, private, or shared) to describe. .

~> **NOTE:** If more or less than a single match is returned by the search,
Terraform will fail.

## Attributes Reference

* `applications` - The applications associated with the image.
    * `display_name` - The application name to display.
    * `enabled` - If there is a problem, the application can be disabled after image creation.
    * `icon_url` - The URL for the application icon. This URL might be time-limited.
    * `launch_parameters` - The arguments that are passed to the application at launch.
    * `launch_path` - The path to the application executable in the instance.
    * `metadata` - Additional attributes that describe the application.
    * `name` - The name of the application.
* `appstream_agent_version` - The version of the AppStream 2.0 agent to use for instances that are launched from this image.
* `base_image_arn` - The ARN of the image from which this image was created.
* `created_time` - The time the image was created.
* `description` - The description to display.
* `display_name` - The image name to display.
* `image_builder_name` - The name of the image builder that was used to create the private image. If the image is shared, this value is null.
* `image_builder_supported` - Indicates whether an image builder can be launched from this image.
* `platform` - The operating system platform of the image.
* `public_base_image_released_date` - The release date of the public base image.
* `state` - The state of the image.
* `visibility` - Indicates whether the image is public or private.

[1]: https://docs.aws.amazon.com/appstream2/latest/APIReference/API_DescribeImages.html
