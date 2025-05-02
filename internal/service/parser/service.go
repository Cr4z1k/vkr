package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Cr4z1k/vkr/internal/model"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"gopkg.in/yaml.v3"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) ParseJsonToBenthosConfig(pipeline configs.PipelineDefinition) (map[string]model.Paths, error) {
	cfgPaths := make(map[string]model.Paths)

	for _, node := range pipeline.Nodes {
		yamlBytes, err := generateBenthosConfig(node, pipeline)
		if err != nil {
			return nil, fmt.Errorf("error in generateBenthosConfig for nodeID - %s: %s", node.ID, err.Error())
		}

		configDir := filepath.Join(os.TempDir(), "ork_benthos_configs")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create config directory: %w", err)
		}
		configFile := fmt.Sprintf("%s_%s.yaml", pipeline.Name, node.ID)
		hostPath := filepath.Join(configDir, configFile)
		if err := os.WriteFile(hostPath, yamlBytes, 0644); err != nil {
			return nil, fmt.Errorf("error writing config file: %w", err)
		}

		if err := validateWithRPK(hostPath); err != nil {
			return nil, fmt.Errorf("%w: %w", model.ErrBenthosValidation, err)
		}

		cfgPaths[node.ID] = model.Paths{
			ConfigDir:  configDir,
			ConfigFile: configFile,
		}
	}

	return cfgPaths, nil
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

	cfg := make(map[string]any)

	// Input: custom plugin if provided, otherwise default Kafka
	if node.Input != nil {
		cfg["input"] = node.Input
	} else {
		cfg["input"] = map[string]any{"kafka": map[string]any{
			"addresses":      []string{"localhost:9092"},
			"topics":         inputs,
			"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
		}}
	}

	// Processors: only for non-sink nodes
	if node.Type != "sink" {
		cfg["pipeline"] = map[string]any{"processors": []any{node.Config}}
	}

	// Output: custom plugin for sink, otherwise default Kafka
	if node.Type == "sink" && node.Output != nil {
		cfg["output"] = node.Output
	} else {
		cfg["output"] = makeOutputMap(outputs)
	}

	return yaml.Marshal(cfg)
}

func makeOutputMap(outputTopics []string) map[string]interface{} {
	if len(outputTopics) == 1 {
		return map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses": []string{"localhost:9092"},
				"topic":     outputTopics[0],
			},
		}
	}
	// multiple topics -> broker fan_out
	var outs []interface{}
	for _, t := range outputTopics {
		outs = append(outs, map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses": []string{"localhost:9092"},
				"topic":     t,
			},
		})
	}
	return map[string]interface{}{
		"broker": map[string]interface{}{
			"pattern": "fan_out",
			"outputs": outs,
		},
	}
}

func validateWithRPK(configPath string) error {
	cmd := exec.Command("rpk", "connect", "lint", configPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rpk lint failed: %w\n%s", err, string(out))
	}
	return nil
}
