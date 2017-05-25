package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/matematik7/didcj/models"
	"github.com/matematik7/didcj/utils"
	"github.com/pkg/errors"
)

type Docker struct {
	cli *client.Client
	ctx context.Context

	image string
}

func New() *Docker {
	return &Docker{
		image: "matematik7/didcj",
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
	filters, err := filters.ParseFlag(fmt.Sprintf("reference=%s", docker.image), filters.NewArgs())
	if err != nil {
		return errors.Wrap(err, "could not parse image filter params")
	}
	imgs, err := docker.cli.ImageList(docker.ctx, types.ImageListOptions{
		Filters: filters,
	})
	if err != nil {
		return errors.Wrap(err, "could not list images")
	}
	if len(imgs) <= 0 {
		reader, err := docker.cli.ImagePull(docker.ctx, docker.image, types.ImagePullOptions{})
		if err != nil {
			return errors.Wrap(err, "could not pull image")
		}
		err = docker.handleOutput(reader)
		if err != nil {
			return errors.Wrap(err, "could not handle output pull image")
		}
	}

	config := &container.Config{
		Image:  docker.image,
		Labels: make(map[string]string),
	}
	config.Labels["didcj"] = "didcj"
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}
	for i := 0; i < n; i++ {
		name := utils.GetName(i)
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
			Name:      container.Names[0],
			Ip:        net.ParseIP(container.NetworkSettings.Networks["bridge"].IPAddress),
			PrivateIp: net.ParseIP(container.NetworkSettings.Networks["bridge"].IPAddress),
			Username:  "root",
		})
	}

	sort.Sort(models.ServerByName(servers))

	return servers, nil
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
