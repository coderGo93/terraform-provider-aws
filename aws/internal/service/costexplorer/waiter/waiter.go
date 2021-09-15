package waiter

import (
	"time"
)

const (
	// CostCategoryDefinitionOperationTimeout Maximum amount of time to wait for Cost Category eventual consistency
	CostCategoryDefinitionOperationTimeout = 4 * time.Minute
	// AnomalyMonitorOperationTimeout Maximum amount of time to wait for AnomalyMonitor eventual consistency
	AnomalyMonitorOperationTimeout = 4 * time.Minute
	// AnomalySubscriptionOperationTimeout Maximum amount of time to wait for AnomalySubscription eventual consistency
	AnomalySubscriptionOperationTimeout = 4 * time.Minute
)
