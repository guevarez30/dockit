package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// Client wraps the Docker client with simplified methods
type Client struct {
	cli *client.Client
	ctx context.Context
}

// NewClient creates a new Docker client wrapper
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{
		cli: cli,
		ctx: context.Background(),
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	return c.cli.Close()
}

// ListContainers returns all containers (running and stopped)
func (c *Client) ListContainers(all bool) ([]types.Container, error) {
	return c.cli.ContainerList(c.ctx, container.ListOptions{All: all})
}

// StartContainer starts a stopped container
func (c *Client) StartContainer(id string) error {
	return c.cli.ContainerStart(c.ctx, id, container.StartOptions{})
}

// StopContainer stops a running container
func (c *Client) StopContainer(id string) error {
	return c.cli.ContainerStop(c.ctx, id, container.StopOptions{})
}

// RestartContainer restarts a container
func (c *Client) RestartContainer(id string) error {
	return c.cli.ContainerRestart(c.ctx, id, container.StopOptions{})
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(id string, force bool) error {
	return c.cli.ContainerRemove(c.ctx, id, container.RemoveOptions{Force: force})
}

// GetContainerLogs returns logs from a container
func (c *Client) GetContainerLogs(id string, follow bool) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(c.ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
}

// InspectContainer returns detailed information about a container
func (c *Client) InspectContainer(id string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(c.ctx, id)
}

// GetContainerStats returns statistics for a container
func (c *Client) GetContainerStats(id string) (container.StatsResponseReader, error) {
	return c.cli.ContainerStats(c.ctx, id, false)
}

// ListImages returns all images
func (c *Client) ListImages() ([]image.Summary, error) {
	return c.cli.ImageList(c.ctx, image.ListOptions{All: true})
}

// RemoveImage removes an image
func (c *Client) RemoveImage(id string, force bool) error {
	_, err := c.cli.ImageRemove(c.ctx, id, image.RemoveOptions{Force: force})
	return err
}

// PullImage pulls an image from a registry
func (c *Client) PullImage(imageName string) (io.ReadCloser, error) {
	return c.cli.ImagePull(c.ctx, imageName, image.PullOptions{})
}

// InspectImage returns detailed information about an image
func (c *Client) InspectImage(id string) (types.ImageInspect, error) {
	inspect, _, err := c.cli.ImageInspectWithRaw(c.ctx, id)
	return inspect, err
}

// PruneContainers removes stopped containers
func (c *Client) PruneContainers() (container.PruneReport, error) {
	return c.cli.ContainersPrune(c.ctx, filters.Args{})
}

// PruneImages removes dangling images
func (c *Client) PruneImages() (image.PruneReport, error) {
	return c.cli.ImagesPrune(c.ctx, filters.Args{})
}

// ListVolumes returns all volumes
func (c *Client) ListVolumes() ([]*volume.Volume, error) {
	volumeList, err := c.cli.VolumeList(c.ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	return volumeList.Volumes, nil
}

// RemoveVolume removes a volume
func (c *Client) RemoveVolume(name string, force bool) error {
	return c.cli.VolumeRemove(c.ctx, name, force)
}

// InspectVolume returns detailed information about a volume
func (c *Client) InspectVolume(name string) (volume.Volume, error) {
	return c.cli.VolumeInspect(c.ctx, name)
}

// PruneVolumes removes unused volumes
func (c *Client) PruneVolumes() (volume.PruneReport, error) {
	return c.cli.VolumesPrune(c.ctx, filters.Args{})
}

// ListNetworks returns all networks
func (c *Client) ListNetworks() ([]*network.Summary, error) {
	networks, err := c.cli.NetworkList(c.ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Convert []network.Summary to []*network.Summary
	result := make([]*network.Summary, len(networks))
	for i := range networks {
		result[i] = &networks[i]
	}
	return result, nil
}

// RemoveNetwork removes a network
func (c *Client) RemoveNetwork(id string) error {
	return c.cli.NetworkRemove(c.ctx, id)
}

// InspectNetwork returns detailed information about a network
func (c *Client) InspectNetwork(id string) (network.Inspect, error) {
	return c.cli.NetworkInspect(c.ctx, id, network.InspectOptions{})
}

// PruneNetworks removes unused networks
func (c *Client) PruneNetworks() (network.PruneReport, error) {
	return c.cli.NetworksPrune(c.ctx, filters.Args{})
}
