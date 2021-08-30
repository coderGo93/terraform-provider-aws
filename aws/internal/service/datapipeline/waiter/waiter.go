package waiter

import (
	"time"
)

const (
	// DataPipelineDefinitionOperationTimeout Maximum amount of time to wait for DataPipeline definition operation eventual consistency
	DataPipelineDefinitionOperationTimeout = 4 * time.Minute
)
