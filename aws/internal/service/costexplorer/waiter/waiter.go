package waiter

import (
	"time"
)

const (
	// CostCategoryDefinitionOperationTimeout Maximum amount of time to wait for Cost Category eventual consistency
	CostCategoryDefinitionOperationTimeout = 4 * time.Minute
)
