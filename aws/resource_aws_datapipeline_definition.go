package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datapipeline/waiter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsDataPipelineDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsDataPipelineDefinitionCreate,
		ReadWithoutTimeout:   resourceAwsDataPipelineDefinitionRead,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parameter_objects": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
									"string_value": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 10240),
									},
								},
							},
							Set: parameterAttributestHash,
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
				Set: parameterObjectHash,
			},
			"parameter_values": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"string_value": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 10240),
						},
					},
				},
				Set: parameterValuesHash,
			},
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"pipeline_objects": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fields": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"ref_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"string_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 10240),
									},
								},
							},
							Set: pipelineFieldHash,
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
				Set: pipelineObjectHash,
			},
		},
	}
}

func resourceAwsDataPipelineDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).datapipelineconn

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.PutPipelineDefinitionInput{
		PipelineId:      aws.String(pipelineID),
		PipelineObjects: expandDataPipelineObjects(d.Get("pipeline_objects").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("parameter_objects"); ok {
		input.ParameterObjects = expandDataPipelineParameterObjects(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("parameter_values"); ok {
		input.ParameterValues = expandDataPipelineParameterValues(v.(*schema.Set).List())
	}

	var err error
	var output *datapipeline.PutPipelineDefinitionOutput
	err = resource.RetryContext(ctx, waiter.DataPipelineDefinitionOperationTimeout, func() *resource.RetryError {
		output, err = conn.PutPipelineDefinitionWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodeInternalServiceError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		if aws.BoolValue(output.Errored) {
			errors := getValidationError(output.ValidationErrors)
			if strings.Contains(errors.Error(), "role") {
				return resource.RetryableError(fmt.Errorf("error validating after creation DataPipeline Definition (%s): %w", d.Id(), errors))
			}
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.PutPipelineDefinitionWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating DataPipeline Definition (%s): %w", d.Id(), err))
	}

	if aws.BoolValue(output.Errored) {
		return diag.FromErr(fmt.Errorf("error validating after creation DataPipeline Definition (%s): %w", d.Id(), getValidationError(output.ValidationErrors)))
	}

	// Activate pipeline if enabled
	input2 := &datapipeline.ActivatePipelineInput{
		PipelineId: aws.String(pipelineID),
	}

	_, err = conn.ActivatePipelineWithContext(ctx, input2)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error activating DataPipeline Definition (%s): %w", d.Id(), err))
	}

	d.SetId(pipelineID)
	return resourceAwsDataPipelineDefinitionRead(ctx, d, meta)
}

func resourceAwsDataPipelineDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).datapipelineconn

	input := &datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(d.Id()),
	}

	resp, err := conn.GetPipelineDefinitionWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) ||
		tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
		log.Printf("[WARN] DataPipeline Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading DataPipeline Definition (%s): %w", d.Id(), err))
	}

	if err = d.Set("parameter_objects", flattenDataPipelineParameterObjects(resp.ParameterObjects)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for DataPipeline Definition (%s): %w", "parameter_objects", d.Id(), err))
	}
	if err = d.Set("parameter_values", flattenDataPipelineParameterValues(resp.ParameterValues)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for DataPipeline Definition (%s): %w", "parameter_objects", d.Id(), err))
	}
	if err = d.Set("pipeline_objects", flattenDataPipelineObjects(resp.PipelineObjects)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for DataPipeline Definition (%s): %w", "parameter_objects", d.Id(), err))
	}
	d.Set("pipeline_id", d.Id())

	return nil
}

func expandDataPipelineParameterObject(tfMap map[string]interface{}) *datapipeline.ParameterObject {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterObject{
		Attributes: expandDataPipelineParameterAttributes(tfMap["attributes"].(*schema.Set).List()),
		Id:         aws.String(tfMap["id"].(string)),
	}

	return apiObject
}

