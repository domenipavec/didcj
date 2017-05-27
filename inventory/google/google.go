package google

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	compute "google.golang.org/api/compute/v1"

	"github.com/matematik7/didcj/models"
	"github.com/matematik7/didcj/utils"
)

type Google struct {
	service      *compute.Service
	config       *Config
	zones        []string
	nodesPerZone int
}

func New() *Google {
	return &Google{
		zones:        []string{"europe-west1-d", "us-east1-b", "us-west1-b", "us-east4-c", "us-central1-f"},
		nodesPerZone: 69,
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

	runningInstances := make(map[string]bool)
	g.listInstances(func(page *compute.InstanceList, zone string) error {
		for _, instance := range page.Items {
			if instance.Status == "RUNNING" {
				runningInstances[instance.Name] = true
				log.Println("Already running", instance.Name)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	instance := &compute.Instance{
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
		if runningInstances[instance.Name] {
			continue
		}

		zone := g.zones[i/g.nodesPerZone]
		log.Println("Starting", instance.Name, "in", zone, "...")

		instance.MachineType = fmt.Sprintf("zones/%s/machineTypes/n1-standard-1", zone)
		_, err := g.service.Instances.Insert(g.config.Installed.ProjectID, zone, instance).Context(context.Background()).Do()
		if err != nil {
			return err
		}

		log.Println("Started", instance.Name)
	}

	started := 0
	for started < n {
		log.Println("Waiting for instances to start...")

		started = 0
		g.listInstances(func(page *compute.InstanceList, zone string) error {
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

func (g *Google) listInstances(cb func(page *compute.InstanceList, zone string) error) error {
	for _, zone := range g.zones {
		req := g.service.Instances.List(g.config.Installed.ProjectID, zone)
		err := req.Pages(context.Background(), func(page *compute.InstanceList) error {
			return cb(page, zone)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Google) Stop() error {
	stillRunning := 0

	err := g.listInstances(func(page *compute.InstanceList, zone string) error {
		for _, instance := range page.Items {
			_, err := g.service.Instances.Delete(g.config.Installed.ProjectID, zone, instance.Name).Context(context.Background()).Do()
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
		err := g.listInstances(func(page *compute.InstanceList, zone string) error {
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

	err := g.listInstances(func(page *compute.InstanceList, zone string) error {
		for _, instance := range page.Items {
			servers = append(servers, &models.Server{
				Name:      instance.Name,
				IP:        net.ParseIP(instance.NetworkInterfaces[0].AccessConfigs[0].NatIP),
				PrivateIP: net.ParseIP(instance.NetworkInterfaces[0].NetworkIP),
				Username:  "domen",
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Sort(models.ServerByName(servers))

	return servers, nil
}
