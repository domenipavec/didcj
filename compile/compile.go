package compile

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/matematik7/didcj/templates"
	"github.com/pkg/errors"
)

var importRegex = regexp.MustCompile("#import *[<\"]([a-zA-Z0-9./]+)[>\"]")

func Compile(file string) error {
	gppCmd := exec.Command("g++", "-std=gnu++0x", "-O2", "-static", "-lm", "-DDIDCJ", "-I.", "-o", file+".app", file+".cpp")
	gppCmd.Stdout = os.Stdout
	gppCmd.Stderr = os.Stderr
	return gppCmd.Run()
}

func Transpile(file string) error {
	dcjFile := file + ".dcj"
	if _, err := os.Stat(dcjFile); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(dcjFile)
	if err != nil {
		return errors.Wrapf(err, "could not read %s", dcjFile)
	}

	for _, match := range importRegex.FindAllSubmatch(data, -1) {
		data = bytes.Replace(data, match[0], templates.Box.Bytes(string(match[1])), 1)
	}

	cppFile := file + ".cpp"
	err = ioutil.WriteFile(cppFile, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "cold not write %s", cppFile)
	}

	return nil
}
