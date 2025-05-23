package configs

// PipelineDefinition describes the overall pipeline graph as received from the frontend (VueFlow).
type PipelineDefinition struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a processing block in the pipeline.
type Node struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Input  string `json:"input,omitempty"`  // YAML string
	Output string `json:"output,omitempty"` // YAML string
	Config string `json:"config"`           // YAML string
}

// Edge connects two nodes in the pipeline.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
