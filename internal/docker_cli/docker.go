package docker_cli

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerCli struct {
}

func New() *DockerCli {
	return &DockerCli{}
}

func (d *DockerCli) LaunchBenthosContainer(ctx context.Context, nodeId string, yamlBytes []byte) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Write config to temporary file
	configPath := fmt.Sprintf("/tmp/%s.yaml", nodeId)
	if err := os.WriteFile(configPath, yamlBytes, 0644); err != nil {
		return err
	}

	containerName := fmt.Sprintf("connect_%s", nodeId)
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: "docker.redpanda.com/redpandadata/connect:latest",
			Cmd:   []string{"run", "/config/" + nodeId + ".yaml"},
		},
		&container.HostConfig{
			Binds: []string{fmt.Sprintf("%s:/config/%s.yaml", configPath, nodeId)},
		},
		nil, nil, containerName)
	if err != nil {
		return err
	}

	return cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}
