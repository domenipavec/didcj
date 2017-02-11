package inventory

import (
	"fmt"

	"github.com/matematik7/didcj/inventory/docker"
	"github.com/matematik7/didcj/inventory/server"
)

type Inventory interface {
	Init() error
	Start(n int) error
	Stop() error
	Get() ([]*server.Server, error)
}

func Init(inventoryType string) (Inventory, error) {
	var inv Inventory

	if inventoryType == "docker" {
		inv = docker.New()
	} else {
		return nil, fmt.Errorf("Invalid inventory type")
	}

	err := inv.Init()
	if err != nil {
		return nil, err
	}

	return inv, nil
}
