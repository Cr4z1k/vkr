package parser

import (
	"fmt"

	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"gopkg.in/yaml.v3"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) ParseJsonToBenthosConfig(pipeline configs.PipelineDefinition) (map[string][]byte, error) {
	cfgs := make(map[string][]byte)

	for _, node := range pipeline.Nodes {
		yamlBytes, err := generateBenthosConfig(node, pipeline)
		if err != nil {
			return nil, fmt.Errorf("error in generateBenthosConfig for nodeID - %s: %s", node.ID, err.Error())
		}

		cfgs[node.ID] = yamlBytes
	}

	return cfgs, nil
}

// generateBenthosConfig creates a minimal Benthos YAML config for a node based on its type,
// config, and connections in the pipeline.
func generateBenthosConfig(node configs.Node, pipeline configs.PipelineDefinition) ([]byte, error) {
	// Gather incoming and outgoing topics
	var inputs, outputs []string
	for _, edge := range pipeline.Edges {
		if edge.To == node.ID {
			inputs = append(inputs, fmt.Sprintf("pipeline.%s.%s", pipeline.Name, edge.From))
		}
		if edge.From == node.ID {
			outputs = append(outputs, fmt.Sprintf("pipeline.%s.%s", pipeline.Name, edge.To))
		}
	}

	config := map[string]interface{}{
		"input": map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses":      []string{"localhost:9092"},
				"topics":         inputs,
				"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
			},
		},
		"pipeline": map[string]interface{}{
			"processors": []interface{}{node.Config},
		},
		"output": map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses":      []string{"localhost:9092"},
				"topics":         outputs,
				"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
			},
		},
	}

	return yaml.Marshal(config)
}
