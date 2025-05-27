package model

// PipelineDefinition describes the overall pipeline graph as received from the frontend (VueFlow).
type PipelineDefinition struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a processing block in the pipeline.
type Node struct {
	ID     string    `json:"id"`
	Type   string    `json:"type"`
	Input  string    `json:"input,omitempty"`  // YAML string
	Output string    `json:"output,omitempty"` // YAML string
	Config string    `json:"config"`           // YAML string
	Meta   *MetaData `json:"meta,omitempty"`   // Additional metadata
}

// MetaData contains additional information for join node
type MetaData struct {
	CacheKey        string  `json:"cacheKey,omitempty"`        // Cache key for the node
	DefaultTTL      *string `json:"defaultTTL,omitempty"`      // Default TTL for cache resources
	FilterCondition *string `json:"filterCondition,omitempty"` // Condition for filtering messages YAML string
}

// Edge connects two nodes in the pipeline.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
