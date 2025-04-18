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

// generateBenthosConfig builds a Benthos YAML config for a single node.
// Allows custom input (node.Input) and custom output (node.Output).
func generateBenthosConfig(node configs.Node, pipeline configs.PipelineDefinition) ([]byte, error) {
	// Determine upstream topics
	var inputs []string
	for _, edge := range pipeline.Edges {
		if edge.To == node.ID {
			inputs = append(inputs, fmt.Sprintf("pipeline.%s.%s", pipeline.Name, edge.From))
		}
	}
	// Determine downstream topics
	var outputs []string
	for _, edge := range pipeline.Edges {
		if edge.From == node.ID {
			outputs = append(outputs, fmt.Sprintf("pipeline.%s.%s", pipeline.Name, edge.To))
		}
	}

	cfg := make(map[string]interface{})

	// Input: custom plugin if provided, otherwise default Kafka
	if node.Input != nil {
		cfg["input"] = node.Input
	} else {
		cfg["input"] = map[string]interface{}{"kafka": map[string]interface{}{
			"addresses":      []string{"localhost:9092"},
			"topics":         inputs,
			"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
		}}
	}

	// Processors: only for non-sink nodes
	if node.Type != "sink" {
		cfg["pipeline"] = map[string]interface{}{"processors": []interface{}{node.Config}}
	}

	// Output: custom plugin for sink, otherwise default Kafka
	if node.Type == "sink" && node.Output != nil {
		cfg["output"] = node.Output
	} else {
		cfg["output"] = map[string]interface{}{"kafka": map[string]interface{}{
			"addresses":      []string{"localhost:9092"},
			"topics":         outputs,
			"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
		}}
	}

	return yaml.Marshal(cfg)
}
