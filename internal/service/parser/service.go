package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/Cr4z1k/vkr/internal/model"
	"gopkg.in/yaml.v3"
)

var kafkaAddres = os.Getenv("KAFKA_BROKER")

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) ParseJsonToBenthosConfig(pipeline model.PipelineDefinition) (map[string]model.Paths, error) {
	cfgPaths := make(map[string]model.Paths)

	for _, node := range pipeline.Nodes {
		if node.Type == "join" && (node.Meta == nil || node.Meta.CacheKey == "") {
			return nil, fmt.Errorf("join node %s must have a non-empty cache key in metadata", node.ID)
		} else if node.Type == "join" && node.Meta != nil {
			if node.Meta.DefaultTTL == nil {
				node.Meta.DefaultTTL = pointer.ToString("5m")
			}
		}

		// parse YAML strings to maps
		inputMap, err := parseYamlStringToMap(node.Input)
		if err != nil {
			return nil, fmt.Errorf("invalid input yaml for node %s: %w", node.ID, err)
		}

		outputMap, err := parseYamlStringToMap(node.Output)
		if err != nil {
			return nil, fmt.Errorf("invalid output yaml for node %s: %w", node.ID, err)
		}

		yamlBytes, err := generateBenthosConfig(node, pipeline, inputMap, outputMap, node.Meta)
		if err != nil {
			return nil, fmt.Errorf("error in generateBenthosConfig for nodeID - %s: %s", node.ID, err.Error())
		}

		if node.Type != "join" && node.Config != "" {
			yamlBytes = append(yamlBytes, []byte(defaultConfigStartingString)...)
			yamlBytes = append(yamlBytes, []byte(indentYaml(node.Config, 4))...)
		} else if node.Type == "join" {
			yamlBytes = append(yamlBytes, []byte(fmt.Sprintf(defaultJoinConfigTemplate, node.Meta.CacheKey))...)

			if node.Meta != nil && node.Meta.FilterCondition != nil && *node.Meta.FilterCondition != "" {
				yamlBytes = append(yamlBytes, []byte(fmt.Sprintf(joinFilterCondition, *node.Meta.FilterCondition))...)
			}
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

// parseYamlStringToMap parses a YAML string into map[string]any.
// Returns nil if the string is empty.
func parseYamlStringToMap(yamlStr string) (map[string]any, error) {
	if yamlStr == "" {
		return nil, nil
	}

	var m map[string]any
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return nil, err
	}

	return m, nil
}

// generateBenthosConfig builds a Benthos YAML config for a single node.
// Allows custom input (node.Input) and custom output (node.Output).
func generateBenthosConfig(
	node model.Node,
	pipeline model.PipelineDefinition,
	inputMap, outputMap map[string]any,
	meta *model.MetaData,
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

	cfg["logger"] = map[string]any{
		"level":  "TRACE",  // всегда TRACE, чтобы видеть все логи
		"format": "logfmt", // формат логов
	}

	// Input если inputMap задан — используем его, иначе дефолтный Kafka
	if inputMap != nil {
		cfg["input"] = inputMap
	} else {
		cfg["input"] = map[string]any{"kafka": map[string]any{
			"addresses":      []string{kafkaAddres},
			"topics":         inputs,
			"consumer_group": fmt.Sprintf("%s_%s_group", pipeline.Name, node.ID),
		}}
	}

	// Cache resources for join nodes
	if node.Type == "join" {
		cfg["cache_resources"] = []any{
			map[string]any{
				"label": "join_cache",
				"memory": map[string]any{
					"default_ttl": meta.DefaultTTL,
				},
			},
		}
	}

	// Output: если outputMap задан — используем его, иначе дефолтный Kafka
	if outputMap != nil {
		cfg["output"] = outputMap
	} else {
		cfg["output"] = makeOutputMap(outputs)
	}

	return yaml.Marshal(cfg)
}

func makeOutputMap(outputTopics []string) map[string]any {
	if len(outputTopics) == 1 {
		return map[string]any{
			"kafka": map[string]any{
				"addresses": []string{kafkaAddres},
				"topic":     outputTopics[0],
			},
		}
	}

	// multiple topics -> broker fan_out
	var outs []any
	for _, t := range outputTopics {
		out := map[string]any{
			"kafka": map[string]any{
				"addresses": []string{kafkaAddres},
				"topic":     t,
			},
		}

		outs = append(outs, out)
	}

	return map[string]any{
		"broker": map[string]any{
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

func indentYaml(yamlStr string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(yamlStr, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			lines[i] = pad + l
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
