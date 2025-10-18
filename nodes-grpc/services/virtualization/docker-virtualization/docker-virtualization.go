package docker_virtualization

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"nodes-grpc/services/db"
	virtualization_model "nodes-grpc/services/model/virtualization-model"
	"nodes-grpc/services/virtualization"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	DOCKERFILE_NAME = "k3s_alpine.Dockerfile"
)

type DockerVirtualization struct {
	dockerConnection *client.Client
}

func NewDockerVirtualization(
	dockerConnection *client.Client, dbConnection db.DatabaseInterface,
) virtualization.VirtualizationInterface {
	return &DockerVirtualization{
		dockerConnection: dockerConnection,
	}
}

func (d *DockerVirtualization) CreateMaster(ctx context.Context, virtRequest virtualization_model.NodeCreateRequest) (string, error) {
    return "", nil
}

func (d *DockerVirtualization) CreateWorker(ctx context.Context, virtRequest virtualization_model.NodeCreateRequest) error {
    return nil
}

func (d *DockerVirtualization) StopNode(ctx context.Context, instanceId virtualization_model.InstanceIdentification) error {
    return nil
}

func (d *DockerVirtualization) Spawn(virtRequest virtualization_model.NodeCreateRequest) error {
	// check if image exists in local
	localImages, err := d.dockerConnection.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return err
	}

	for _, image := range localImages {
		imageTags := image.RepoTags

		for _, tag := range imageTags {
			if strings.Contains(tag, "docker-virt") {
				slog.Info("Spawn(): docker-virt image exists on local")
				return nil
			}
		}
	}

	err = d.buildImage()
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerVirtualization) Stop(instanceIdentification virtualization_model.InstanceIdentification) error {
	return nil
}

func (d *DockerVirtualization) List() error {
	return nil
}

func (d *DockerVirtualization) buildImage() error {
	dockerfileContext, err := getDockerfileTar("containers/k3s_alpine.Dockerfile")
	if err != nil {
		return err
	}

	buildResponse, err := d.dockerConnection.ImageBuild(context.Background(), dockerfileContext, types.ImageBuildOptions{
		Tags: []string{
			"localhost/docker-virt",
		},
		Context:    dockerfileContext,
		Dockerfile: DOCKERFILE_NAME,
		Remove:     true,
	})
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerVirtualization) listContainers() []container.Summary {
	containers, err := d.dockerConnection.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		fmt.Println(err.Error())
	}

	return containers
}

func (d *DockerVirtualization) listPulledImages() []string {
	imagesList, err := d.dockerConnection.ImageList(context.Background(), image.ListOptions{All: true})
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

func (d *DockerVirtualization) pullImage() error {
	thisNodeImageList := d.listPulledImages()

	for _, image := range thisNodeImageList {
		if strings.Contains(image, "rancher/k3s") {
			slog.Info("pullImage(): docker image is available on local")
			return nil
		}
	}

	reader, err := d.dockerConnection.ImagePull(context.Background(), "docker.io/library/k3s", image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	return nil
}

func (d *DockerVirtualization) createHeadContainer(token string) error {
	err := d.pullImage()
	if err != nil {
		return err
	}

	resp, err := d.dockerConnection.ContainerCreate(context.Background(), &container.Config{
		Image:    "rancher/k3s",
		Cmd:      []string{"server", "--token", token},
		Hostname: "server-1",
	}, &container.HostConfig{
		Privileged:    true,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{},
		nil, "k3s-server-1")
	if err != nil {
		return err
	}

	err = d.dockerConnection.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerVirtualization) createWorkerContainer(token string) error {
	err := d.pullImage()
	if err != nil {
		return err
	}

	_, err = d.dockerConnection.ContainerCreate(context.Background(), &container.Config{
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

func getDockerfileTar(dockerfilePath string) (*bytes.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFileReader, err := os.Open(dockerfilePath)
	if err != nil {
		return new(bytes.Reader), err
	}
	readDockerFile, err := io.ReadAll(dockerFileReader)
	if err != nil {
		return new(bytes.Reader), err
	}

	tarHeader := &tar.Header{
		Name: DOCKERFILE_NAME,
		Size: int64(len(readDockerFile)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return new(bytes.Reader), err
	}
	_, err = tw.Write(readDockerFile)
	if err != nil {
		return new(bytes.Reader), err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	return dockerFileTarReader, nil
}
