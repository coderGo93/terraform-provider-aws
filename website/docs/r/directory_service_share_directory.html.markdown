---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_share_directory"
description: |-
  Shares a directory in AWS Directory Service.
---

# Resource: aws_directory_service_share_directory

Shares the Managed Microsoft directory in AWS Directory Service.

## Example Usage

```terraform
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.main.id
  availability_zone = "us-west-2a"
  cidr_block        = "10.0.1.0/24"
}

resource "aws_subnet" "bar" {
  vpc_id            = aws_vpc.main.id
  availability_zone = "us-west-2b"
  cidr_block        = "10.0.2.0/24"
}

data "aws_caller_identity" "admin" {
  provider = "awsalternate"
}

data "aws_caller_identity" "member" {}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]
  }
  depends_on = [aws_caller_identity.admin]
}

resource "aws_directory_service_share_directory" "example" {
  provider = "awsalternate"
  directory_id = aws_directory_service_directory.test.id
  share_method = "ORGANIZATIONS"

  share_target {
    id   = data.aws_caller_identity.member.account_id
    type = "ACCOUNT"
  }
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) Identifier of the AWS Managed Microsoft AD directory that you want to share with other AWS accounts.
* `share_method` - (Required) The method used when sharing a directory to determine whether the directory should be shared within your AWS organization (`ORGANIZATIONS`) or with any AWS account by sending a directory sharing request (`HANDSHAKE`). Valid values `ORGANIZATIONS` , `HANDSHAKE`
* `share_target` - (Required) Configuration block for the directory consumer account with whom the directory is to be shared. See below.

The following arguments are optional:

* `share_notes` - (Optional) A directory share request that is sent by the directory owner to the directory consumer.

### share_target Reference

* `id` - (Required) Identifier of the directory consumer account.
* `type` - (Required) Type of identifier to be used in the `id` field. Valid values `ACCOUNT`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the shared directory was created.
* `id` - Shared directory identifier.
* `last_updated_time` - Date and time, in UTC and extended RFC 3339 format, when the shared directory was last updated.
* `owner_account_id` - Identifier of the directory owner account, which contains the directory that has been shared to the consumer account.
* `owner_directory_id` - Identifier of the directory in the directory owner account.
* `shared_account_id` - Identifier of the directory consumer account that has access to the shared directory (`owner_directory_id`) in the directory owner account.
* `shared_directory_id` - Date and time, in UTC and extended RFC 3339 format, when the directory service shared.

## Import

DirectoryService shared  directories can be imported using the directory `id`, e.g.

```
$ terraform import aws_directory_service_share_directory.example sharedDirectoryID
```
