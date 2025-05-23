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

const (
	configDir = "/config"
)

var kafkaAddres = os.Getenv("KAFKA_BROKERS")

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) ParseJsonToBenthosConfig(pipeline configs.PipelineDefinition) (map[string]model.Paths, error) {
	cfgPaths := make(map[string]model.Paths)

	for _, node := range pipeline.Nodes {
		// parse YAML strings to maps
		inputMap, err := parseYamlStringToMap(node.Input)
		if err != nil {
			return nil, fmt.Errorf("invalid input yaml for node %s: %w", node.ID, err)
		}
		outputMap, err := parseYamlStringToMap(node.Output)
		if err != nil {
			return nil, fmt.Errorf("invalid output yaml for node %s: %w", node.ID, err)
		}
		configMap, err := parseYamlStringToMap(node.Config)
		if err != nil {
			return nil, fmt.Errorf("invalid config yaml for node %s: %w", node.ID, err)
		}

		yamlBytes, err := generateBenthosConfig(node, pipeline, inputMap, outputMap, configMap)
		if err != nil {
			return nil, fmt.Errorf("error in generateBenthosConfig for nodeID - %s: %s", node.ID, err.Error())
		}

		fmt.Printf("%s:\n %s\n", node.ID, string(yamlBytes))

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

// parseYamlStringToMap parses a YAML string into map[string]interface{}.
// Returns nil if the string is empty.
func parseYamlStringToMap(yamlStr string) (map[string]interface{}, error) {
	if yamlStr == "" {
		return nil, nil
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// generateBenthosConfig builds a Benthos YAML config for a single node.
// Allows custom input (node.Input) and custom output (node.Output).
func generateBenthosConfig(
	node configs.Node,
	pipeline configs.PipelineDefinition,
	inputMap, outputMap, configMap map[string]interface{},
) ([]byte, error) {
	// Determine upstream topics (input topics for this node)
	var inputs []string
	for _, edge := range pipeline.Edges {
		if edge.To == node.ID {
			inputs = append(inputs, fmt.Sprintf("pipeline.%s.%s.%s", pipeline.Name, edge.From, edge.To))
		}
	}
	// Determine downstream topics (output topics for this node)
	var outputs []string
	for _, edge := range pipeline.Edges {
		if edge.From == node.ID {
			outputs = append(outputs, fmt.Sprintf("pipeline.%s.%s.%s", pipeline.Name, edge.From, edge.To))
		}
	}

	cfg := make(map[string]any)

	// Input: custom plugin if provided, otherwise default Kafka
	if inputMap != nil {
		cfg["input"] = inputMap
	} else {
		cfg["input"] = map[string]any{"kafka": map[string]any{
			"addresses":      []string{kafkaAddres},
			"topics":         inputs,
			"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
		}}
	}

	// Processors: only for non-sink nodes
	if node.Type != "sink" && configMap != nil {
		cfg["pipeline"] = map[string]any{"processors": []any{configMap}}
	}

	// Output: custom plugin for sink, otherwise default Kafka
	if node.Type == "sink" && outputMap != nil {
		cfg["output"] = outputMap
	} else {
		cfg["output"] = makeOutputMap(outputs)
	}

	return yaml.Marshal(cfg)
}

func makeOutputMap(outputTopics []string) map[string]interface{} {
	if len(outputTopics) == 1 {
		return map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses": []string{kafkaAddres},
				"topic":     outputTopics[0],
			},
		}
	}
	// multiple topics -> broker fan_out
	var outs []interface{}
	for _, t := range outputTopics {
		outs = append(outs, map[string]interface{}{
			"kafka": map[string]interface{}{
				"addresses": []string{kafkaAddres},
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
