package aws

import (
	"testing"
)

func TestAccAWSCE_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AnomalyMonitor": {
			"basic":      testAccAwsCEAnomalyMonitor_basic,
			"disappears": testAccAwsCEAnomalyMonitor_disappears,
		},
		"AnomalySubscription": {
			"basic":      TestAccAwsCEAnomalySubscription_basic,
			"disappears": TestAccAwsCEAnomalySubscription_disappears,
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
