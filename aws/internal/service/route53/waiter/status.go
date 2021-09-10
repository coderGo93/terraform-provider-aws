package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/finder"
)

func ChangeInfoStatus(conn *route53.Route53, changeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53.GetChangeInput{
			Id: aws.String(changeID),
		}

		output, err := conn.GetChange(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ChangeInfo == nil {
			return nil, "", nil
		}

		return output.ChangeInfo, aws.StringValue(output.ChangeInfo.Status), nil
	}
}

func HostedZoneDnssecStatus(conn *route53.Route53, hostedZoneID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hostedZoneDnssec, err := finder.HostedZoneDnssec(conn, hostedZoneID)

		if err != nil {
			return nil, "", err
		}

		if hostedZoneDnssec == nil || hostedZoneDnssec.Status == nil {
			return nil, "", nil
		}

		return hostedZoneDnssec.Status, aws.StringValue(hostedZoneDnssec.Status.ServeSignature), nil
	}
}

func KeySigningKeyStatus(conn *route53.Route53, hostedZoneID string, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		keySigningKey, err := finder.KeySigningKey(conn, hostedZoneID, name)

		if err != nil {
			return nil, "", err
		}

		if keySigningKey == nil {
			return nil, "", nil
		}

		return keySigningKey, aws.StringValue(keySigningKey.Status), nil
	}
}

//TrafficPolicyInstanceState fetches the traffic policy instance and its state
func TrafficPolicyInstanceState(ctx context.Context, conn *route53.Route53, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := finder.TrafficPolicyInstanceID(ctx, conn, id)

		if err != nil {
			return nil, "Unknown", err
		}

		if object == nil && object.TrafficPolicyInstance == nil {
			return object, "NotFound", nil
		}

		return object, aws.StringValue(object.TrafficPolicyInstance.State), nil
	}
}
