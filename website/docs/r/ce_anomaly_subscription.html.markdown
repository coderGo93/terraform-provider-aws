---
subcategory: "CostExplorer"
layout: "aws"
page_title: "AWS: aws_ce_anomaly_subscription"
description: |-
  Provides a CostExplorer Anomaly Monitor
---

# Resource: aws_ce_anomaly_subscription

Provides an CostExplorer Anomaly Subscription.

## Example Usage

```terraform
resource "aws_ce_anomaly_monitor" "example" {
  monitor_dimension = "SERVICE"
  monitor_name      = "EXAMPLE ANOMALY MONITOR"
  monitor_type      = "DIMENSIONAL"
}
resource "aws_ce_anomaly_monitor" "example" {
  subscription_name = "EXAMPLE ANOMALY SUBSCRIPTION"
  threshold         = 0
  frequency         = "IMMEDIATE"
  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.id,
  ]
  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
```

## Argument Reference

The following arguments are required:

* `frequency` - (Required) Frequency that anomaly reports are sent over email.
* `monitor_arn_list` - (Required) A list of cost anomaly monitors.
* `subscriber` - (Required) Configuration block for the list of subscribers to notify.
* `subscription_name` - (Required) Name for the subscription.
* `threshold` - (Required) Dollar value that triggers a notification if the threshold is exceeded.

### `subscriber`

* `address` - (Optional) Email address or SNS Amazon Resource Name (ARN). This depends on the `type`.
* `status` - (Optional) Indicates if the subscriber accepts the notifications. Valid values are `CONFIRMED`, `DECLINED`
* `type` - (Optional) Notification delivery channel. Valid values are `EMAIL`, `SNS`


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - account ID of the Anomaly Subscription.
* `id` - Unique ID of the anomaly subscription.


## Import

`aws_ce_anomaly_subscription` can be imported using the id, e.g.

```
$ terraform import aws_ce_anomaly_subscription.example anomalySubscriptionID
```
