package configs

import (
	"context"

	"github.com/Cr4z1k/vkr/internal/model"
)

type Service interface {
	ParseJsonToBenthosConfig(pipeline PipelineDefinition) (map[string]model.Paths, error)
}

type Docker interface {
	LaunchBenthosContainer(ctx context.Context, pipelineName, nodeID string, cfgPaths model.Paths) error
	CleanupRemovedContainers(ctx context.Context, pipelines []PipelineDefinition) error
}
