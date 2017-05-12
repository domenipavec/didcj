package google

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	compute "google.golang.org/api/compute/v1"

	"github.com/matematik7/didcj/models"
	"github.com/matematik7/didcj/utils"
)

type Google struct {
	service *compute.Service
	config  *Config
	zone    string
}

func New() *Google {
	return &Google{
		zone: "europe-west1-d",
	}
}

func (g *Google) Init() error {
	client, err := buildOAuthHTTPClient(compute.ComputeScope)
	if err != nil {
		return err
	}

	g.service, err = compute.New(client)
	if err != nil {
		return err
	}

	config, err := getConfig()
	if err != nil {
		return err
	}
	g.config = config

	return nil
}

func (g *Google) getImage() (string, error) {
	resp, err := g.service.Images.GetFromFamily("ubuntu-os-cloud", "ubuntu-1604-lts").Context(context.Background()).Do()
	if err != nil {
		return "", err
	}

	return resp.SelfLink, nil
}

func (g *Google) Start(n int) error {
	sourceDiskImage, err := g.getImage()
	if err != nil {
		return err
	}

	machineType := fmt.Sprintf("zones/%s/machineTypes/n1-standard-1", g.zone)

	instance := &compute.Instance{
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			&compute.AttachedDisk{
				Boot:       true,
				AutoDelete: true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: sourceDiskImage,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				Network: "global/networks/default",
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
			},
		},
	}
	for i := 0; i < n; i++ {
		instance.Name = utils.GetName(i)
		_, err := g.service.Instances.Insert(g.config.Installed.ProjectID, g.zone, instance).Context(context.Background()).Do()
		if err != nil {
			return err
		}

		log.Println("Started", instance.Name)
	}

	started := 0
	for started < n {
		log.Println("Waiting for instances to start...")

		started = 0
		req := g.service.Instances.List(g.config.Installed.ProjectID, g.zone)
		err := req.Pages(context.Background(), func(page *compute.InstanceList) error {
			for _, instance := range page.Items {
				if instance.Status == "RUNNING" {
					started++
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		time.Sleep(time.Second)
	}
	return nil
}

func (g *Google) Stop() error {
	stillRunning := 0

	req := g.service.Instances.List(g.config.Installed.ProjectID, g.zone)
	err := req.Pages(context.Background(), func(page *compute.InstanceList) error {
		for _, instance := range page.Items {
			_, err := g.service.Instances.Delete(g.config.Installed.ProjectID, g.zone, instance.Name).Context(context.Background()).Do()
			if err != nil {
				return err
			}
			log.Println("Removed", instance.Name)

			stillRunning++
		}
		return nil
	})
	if err != nil {
		return err
	}

	for stillRunning > 0 {
		log.Println("Waiting for instances to stop...")

		stillRunning = 0
		req := g.service.Instances.List(g.config.Installed.ProjectID, g.zone)
		err := req.Pages(context.Background(), func(page *compute.InstanceList) error {
			for _, instance := range page.Items {
				if instance.Status != "TERMINATED" {
					stillRunning++
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		time.Sleep(time.Second)
	}
	return nil
}

func (g *Google) Get() ([]*models.Server, error) {
	servers := []*models.Server{}

	req := g.service.Instances.List(g.config.Installed.ProjectID, g.zone)
	err := req.Pages(context.Background(), func(page *compute.InstanceList) error {
		for _, instance := range page.Items {
			servers = append(servers, &models.Server{
				Ip:        net.ParseIP(instance.NetworkInterfaces[0].AccessConfigs[0].NatIP),
				PrivateIp: net.ParseIP(instance.NetworkInterfaces[0].NetworkIP),
				Username:  "domen",
				Password:  "",
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return servers, nil
}
