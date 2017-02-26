package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/matematik7/didcj/inventory/server"
	"github.com/pkg/errors"
)

func FindFileBasename(extension string) (string, error) {
	cppFiles, err := filepath.Glob("*." + extension)
	if err != nil {
		return "", errors.Wrap(err, "FindFileBasename glob")
	}

	if len(cppFiles) < 1 {
		return "", fmt.Errorf("No .%s files found!", extension)
	}

	if len(cppFiles) > 1 {
		return "", fmt.Errorf("More than 1 .%s file found!", extension)
	}

	return strings.TrimSuffix(cppFiles[0], "."+extension), nil
}

func Upload(srcFile, destFile string, servers ...*server.Server) error {
	scpCmds := make([]*exec.Cmd, 0, len(servers))
	for _, server := range servers {
		scpCmd := exec.Command(
			"sshpass",
			"-p",
			server.Password,
			"scp",
			"-o",
			"StrictHostKeyChecking=no",
			srcFile,
			fmt.Sprintf("%s@%s:~/%s", server.Username, server.Ip.String(), destFile),
		)

		scpCmd.Stdout = os.Stdout
		scpCmd.Stderr = os.Stderr

		err := scpCmd.Start()
		if err != nil {
			return errors.Wrap(err, "Upload start")
		}

		scpCmds = append(scpCmds, scpCmd)
	}
	for _, scpCmd := range scpCmds {
		err := scpCmd.Wait()
		if err != nil {
			return errors.Wrap(err, "Upload wait")
		}
	}
	return nil
}

func Json2File(prefix string, v interface{}) (string, error) {
	tmpJsonFile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", errors.Wrap(err, "Json2File temp file")
	}

	encoder := json.NewEncoder(tmpJsonFile)
	err = encoder.Encode(v)
	if err != nil {
		return "", errors.Wrap(err, "Json2File json encode")
	}
	tmpJsonFile.Close()

	return tmpJsonFile.Name(), nil
}
