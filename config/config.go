package config

import (
	"encoding/json"
	"os"

	"github.com/matematik7/didcj/models"
	"github.com/pkg/errors"
)

const KB = 1024
const MB = KB * 1024

const DaemonPort = "3333"

type Config struct {
	NumberOfNodes  int `json:"number_of_nodes"`
	MaxMsgsPerNode int `json:"max_msgs_per_node"`
	MaxMsgSizeMb   int `json:"max_msg_size_mb,omitempty"`
	MaxMsgSizeKb   int `json:"max_msg_size_kb,omitempty"`
	MaxMsgSize     int `json:"max_msg_size,omitempty"`
	MaxMemoryMb    int `json:"max_memory_mb,omitempty"`
	MaxMemoryKb    int `json:"max_memory_kb,omitempty"`
	MaxMemory      int `json:"max_memory,omitempty"`
	MaxTimeSeconds int `json:"max_time_seconds"`

	Input []Input `json:"input"`

	Servers []*models.Server `json:"servers"`
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

	if config.MaxMemory == 0 {
		if config.MaxMemoryKb != 0 {
			config.MaxMemory = config.MaxMemoryKb * KB
			config.MaxMemoryKb = 0
			config.MaxMemoryMb = 0
		} else if config.MaxMemoryMb != 0 {
			config.MaxMemory = config.MaxMemoryMb * MB
			config.MaxMemoryKb = 0
			config.MaxMemoryMb = 0
		} else {
			return nil, errors.New("Need to specify one of max memory options")
		}
	}

	if config.MaxMsgSize == 0 {
		if config.MaxMsgSizeKb != 0 {
			config.MaxMsgSize = config.MaxMsgSizeKb * KB
			config.MaxMsgSizeKb = 0
			config.MaxMsgSizeMb = 0
		} else if config.MaxMsgSizeMb != 0 {
			config.MaxMsgSize = config.MaxMsgSizeMb * MB
			config.MaxMsgSizeKb = 0
			config.MaxMsgSizeMb = 0
		} else {
			return nil, errors.New("Need to specify one of max msg size options")
		}
	}

	return config, nil
}
