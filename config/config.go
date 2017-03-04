package config

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

const DaemonPort = "3333"

type StartConfig struct {
	NumberOfNodes int `json:"number_of_nodes"`
}

type RunConfig struct {
	MaxMsgsPerNode int `json:"max_msgs_per_node"`
	MaxMsgSizeMb   int `json:"max_msg_size_mb"`
	MaxMemoryMb    int `json:"max_memory_mb"`
	MaxTimeSeconds int `json:"max_time_seconds"`
}

type RemoteConfig struct {
	Start StartConfig `json:"start"`
	Run   RunConfig   `json:"run"`
}

func Get() (*RemoteConfig, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "config.get")
	}
	defer f.Close()

	config := &RemoteConfig{}
	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, errors.Wrap(err, "config.get")
	}

	return config, nil
}
