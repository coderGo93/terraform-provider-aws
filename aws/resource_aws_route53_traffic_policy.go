package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/waiter"
)

func resourceAwsRoute53TrafficPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsRoute53TrafficPolicyCreate,
		ReadWithoutTimeout:   resourceAwsRoute53TrafficPolicyRead,
		UpdateWithoutTimeout: resourceAwsRoute53TrafficPolicyUpdate,
		DeleteWithoutTimeout: resourceAwsRoute53TrafficPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"document": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 102400),
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53TrafficPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).r53conn

	input := &route53.CreateTrafficPolicyInput{
		Document: aws.String(d.Get("document").(string)),
		Name:     aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		input.Comment = aws.String(v.(string))
	}

	var err error
	var output *route53.CreateTrafficPolicyOutput
	err = resource.RetryContext(ctx, waiter.TrafficPolicyTimeout, func() *resource.RetryError {
		output, err = conn.CreateTrafficPolicyWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
				resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateTrafficPolicyWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Route53 traffic policy: %w", err))
	}

	d.SetId(fmt.Sprintf("%s/%d", aws.StringValue(output.TrafficPolicy.Id), aws.Int64Value(output.TrafficPolicy.Version)))

	return resourceAwsRoute53TrafficPolicyRead(ctx, d, meta)
}

func resourceAwsRoute53TrafficPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).r53conn

	id, version, err := decodeTrafficPolicyID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error decoding Route53 Traffic Policy %s : %w", d.Get("name").(string), err))
	}

	request := &route53.GetTrafficPolicyInput{
		Id:      aws.String(id),
		Version: aws.Int64(version),
	}

	response, err := conn.GetTrafficPolicy(request)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
		log.Printf("[WARN] Route53 Traffic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Route53 Traffic Policy %s, version %d: %w", d.Get("name").(string), d.Get("latest_version").(int), err))
	}

	d.Set("comment", response.TrafficPolicy.Comment)
	d.Set("document", response.TrafficPolicy.Document)
	d.Set("name", response.TrafficPolicy.Name)
	d.Set("type", response.TrafficPolicy.Type)

	return nil
}

func resourceAwsRoute53TrafficPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).r53conn

	id, version, err := decodeTrafficPolicyID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error decoding Route53 Traffic Policy %s : %w", d.Get("name").(string), err))
	}

	input := &route53.UpdateTrafficPolicyCommentInput{
		Id:      aws.String(id),
		Version: aws.Int64(version),
	}

	if d.HasChange("comment") {
		input.Comment = aws.String(d.Get("comment").(string))
	}

	_, err = conn.UpdateTrafficPolicyCommentWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Route53 Traffic Policy: %s. %w", d.Get("name").(string), err))
	}

	return resourceAwsRoute53TrafficPolicyRead(ctx, d, meta)
}

func resourceAwsRoute53TrafficPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).r53conn

	idTraffic, version, err := decodeTrafficPolicyID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error decoding Route53 Traffic Policy %s : %w", d.Get("name").(string), err))
	}

	input := &route53.DeleteTrafficPolicyInput{
		Id:      aws.String(idTraffic),
		Version: aws.Int64(version),
	}

	_, err = conn.DeleteTrafficPolicyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Route53 Traffic Policy %s, version %d: %s", idTraffic, version, err))
	}

	return nil
}

func getTrafficPolicyById(ctx context.Context, conn *route53.Route53, trafficPolicyId string, version int64) (*route53.TrafficPolicySummary, error) {
	var idMarker *string

	for allPoliciesListed := false; !allPoliciesListed; {
		input := &route53.ListTrafficPoliciesInput{}

		if idMarker != nil {
			input.TrafficPolicyIdMarker = idMarker
		}

		listResponse, err := conn.ListTrafficPoliciesWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, summary := range listResponse.TrafficPolicySummaries {
			if aws.StringValue(summary.Id) == trafficPolicyId && aws.Int64Value(summary.LatestVersion) == version {
				return summary, nil
			}
		}

		if aws.BoolValue(listResponse.IsTruncated) {
			idMarker = listResponse.TrafficPolicyIdMarker
		} else {
			allPoliciesListed = true
		}
	}

	return nil, nil
}

func decodeTrafficPolicyID(id string) (string, int64, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", int64(0), fmt.Errorf("expected ID in the form of id/version, given: %q", id)
	}
	version, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		return "", int64(0), err
	}

	return idParts[0], version, nil
}
