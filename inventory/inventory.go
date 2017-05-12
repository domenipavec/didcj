package inventory

import (
	"fmt"

	"github.com/matematik7/didcj/inventory/docker"
	"github.com/matematik7/didcj/inventory/google"
	"github.com/matematik7/didcj/models"
)

type Inventory interface {
	Init() error
	Start(n int) error
	Stop() error
	Get() ([]*models.Server, error)
}

func Init(inventoryType string) (Inventory, error) {
	var inv Inventory

	if inventoryType == "docker" {
		inv = docker.New()
	} else if inventoryType == "google" {
		inv = google.New()
	} else {
		return nil, fmt.Errorf("Invalid inventory type")
	}

	err := inv.Init()
	if err != nil {
		return nil, err
	}

	return inv, nil
}
