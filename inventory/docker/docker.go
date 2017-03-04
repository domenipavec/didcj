package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/matematik7/didcj/models"
)

type Docker struct {
	cli *client.Client
	ctx context.Context

	image string
}

func New() *Docker {
	return &Docker{
		image: "rastasheep/ubuntu-sshd:16.04",
	}
}

func (docker *Docker) Init() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	docker.cli = cli
	docker.ctx = context.Background()

	return nil
}

func (docker *Docker) Start(n int) error {
	reader, err := docker.cli.ImagePull(docker.ctx, docker.image, types.ImagePullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	err = docker.handleOutput(reader)
	if err != nil {
		log.Fatal(err)
	}

	config := &container.Config{
		Image:  docker.image,
		Labels: make(map[string]string),
	}
	config.Labels["didcj"] = "didcj"
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}
	for i := 0; i < n; i++ {
		name := docker.getName(i)
		container, err := docker.cli.ContainerCreate(
			docker.ctx,
			config,
			hostConfig,
			networkingConfig,
			name,
		)
		if err != nil {
			return err
		}
		err = docker.cli.ContainerStart(
			docker.ctx,
			container.ID,
			types.ContainerStartOptions{},
		)
		if err != nil {
			return err
		}
		log.Println("Started", name)
	}

	return nil
}

func (docker *Docker) Stop() error {
	args := filters.NewArgs()
	args.Add("label", "didcj=didcj")
	containers, err := docker.cli.ContainerList(docker.ctx, types.ContainerListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return err
	}

	for _, container := range containers {
		err = docker.cli.ContainerRemove(
			docker.ctx,
			container.ID,
			types.ContainerRemoveOptions{Force: true},
		)
		if err != nil {
			return err
		}
		fmt.Println("Removed", container.Names[0])
	}
	return nil
}

func (docker *Docker) Get() ([]*models.Server, error) {
	args := filters.NewArgs()
	args.Add("label", "didcj=didcj")
	containers, err := docker.cli.ContainerList(docker.ctx, types.ContainerListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	servers := make([]*models.Server, 0, len(containers))
	for _, container := range containers {
		servers = append(servers, &models.Server{
			Ip:       net.ParseIP(container.NetworkSettings.Networks["bridge"].IPAddress),
			Username: "root",
			Password: "root",
		})
	}

	return servers, nil
}

func (docker *Docker) getName(i int) string {
	return fmt.Sprintf("didcj-%d", i)
}

func (docker *Docker) handleOutput(reader io.ReadCloser) error {
	defer reader.Close()

	status := &struct {
		Status string `json:"status"`
	}{}

	decoder := json.NewDecoder(reader)
	for {
		err := decoder.Decode(status)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		log.Println(status.Status)
	}

	return nil
}
