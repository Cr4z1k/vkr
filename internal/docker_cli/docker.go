package docker_cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Cr4z1k/vkr/internal/model"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	// dockerImage is the Redpanda Connect (Benthos) image
	dockerImage = "redpandadata/connect:latest"
)

type DockerCli struct {
	cli *client.Client
}

// New returns a DockerCli wrapper
func New(cli *client.Client) *DockerCli {
	return &DockerCli{cli: cli}
}

// LaunchBenthosContainer pulls the image, validates the config, and ensures the container is running
// attached to the specified Docker network (env DOCKER_NETWORK or default 'pipeline_net').
func (d *DockerCli) LaunchBenthosContainer(
	ctx context.Context,
	pipelineName, nodeID string,
	cfgPaths model.Paths,
) error {
	// Ensure image is present
	if err := d.ensureImage(ctx, dockerImage); err != nil {
		return fmt.Errorf("cannot pull image: %w", err)
	}

	// Determine network mode
	networkMode := os.Getenv("DOCKER_NETWORK")
	if networkMode == "" {
		networkMode = "pipeline_net"
	}

	containerName := fmt.Sprintf("connect_%s_%s", pipelineName, nodeID)

	// Stop and remove existing container
	existing, err := d.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}
	for _, ctr := range existing {
		_ = d.cli.ContainerStop(ctx, ctr.ID, container.StopOptions{})
		_ = d.cli.ContainerRemove(ctx, ctr.ID, container.RemoveOptions{Force: true})
	}
	// Create container with mounted config and network
	resp, err := d.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: dockerImage,
			Cmd:   []string{"run", "/config/" + cfgPaths.ConfigFile},
		},
		&container.HostConfig{
			Binds:       []string{"vkr_configs:/config"},
			NetworkMode: container.NetworkMode(networkMode),
		},
		nil, nil,
		containerName,
	)
	if err != nil {
		return fmt.Errorf("error creating container: %w", err)
	}

	// Start it
	if err := d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %w", err)
	}

	return nil
}

// CleanupRemovedContainers stops/removes containers not in desired pipelines
func (d *DockerCli) CleanupRemovedContainers(
	ctx context.Context,
	pipelines []configs.PipelineDefinition,
) error {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", "connect_")),
	})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
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
				_ = d.cli.ContainerStop(ctx, ctr.ID, container.StopOptions{})
				_ = d.cli.ContainerRemove(ctx, ctr.ID, container.RemoveOptions{Force: true})
			}
		}
	}
	return nil
}

// ensureImage pulls the Docker image if missing locally
func (d *DockerCli) ensureImage(
	ctx context.Context,
	imageName string,
) error {
	// Check local
	if _, err := d.cli.ImageInspect(ctx, imageName); err == nil {
		return nil
	}
	// Pull the image
	rc, err := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("error pulling image %s: %w", imageName, err)
	}
	defer rc.Close()
	_, _ = io.Copy(io.Discard, rc)
	return nil
}
