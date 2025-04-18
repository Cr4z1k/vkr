package configs

import "context"

type Service interface {
	ParseJsonToBenthosConfig(pipeline PipelineDefinition) (map[string][]byte, error)
}

type Docker interface {
	LaunchBenthosContainer(ctx context.Context, pipelineName, nodeID string, yamlBytes []byte) error
	СleanupRemovedContainers(ctx context.Context, pipelines []PipelineDefinition) error
}
