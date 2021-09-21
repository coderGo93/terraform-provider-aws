package aws

import (
	"bytes"
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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/costexplorer/waiter"
)

func resourceAwsCEAnomalySubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsCEAnomalySuscriptionCreate,
		ReadWithoutTimeout:   resourceAwsCEAnomalySuscriptionRead,
		UpdateWithoutTimeout: resourceAwsCEAnomalySuscriptionUpdate,
		DeleteWithoutTimeout: resourceAwsCEAnomalySuscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frequency": {
				Type:     schema.TypeString,
				Required: true,
			},
			"monitor_arn_list": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(6, 302),
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.SubscriberStatus_Values(), false),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.SubscriberType_Values(), false),
						},
					},
				},
				Set: ceAnomalySubscriptionSubscriber,
			},
			"subscription_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"threshold": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatAtLeast(0.0),
			},
		},
	}
}

func resourceAwsCEAnomalySuscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn
	input := &costexplorer.AnomalySubscription{
		Frequency:        aws.String(d.Get("frequency").(string)),
		MonitorArnList:   expandStringSet(d.Get("monitor_arn_list").(*schema.Set)),
		Subscribers:      expandCEAnomalySubscriptionSubscribers(d.Get("subscriber").(*schema.Set).List()),
		SubscriptionName: aws.String(d.Get("subscription_name").(string)),
		Threshold:        aws.Float64(d.Get("threshold").(float64)),
	}

	var err error
	var output *costexplorer.CreateAnomalySubscriptionOutput
	err = resource.RetryContext(ctx, waiter.AnomalySubscriptionOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateAnomalySubscription(&costexplorer.CreateAnomalySubscriptionInput{
			AnomalySubscription: input,
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
		output, err = conn.CreateAnomalySubscription(&costexplorer.CreateAnomalySubscriptionInput{
			AnomalySubscription: input,
		})
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating CE Anomaly Subscription (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.SubscriptionArn))

	return resourceAwsCEAnomalySuscriptionRead(ctx, d, meta)
}

func resourceAwsCEAnomalySuscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	resp, err := conn.GetAnomalySubscriptionsWithContext(ctx, &costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: []*string{aws.String(d.Id())}})
	if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, costexplorer.ErrCodeUnknownSubscriptionException, "No anomaly subscription") {
		log.Printf("[WARN] CE Anomaly Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CE Anomaly Subscription (%s): %w", d.Id(), err))
	}
	for _, v := range resp.AnomalySubscriptions {
		d.Set("account_id", v.AccountId)
		d.Set("frequency", v.Frequency)
		d.Set("monitor_arn_list", flattenStringSet(v.MonitorArnList))
		if err = d.Set("subscriber", flattenCEAnomalySubscriptionSubscribers(v.Subscribers)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for CE Anomaly Subscription (%s): %w", "subscriber", d.Id(), err))
		}
		d.Set("subscription_name", v.SubscriptionName)
		d.Set("threshold", v.Threshold)
	}

	return nil
}

func resourceAwsCEAnomalySuscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	input := &costexplorer.UpdateAnomalySubscriptionInput{
		SubscriptionArn: aws.String(d.Id()),
	}

	if d.HasChange("frequency") {
		input.Frequency = aws.String(d.Get("frequency").(string))
	}
	if d.HasChange("subscriber") {
		input.Subscribers = expandCEAnomalySubscriptionSubscribers(d.Get("monitor_arn_list").(*schema.Set).List())
	}
	if d.HasChange("monitor_arn_list") {
		input.MonitorArnList = expandStringSet(d.Get("monitor_arn_list").(*schema.Set))
	}
	if d.HasChange("subscription_name") {
		input.SubscriptionName = aws.String(d.Get("subscription_name").(string))
	}
	if d.HasChange("threshold") {
		input.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	_, err := conn.UpdateAnomalySubscriptionWithContext(ctx, input)

	if err != nil {
		diag.FromErr(fmt.Errorf("error updating CE Anomaly Subscription (%s): %w", d.Id(), err))
	}

	return resourceAwsCEAnomalySuscriptionRead(ctx, d, meta)
}

func resourceAwsCEAnomalySuscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).costexplorerconn

	_, err := conn.DeleteAnomalySubscriptionWithContext(ctx, &costexplorer.DeleteAnomalySubscriptionInput{
		SubscriptionArn: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, costexplorer.ErrCodeUnknownSubscriptionException, "No anomaly subscription") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting CE Anomaly Subscription (%s): %w", d.Id(), err))
	}

	return nil
}

func expandCEAnomalySubscriptionSubscriber(tfMap map[string]interface{}) *costexplorer.Subscriber {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.Subscriber{}

	if v, ok := tfMap["address"]; ok && v.(string) != "" {
		apiObject.Address = aws.String(v.(string))
	}
	if v, ok := tfMap["status"]; ok && v.(string) != "" {
		apiObject.Status = aws.String(v.(string))
	}
	if v, ok := tfMap["type"]; ok && v.(string) != "" {
		apiObject.Type = aws.String(v.(string))
	}

	return apiObject
}

func expandCEAnomalySubscriptionSubscribers(tfList []interface{}) []*costexplorer.Subscriber {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.Subscriber

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCEAnomalySubscriptionSubscriber(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCEAnomalySubscriptionSubscriber(apiObject *costexplorer.Subscriber) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["address"] = aws.StringValue(apiObject.Address)
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["status"] = aws.StringValue(apiObject.Status)

	return tfMap
}

func flattenCEAnomalySubscriptionSubscribers(apiObjects []*costexplorer.Subscriber) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCEAnomalySubscriptionSubscriber(apiObject))
	}

	return tfList
}

func ceAnomalySubscriptionSubscriber(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["address"].(string))
	buf.WriteString(m["status"].(string))
	buf.WriteString(m["type"].(string))
	return hashcode.String(buf.String())
}