func expandDataPipelineParameterAttribute(tfMap map[string]interface{}) *datapipeline.ParameterAttribute {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterAttribute{
		Key:         aws.String(tfMap["key"].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandDataPipelineParameterAttributes(tfList []interface{}) []*datapipeline.ParameterAttribute {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterAttribute

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataPipelineParameterAttribute(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataPipelineParameterObjects(tfList []interface{}) []*datapipeline.ParameterObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterObject

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataPipelineParameterObject(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDataPipelineParameterObject(apiObject *datapipeline.ParameterObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["attributes"] = flattenDataPipelineParameterAttributes(apiObject.Attributes)
	tfMap["id"] = aws.StringValue(apiObject.Id)

	return tfMap
}

func flattenDataPipelineParameterAttribute(apiObject *datapipeline.ParameterAttribute) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenDataPipelineParameterAttributes(apiObjects []*datapipeline.ParameterAttribute) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDataPipelineParameterAttribute(apiObject))
	}

	return tfList
}

func flattenDataPipelineParameterObjects(apiObjects []*datapipeline.ParameterObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDataPipelineParameterObject(apiObject))
	}

	return tfList
}

func expandDataPipelineParameterValue(tfMap map[string]interface{}) *datapipeline.ParameterValue {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterValue{
		Id:          aws.String(tfMap["id"].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandDataPipelineParameterValues(tfList []interface{}) []*datapipeline.ParameterValue {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterValue

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataPipelineParameterValue(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDataPipelineParameterValue(apiObject *datapipeline.ParameterValue) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["id"] = aws.StringValue(apiObject.Id)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenDataPipelineParameterValues(apiObjects []*datapipeline.ParameterValue) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDataPipelineParameterValue(apiObject))
	}

	return tfList
}

func expandDataPipelineObject(tfMap map[string]interface{}) *datapipeline.PipelineObject {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.PipelineObject{
		Fields: expandDataPipelinePipelineFields(tfMap["fields"].(*schema.Set).List()),
		Id:     aws.String(tfMap["id"].(string)),
		Name:   aws.String(tfMap["name"].(string)),
	}

	return apiObject
}

func expandDataPipelinePipelineField(tfMap map[string]interface{}) *datapipeline.Field {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.Field{
		Key: aws.String(tfMap["key"].(string)),
	}

	if v, ok := tfMap["ref_value"]; ok && v.(string) != "" {
		apiObject.RefValue = aws.String(v.(string))
	}
	if v, ok := tfMap["string_value"]; ok && v.(string) != "" {
		apiObject.StringValue = aws.String(v.(string))
	}

	return apiObject
}

func expandDataPipelinePipelineFields(tfList []interface{}) []*datapipeline.Field {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.Field

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataPipelinePipelineField(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataPipelineObjects(tfList []interface{}) []*datapipeline.PipelineObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.PipelineObject

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataPipelineObject(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDataPipelineObject(apiObject *datapipeline.PipelineObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["fields"] = flattenDataPipelineParameterFields(apiObject.Fields)
	tfMap["id"] = aws.StringValue(apiObject.Id)
	tfMap["name"] = aws.StringValue(apiObject.Name)

	return tfMap
}

func flattenDataPipelineParameterField(apiObject *datapipeline.Field) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["ref_value"] = aws.StringValue(apiObject.RefValue)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenDataPipelineParameterFields(apiObjects []*datapipeline.Field) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDataPipelineParameterField(apiObject))
	}

	return tfList
}

func flattenDataPipelineObjects(apiObjects []*datapipeline.PipelineObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDataPipelineObject(apiObject))
	}

	return tfList
}

func parameterObjectHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%+v", m["attributes"].(*schema.Set)))
	buf.WriteString(m["id"].(string))
	return hashcode.String(buf.String())
}

func parameterAttributestHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["key"].(string))
	buf.WriteString(m["string_value"].(string))
	return hashcode.String(buf.String())
}

func parameterValuesHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["id"].(string))
	buf.WriteString(m["string_value"].(string))
	return hashcode.String(buf.String())
}

func pipelineObjectHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%+v", m["fields"].(*schema.Set)))
	buf.WriteString(m["id"].(string))
	buf.WriteString(m["name"].(string))
	return hashcode.String(buf.String())
}

func pipelineFieldHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["key"].(string))
	buf.WriteString(m["ref_value"].(string))
	buf.WriteString(m["string_value"].(string))
	return hashcode.String(buf.String())
}

func getValidationError(validationError []*datapipeline.ValidationError) error {
	var validationErrors error
	for _, error := range validationError {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("id: %s, error: %v", aws.StringValue(error.Id), aws.StringValueSlice(error.Errors)))
	}

	return validationErrors
}
