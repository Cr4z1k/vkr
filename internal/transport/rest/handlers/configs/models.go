package configs

// PipelineDefinition describes the overall pipeline graph as received from the frontend (VueFlow).
type PipelineDefinition struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a processing block in the pipeline.
type Node struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`   // e.g., "transform", "filter", etc.
	Config map[string]interface{} `json:"config"` // user-defined settings
}

// Edge connects two nodes in the pipeline.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
