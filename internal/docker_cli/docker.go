package docker_cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	dockerImage = "docker.redpanda.com/redpandadata/connect:latest"
)

type DockerCli struct {
	cli *client.Client
}

func New(cli *client.Client) *DockerCli {
	return &DockerCli{
		cli: cli,
	}
}

func (d *DockerCli) LaunchBenthosContainer(ctx context.Context, pipelineName, nodeID string, yamlBytes []byte) error {
	containerName := fmt.Sprintf("connect_%s_%s", pipelineName, nodeID)

	existing, err := d.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: filters.NewArgs(filters.Arg("name", containerName))})
	if err != nil {
		return err
	}
	for _, ctr := range existing {
		_ = d.cli.ContainerStop(ctx, ctr.ID, container.StopOptions{})
		_ = d.cli.ContainerRemove(ctx, ctr.ID, container.RemoveOptions{Force: true})
	}

	configPath := fmt.Sprintf("/tmp/%s_%s.yaml", pipelineName, nodeID)
	if err := os.WriteFile(configPath, yamlBytes, 0644); err != nil {
		return err
	}

	resp, err := d.cli.ContainerCreate(
		ctx,
		&container.Config{Image: dockerImage, Cmd: []string{"run", "/config/" + pipelineName + "_" + nodeID + ".yaml"}},
		&container.HostConfig{Binds: []string{fmt.Sprintf("%s:/config/%s_%s.yaml", configPath, pipelineName, nodeID)}},
		nil, nil, containerName,
	)
	if err != nil {
		return err
	}

	return d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

// cleanupRemovedContainers stops and removes containers not in any of the given pipelines.
func (d *DockerCli) СleanupRemovedContainers(ctx context.Context, pipelines []configs.PipelineDefinition) error {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: filters.NewArgs(filters.Arg("name", "connect_"))})
	if err != nil {
		return err
	}

	desired := make(map[string]struct{})
	for _, p := range pipelines {
		for _, n := range p.Nodes {
			desired[fmt.Sprintf("connect_%s_%s", p.Name, n.ID)] = struct{}{}
		}
	}

	for _, ctr := range containers {
		for _, raw := range ctr.Names {
			name := strings.TrimPrefix(raw, "/")
			if _, ok := desired[name]; !ok {
				log.Printf("removing obsolete container %s", name)
				_ = d.cli.ContainerStop(ctx, ctr.ID, container.StopOptions{})
				_ = d.cli.ContainerRemove(ctx, ctr.ID, container.RemoveOptions{Force: true})
			}
		}
	}

	return nil
}
