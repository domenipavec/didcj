package config

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

const DaemonPort = "3333"

type Config struct {
	NumberOfNodes  int `json:"number_of_nodes"`
	MaxMsgsPerNode int `json:"max_msgs_per_node"`
	MaxMsgSizeMb   int `json:"max_msg_size_mb"`
	MaxMemoryMb    int `json:"max_memory_mb"`
	MaxTimeSeconds int `json:"max_time_seconds"`
}

func Get() (*Config, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "config.get")
	}
	defer f.Close()

	config := &Config{}
	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, errors.Wrap(err, "config.get")
	}

	return config, nil
}
