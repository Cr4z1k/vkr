package configs

import "context"

type Service interface {
	ParseJsonToBenthosConfig(pipeline PipelineDefinition) (map[string][]byte, error)
}

type Docker interface {
	LaunchBenthosContainer(ctx context.Context, nodeId string, yamlBytes []byte) error
}
