package virtualization

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// initialize docker connection using docker socket
func initDockerConnection() *client.Client {
	apiClient, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to docker daemon: %s", err.Error())
		os.Exit(1)
	}

	return apiClient
}

type DockerVirtualization struct {
	Connection *client.Client
}

func NewDockerVirtualization() *DockerVirtualization {
	return &DockerVirtualization{
		Connection: initDockerConnection(),
	}
}

func (d DockerVirtualization) ListContainers() []container.Summary {
	containers, err := d.Connection.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		fmt.Println(err.Error())
	}

	return containers
}

func (d DockerVirtualization) ListPulledImages() []string {
	imagesList, err := d.Connection.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		fmt.Println(err)
	}

	// for now, only return a list of image name
	imagesNameList := []string{}
	for _, image := range imagesList {
		for _, imageTags := range image.RepoTags {
			imagesNameList = append(imagesNameList, imageTags)
		}
	}

	return imagesNameList
}

func (d DockerVirtualization) pullImage() error {
	thisNodeImageList := d.ListPulledImages()

	for _, image := range thisNodeImageList {
		if strings.Contains(image, "rancher/k3s") {
			return nil
		}
	}

	reader, err := d.Connection.ImagePull(context.Background(), "docker.io/library/k3s", image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	return nil
}

func (d DockerVirtualization) CreateMainContainer(token string) error {
	err := d.pullImage()
	if err != nil {
		return err
	}

	resp, err := d.Connection.ContainerCreate(context.Background(), &container.Config{
		Image:    "rancher/k3s",
		Cmd:      []string{"agent", "--server", "https://", "--token"},
		Hostname: "server-1",
	}, &container.HostConfig{
		Privileged:    true,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{},
		nil, "k3s-server-1")
	if err != nil {
		return err
	}

	err = d.Connection.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d DockerVirtualization) CreateWorkerContainer() error {
	err := d.pullImage()
	if err != nil {
		return err
	}

	_, err = d.Connection.ContainerCreate(context.Background(), &container.Config{
		Image:    "rancher/k3s",
		Cmd:      []string{"server"},
		Hostname: "worker-1",
	}, &container.HostConfig{
		Privileged: true,
		PortBindings: nat.PortMap{
			"6443/tcp": {{HostIP: "0.0.0.0", HostPort: "6443"}},
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{},
		nil, "k3s-server-1")
	if err != nil {
		return err
	}

	return nil
}
