---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_share_directory"
description: |-
  Accepts a directory sharing in AWS Directory Service.
---

# Resource: aws_directory_service_share_directory

Accepts a directory sharing in AWS Directory Service.

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
  share_method = "HANDSHAKE"
  share_notes  = "Terraform testing"

  share_target {
    id   = data.aws_caller_identity.member.account_id
    type = "ACCOUNT"
  }
}

resource "aws_directory_service_share_directory_accepter" "example" {
  shared_directory_id = aws_directory_service_share_directory.test.shared_directory_id
}
```

## Argument Reference

The following arguments are required:

* `shared_directory_id` - (Required) Identifier of the shared directory in the directory consumer account. This identifier is different for each directory owner account.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of shared directory accepter.
