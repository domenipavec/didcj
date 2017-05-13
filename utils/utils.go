package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/matematik7/didcj/config"
	"github.com/matematik7/didcj/models"
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

func Upload(srcFile, destFile string, servers ...*models.Server) error {
	scpCmds := make([]*exec.Cmd, 0, len(servers))
	for _, server := range servers {
		scpCmd := exec.Command(
			"sshpass",
			"-p",
			server.Password,
			"scp",
			"-C",
			"-o",
			"UserKnownHostsFile=/dev/null",
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

func Send(destServer *models.Server, path string, input interface{}, output interface{}, private ...bool) error {
	var err error
	var response *http.Response

	ip := destServer.Ip.String()
	if len(private) > 0 {
		ip = destServer.PrivateIp.String()
	}

	url := fmt.Sprintf("http://%s:%s%s", ip, config.DaemonPort, path)

	var body io.Reader
	if input != nil {
		if inputReader, ok := input.(io.Reader); ok {
			body = inputReader
		} else {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(input)
			if err != nil {
				return errors.Wrap(err, "post json encode")
			}
			body = buf
		}
		response, err = http.Post(url, "application/json", body)
	} else {
		response, err = http.Get(url)
	}

	if err != nil {
		return errors.Wrap(err, "Post post")
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		errMsg, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "Post read error")
		}
		return fmt.Errorf("Post %d: %s", response.StatusCode, string(errMsg))
	}

	if output != nil {
		err = json.NewDecoder(response.Body).Decode(output)
		if err != nil {
			return errors.Wrap(err, "post json decode")
		}
	}

	return nil
}

func SendAll(servers []*models.Server, path string, input interface{}, outputs interface{}, private ...bool) error {
	errChan := make(chan error)
	for i, srvr := range servers {
		go func(j int, destServer *models.Server) {
			var output interface{}
			if outputs != nil {
				if outints, ok := outputs.([]int); ok {
					output = &outints[j]
				} else if outstrings, ok := outputs.([]string); ok {
					output = &outstrings[j]
				} else if outslicereports, ok := outputs.([]models.Report); ok {
					output = &outslicereports[j]
				}
			}
			err := Send(destServer, path, input, output, private...)
			errChan <- err
		}(i, srvr)
	}

	for _, server := range servers {
		err := <-errChan
		if err != nil {
			return errors.Wrap(err, server.Ip.String())
		}
	}

	return nil
}

func Nodeid() (int, error) {
	nodeidFile, err := os.Open("nodeid")
	if err != nil {
		return 0, errors.Wrap(err, "Runner.Init open nodeid")
	}
	defer nodeidFile.Close()

	var nodeid int
	_, err = fmt.Fscanf(nodeidFile, "%d", &nodeid)
	if err != nil {
		return 0, errors.Wrap(err, "Runner.Init fscanf nodeid")
	}

	return nodeid, nil
}

func ServerList() ([]*models.Server, error) {
	var servers []*models.Server

	serversFile, err := os.Open("servers.json")
	if err != nil {
		return nil, errors.Wrap(err, "serverlist servers.json open")
	}
	defer serversFile.Close()

	serversDecoder := json.NewDecoder(serversFile)
	err = serversDecoder.Decode(&servers)
	if err != nil {
		return nil, errors.Wrap(err, "serverlist servers.json decode")
	}

	return servers, nil
}

func FormatDuration(ns int64) string {
	if ns > 1000*1000*1000 {
		return fmt.Sprintf("%.1f s", float64(ns)/(1000*1000*1000))
	} else {
		return fmt.Sprintf("%.1f ms", float64(ns)/(1000*1000))
	}
}

func FormatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f kB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}

func GetHFileFromDownloads(basefilename string) {
	hFile := basefilename + ".h"

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	downloadedHFile := path.Join(usr.HomeDir, "Downloads", hFile)
	if _, err = os.Stat(downloadedHFile); err == nil {
		log.Printf("Found new %s file at %s\n", hFile, downloadedHFile)
		os.Rename(downloadedHFile, hFile)
	}
}

func GetName(i int) string {
	return fmt.Sprintf("didcj-%03d", i)
}
