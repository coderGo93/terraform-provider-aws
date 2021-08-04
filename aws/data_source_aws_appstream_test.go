package aws

import (
	"testing"
)

func TestAccAWSAppStreamDatasource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Image": {
			"name_basic": testAccAwsAppStreamImageDataSourceImage_nameBasic,
			"arn_basic":  testAccAwsAppStreamImageDataSourceImage_arnBasic,
		},
		"Images": {
			"basic": testAccAwsAppStreamImageDataSourcesImage,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
