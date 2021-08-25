package aws

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/waiter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
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

	var versionMarker *string

	var trafficPolicies []*route53.TrafficPolicy

	for allPoliciesListed := false; !allPoliciesListed; {
		listRequest := &route53.ListTrafficPolicyVersionsInput{
			Id: aws.String(d.Id()),
		}
		if versionMarker != nil {
			listRequest.TrafficPolicyVersionMarker = versionMarker
		}

		listResponse, err := conn.ListTrafficPolicyVersions(listRequest)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing Route 53 Traffic Policy versions: %v", err))
		}

		trafficPolicies = append(trafficPolicies, listResponse.TrafficPolicies...)

		if *listResponse.IsTruncated {
			versionMarker = listResponse.TrafficPolicyVersionMarker
		} else {
			allPoliciesListed = true
		}
	}

	for _, trafficPolicy := range trafficPolicies {
		deleteRequest := &route53.DeleteTrafficPolicyInput{
			Id:      trafficPolicy.Id,
			Version: trafficPolicy.Version,
		}

		_, err := conn.DeleteTrafficPolicy(deleteRequest)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error deleting Route53 Traffic Policy %s, version %d: %s", *trafficPolicy.Id, *trafficPolicy.Version, err))
		}
	}

	return nil
}

func getTrafficPolicyById(ctx context.Context, conn *route53.Route53, trafficPolicyId string) (*route53.TrafficPolicySummary, error) {
	var idMarker *string

	for allPoliciesListed := false; !allPoliciesListed; {
		input := &route53.ListTrafficPoliciesInput{}

		if idMarker != nil {
			input.TrafficPolicyIdMarker = idMarker
		}

		listResponse, err := conn.ListTrafficPoliciesWithContext(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error listing Route 53 Traffic Policies: %w", err)
		}

		for _, summary := range listResponse.TrafficPolicySummaries {
			if aws.StringValue(summary.Id) == trafficPolicyId {
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
