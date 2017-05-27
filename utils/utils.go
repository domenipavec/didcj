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

var SSHParams = []string{
	"-o",
	"UserKnownHostsFile=/dev/null",
	"-o",
	"StrictHostKeyChecking=no",
	"-o",
	"LogLevel=ERROR",
	"-o",
	"ForwardAgent=yes",
}

func FindFileBasename(extension string) (string, error) {
	cppFiles, err := filepath.Glob("*." + extension)
	if err != nil {
		return "", errors.Wrap(err, "FindFileBasename glob")
	}

	if len(cppFiles) < 1 {
		return "", fmt.Errorf("no .%s files found", extension)
	}

	if len(cppFiles) > 1 {
		return "", fmt.Errorf("more than 1 .%s file found", extension)
	}

	return strings.TrimSuffix(cppFiles[0], "."+extension), nil
}

func Upload(srcFile, destFile string, servers ...*models.Server) error {
	if len(servers) == 0 {
		return nil
	}

	allParams := append(SSHParams,
		"-C", // compression
		"-q", // no progress bar
		srcFile,
		fmt.Sprintf("%s@%s:~/%s", servers[0].Username, servers[0].IP.String(), destFile),
	)
	scpCmd := exec.Command(
		"scp",
		allParams...,
	)

	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr

	err := scpCmd.Run()
	if err != nil {
		return errors.Wrap(err, "could not upload")
	}

	if len(servers) > 1 {
		allUploads := []string{}
		for _, server := range servers[1:] {
			allUploads = append(allUploads, "scp")
			allUploads = append(allUploads, SSHParams...)
			allUploads = append(allUploads,
				"-q",
				destFile,
				fmt.Sprintf("%s@%s:~/%s", server.Username, server.PrivateIP.String(), destFile),
				"&",
			)
		}
		allUploads = append(allUploads, "wait")
		Run(servers[:1], allUploads...)
	}

	return nil
}

func Run(servers []*models.Server, params ...string) error {
	sshCmds := make([]*exec.Cmd, 0, len(servers))
	for _, server := range servers {
		allParams := append(SSHParams,
			fmt.Sprintf("%s@%s", server.Username, server.IP.String()),
		)
		allParams = append(allParams, params...)

		cmd := exec.Command(
			"ssh",
			allParams...,
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			return errors.Wrap(err, "Run start")
		}

		sshCmds = append(sshCmds, cmd)
	}
	for _, cmd := range sshCmds {
		err := cmd.Wait()
		if err != nil {
			return errors.Wrap(err, "Run wait")
		}
	}
	return nil
}

func Send(destServer *models.Server, path string, input interface{}, output interface{}, private ...bool) error {
	var err error
	var response *http.Response

	ip := destServer.IP.String()
	if len(private) > 0 {
		ip = destServer.PrivateIP.String()
	}

	url := fmt.Sprintf("http://%s:%s%s", ip, config.DaemonPort, path)

	var body io.Reader
	if input != nil {
		if inputReader, ok := input.(io.Reader); ok {
			body = inputReader
		} else {
			buf := &bytes.Buffer{}
			err = json.NewEncoder(buf).Encode(input)
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
		var errMsg []byte
		errMsg, err = ioutil.ReadAll(response.Body)
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
			return errors.Wrap(err, server.Name)
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

func FormatDuration(ns int64) string {
	if ns > 1000*1000*1000 {
		return fmt.Sprintf("%.1f s", float64(ns)/(1000*1000*1000))
	}
	return fmt.Sprintf("%.1f ms", float64(ns)/(1000*1000))
}

func FormatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f kB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
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
