---
subcategory: "CostExplorer"
layout: "aws"
page_title: "AWS: aws_ce_anomaly_monitor"
description: |-
  Provides a CostExplorer Anomaly Monitor
---

# Resource: aws_ce_anomaly_monitor

Provides an CostExplorer Anomaly Monitor.

## Example Usage

```terraform
resource "aws_ce_anomaly_monitor" "example" {
  monitor_dimension = "SERVICE"
  monitor_name      = "EXAMPLE ANOMALY MONITOR"
  monitor_type      = "DIMENSIONAL"
}
```

## Argument Reference

The following arguments are required:

* `monitor_name` - (Required) Name of the monitor.
* `monitor_type` - (Required) Possible type values.

The following arguments are optional:

* `monitor_dimension` - (Optional) Dimensions to evaluate.
* `monitor_specification` - (Optional) Configuration block for the `Expression` object used to categorize costs. See below.


### `monitor_specification`

* `and` - (Optional) Return results that match both `Dimension` objects.
* `cost_category` - (Optional) Configuration block for the filter that's based on `CostCategory` values. See below.
* `dimension` - (Optional) Configuration block for the specific `Dimension` to use for `Expression`. See below.
* `not` - (Optional) Return results that match both `Dimension` object.
* `or` - (Optional) Return results that match both `Dimension` object.
* `tag` - (Optional) Configuration block for the specific `Tag` to use for `Expression`. See below.

### `cost_category`

* `key` - (Optional) Unique name of the Cost Category. 
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `dimension`

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `tag`

* `key` - (Optional) Key for the tag.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `creation_date` - Date when the monitor was created.
* `dimensional_value_count` - Value for evaluated dimensions.
* `id` - Unique ID of the anomaly monitor.
* `last_evaluated_date` - Date when the monitor last evaluated for anomalies.
* `last_updated_date` - Date when the monitor was last updated.


## Import

`aws_ce_anomaly_monitor` can be imported using the id, e.g.

```
$ terraform import aws_ce_anomaly_monitor.example anomalyMonitorID
```
