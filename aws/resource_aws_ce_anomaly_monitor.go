package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/costexplorer/waiter"
)

func resourceAwsCEAnomalyMonitor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsCEAnomalyMonitorCreate,
		ReadWithoutTimeout:   resourceAwsCEAnomalyMonitorRead,
		UpdateWithoutTimeout: resourceAwsCEAnomalyMonitorUpdate,
		DeleteWithoutTimeout: resourceAwsCEAnomalyMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dimensional_value_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_evaluated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitor_dimension": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(costexplorer.MonitorDimension_Values(), false),
			},
			"monitor_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"monitor_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     schemaAWSCECostCategoryRule(),
			},
			"monitor_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(costexplorer.MonitorType_Values(), false),
			},
		},
	}
}

func schemaAWSCECostCategoryRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cost_category": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 50),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"dimension": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.Dimension_Values(), false),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceAwsCEAnomalyMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn
	input := &costexplorer.AnomalyMonitor{
		MonitorName: aws.String(d.Get("monitor_name").(string)),
		MonitorType: aws.String(d.Get("monitor_type").(string)),
	}

	if v, ok := d.GetOk("monitor_dimension"); ok {
		input.MonitorDimension = aws.String(v.(string))
	}

	if v, ok := d.GetOk("monitor_specification"); ok {
		input.MonitorSpecification = expandCECostExpressions(v.([]interface{}))[0]
	}
	var err error
	var output *costexplorer.CreateAnomalyMonitorOutput
	err = resource.RetryContext(ctx, waiter.AnomalyMonitorOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateAnomalyMonitor(&costexplorer.CreateAnomalyMonitorInput{
			AnomalyMonitor: input,
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateAnomalyMonitor(&costexplorer.CreateAnomalyMonitorInput{
			AnomalyMonitor: input,
		})
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating CE Anomaly Monitor (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.MonitorArn))

	return resourceAwsCEAnomalyMonitorRead(ctx, d, meta)
}

func resourceAwsCEAnomalyMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	resp, err := conn.GetAnomalyMonitorsWithContext(ctx, &costexplorer.GetAnomalyMonitorsInput{MonitorArnList: []*string{aws.String(d.Id())}})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CE Anomaly Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CE Anomaly Monitor (%s): %w", d.Id(), err))
	}
	for _, v := range resp.AnomalyMonitors {
		d.Set("creation_date", v.CreationDate)
		d.Set("dimensional_value_count", v.DimensionalValueCount)
		d.Set("last_evaluated_date", v.LastEvaluatedDate)
		d.Set("last_updated_date", v.LastUpdatedDate)
		d.Set("monitor_dimension", v.MonitorDimension)
		d.Set("monitor_name", v.MonitorName)
		if err = d.Set("monitor_specification", flattenCECostCategoryRuleExpression(v.MonitorSpecification)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for CE Anomaly Monitor (%s): %w", "monitor_specification", d.Id(), err))
		}
		d.Set("monitor_type", v.MonitorType)
	}

	return nil
}

func resourceAwsCEAnomalyMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	input := &costexplorer.UpdateAnomalyMonitorInput{
		MonitorArn: aws.String(d.Id()),
	}

	if d.HasChange("monitor_name") {
		input.MonitorName = aws.String(d.Get("monitor_name").(string))
	}

	_, err := conn.UpdateAnomalyMonitorWithContext(ctx, input)

	if err != nil {
		diag.FromErr(fmt.Errorf("error updating CE Anomaly Monitor (%s): %w", d.Id(), err))
	}

	return resourceAwsCEAnomalyMonitorRead(ctx, d, meta)
}

func resourceAwsCEAnomalyMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	_, err := conn.DeleteAnomalyMonitorWithContext(ctx, &costexplorer.DeleteAnomalyMonitorInput{
		MonitorArn: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting CE Anomaly Monitor (%s): %w", d.Id(), err))
	}

	return nil
}
func expandCECostExpression(tfMap map[string]interface{}) *costexplorer.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.Expression{}
	if v, ok := tfMap["cost_category"]; ok {
		apiObject.CostCategories = expandCECostExpressionCostCategory(v.([]interface{}))
	}
	if v, ok := tfMap["dimension"]; ok {
		apiObject.Dimensions = expandCECostExpressionDimension(v.([]interface{}))
	}
	if v, ok := tfMap["tags"]; ok {
		apiObject.Tags = expandCECostExpressionTag(v.([]interface{}))
	}

	return apiObject
}

func expandCECostExpressionCostCategory(tfList []interface{}) *costexplorer.CostCategoryValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.CostCategoryValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = expandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = expandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressionDimension(tfList []interface{}) *costexplorer.DimensionValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.DimensionValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = expandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = expandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressionTag(tfList []interface{}) *costexplorer.TagValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.TagValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = expandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = expandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressions(tfList []interface{}) []*costexplorer.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCECostExpression(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCECostCategoryRuleExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cost_category"] = flattenCECostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCECostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["tags"] = flattenCECostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCECostCategoryRuleExpressionCostCategory(apiObject *costexplorer.CostCategoryValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpressionDimension(apiObject *costexplorer.DimensionValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpressionTag(apiObject *costexplorer.TagValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}
