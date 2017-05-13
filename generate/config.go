package generate

import (
	"encoding/json"
	"os"

	"github.com/matematik7/didcj/config"
	"github.com/pkg/errors"
)

func ConfigJson() error {
	constantReturnConfig, err := json.MarshalIndent(map[string]string{
		"value": "1e8",
	}, "\t\t\t", "\t")
	if err != nil {
		return errors.Wrap(err, "generate.ConfigJson constantReturnConfig")
	}
	randomRangeReturnConfig, err := json.MarshalIndent(map[string]float64{
		"min": 0,
		"max": 1e8,
	}, "\t\t\t", "\t")
	if err != nil {
		return errors.Wrap(err, "generate.ConfigJson randomRangeReturnConfig")
	}
	randomListReturnConfig, err := json.MarshalIndent(map[string][]string{
		"values": []string{"'('", "')'"},
	}, "\t\t\t", "\t")
	if err != nil {
		return errors.Wrap(err, "generate.ConfigJson randomListReturnConfig")
	}
	cfg := config.Config{
		NumberOfNodes:  100,
		MaxMsgsPerNode: 1000,
		MaxMsgSizeMb:   8,
		MaxMemoryMb:    128,
		MaxTimeSeconds: 2,
		Input: []config.Input{
			config.Input{
				Name:            "GetN",
				Inputs:          []string{},
				ReturnType:      "int64",
				ReturnGenerator: "CONSTANT",
				ReturnConfig:    constantReturnConfig,
			},
			config.Input{
				Name:            "GetA",
				Inputs:          []string{"int64"},
				ReturnType:      "int64",
				ReturnGenerator: "RANDOM_RANGE",
				ReturnConfig:    randomRangeReturnConfig,
			},
			config.Input{
				Name:            "GetB",
				Inputs:          []string{"int64"},
				ReturnType:      "int8",
				ReturnGenerator: "RANDOM_LIST",
				ReturnConfig:    randomListReturnConfig,
			},
		},
	}

	f, err := os.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "generate.ConfigJson file create")
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(cfg)
	if err != nil {
		return errors.Wrap(err, "generate.ConfigJson json.Encode")
	}

	return nil
}
