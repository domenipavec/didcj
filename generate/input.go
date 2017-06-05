package generate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/matematik7/didcj/config"
	"github.com/pkg/errors"
)

const inputh = `
#include <chrono>
#include <random>

std::random_device rd;
std::mt19937 gen(rd());

%s
`

const function = `
%s %s(%s) {
	%s result;
	std::chrono::high_resolution_clock::time_point startTime(std::chrono::high_resolution_clock::now());
%s
	while (std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::high_resolution_clock::now() - startTime).count() < %d);
	return result;
}
`

var typeMap = map[string]string{
	"int64":  "long long",
	"uint64": "unsigned long long",
	"int32":  "long",
	"uint32": "unsigned long",
	"int16":  "short",
	"uint16": "unsigned short",
	"int8":   "char",
	"uint8":  "unsigned char",
}

type returnGenerator func(rawConfig json.RawMessage, typ string) (string, error)

var returnGenerators = map[string]returnGenerator{
	"CONSTANT":                constantGenerator,
	"RANDOM_RANGE":            randomRangeGenerator,
	"INCREASING_RANDOM_RANGE": increasingRandomRangeGenerator,
	"DECREASING_RANDOM_RANGE": decreasingRandomRangeGenerator,
	"RANDOM_LIST":             randomListGenerator,
}

const constantFunction = `
    result = %s;
`

type constantConfig struct {
	Value string `json:"value"`
}

func constantGenerator(rawConfig json.RawMessage, typ string) (string, error) {
	config := constantConfig{}
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(constantFunction, config.Value), nil
}

const randomRangeFunction = `
    std::uniform_int_distribution<%s> dis(%v, %v);
    result = dis(gen);
`

type randomRangeConfig struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

func randomRangeGenerator(rawConfig json.RawMessage, typ string) (string, error) {
	config := randomRangeConfig{}
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(randomRangeFunction, typ, config.Min, config.Max-1), nil
}

const increasingRandomRangeFunction = `
	static const uint64_t window = %v/NumberOfNodes();
    std::uniform_int_distribution<%s> dis(%v + i0 * window, %v + (i0 + 1) * window - 1);
    result = dis(gen);
`

func increasingRandomRangeGenerator(rawConfig json.RawMessage, typ string) (string, error) {
	config := randomRangeConfig{}
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		increasingRandomRangeFunction,
		config.Max-config.Min,
		typ,
		config.Min,
		config.Min,
	), nil
}

const decreasingRandomRangeFunction = `
	static const uint64_t window = %v/NumberOfNodes();
    std::uniform_int_distribution<%s> dis(%v - (i0 + 1) * window, %v - (i0 * window) - 1);
    result = dis(gen);
`

func decreasingRandomRangeGenerator(rawConfig json.RawMessage, typ string) (string, error) {
	config := randomRangeConfig{}
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		decreasingRandomRangeFunction,
		config.Max-config.Min,
		typ,
		config.Max,
		config.Max,
	), nil
}

const randomListFunction = `
    const %s values[] = {%s};
    std::uniform_int_distribution<unsigned long long> dis(0, %v);
    result = values[dis(gen)];
`

type randomListConfig struct {
	Values []string `json:"values"`
}

func randomListGenerator(rawConfig json.RawMessage, typ string) (string, error) {
	config := randomListConfig{}
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return "", err
	}
	values := ""
	for _, value := range config.Values {
		if values != "" {
			values += ", "
		}
		values += "(" + typ + ")(" + value + ")"
	}
	return fmt.Sprintf(randomListFunction,
		typ,
		values,
		len(config.Values)-1,
	), nil
}

func InputH(basename string, inputs []config.Input) error {
	inputFuncs := ""
	for _, input := range inputs {
		str, err := formatInput(input)
		if err != nil {
			return errors.Wrap(err, input.Name)
		}
		inputFuncs += str
	}

	f, err := os.Create(basename + ".h")
	if err != nil {
		return errors.Wrap(err, "generate.InputH file create")
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, inputh, inputFuncs)
	if err != nil {
		return errors.Wrap(err, "generate.InputH fprintf")
	}

	return nil
}

func formatInput(input config.Input) (string, error) {
	inputs := ""
	for i, typ := range input.Inputs {
		name := fmt.Sprintf("i%d", i)
		if inputs != "" {
			inputs += ", "
		}
		inputs += typeMap[typ] + " " + name
	}
	returnType := typeMap[input.ReturnType]
	code, err := returnGenerators[input.ReturnGenerator](input.ReturnConfig, returnType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(function,
		returnType,
		input.Name,
		inputs,
		returnType,
		code,
		input.DurationNs,
	), nil
}
