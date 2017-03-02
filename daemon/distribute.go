package daemon

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/matematik7/didcj/config"
)

func (d *Daemon) DistributeHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	filename := vars["filename"]

	query := request.URL.Query()

	mode := os.FileMode(0644)
	if query.Get("exec") != "" {
		mode = os.FileMode(0755)
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, mode)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = io.Copy(file, request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	file.Close()

	destinationsStr := query["dest"]
	var destinations []int
	if len(destinationsStr) > 0 {
		for _, dest := range destinationsStr {
			destination, err := strconv.Atoi(dest)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if destination != d.nodeid {
				destinations = append(destinations, destination)
			}
		}
	} else {
		for destination := range d.servers {
			if destination != d.nodeid {
				destinations = append(destinations, destination)
			}
		}
	}

	sends := 0
	responseChan := make(chan error)

	for len(destinations) > 0 {
		destId := len(destinations) / 2

		values := url.Values{}
		values.Add("exec", query.Get("exec"))
		for _, destination := range destinations[destId:] {
			values.Add("dest", strconv.Itoa(destination))
		}

		url := fmt.Sprintf("http://%s:%s/distribute/%s/?%s",
			d.servers[destinations[destId]].Ip.String(),
			config.DaemonPort,
			filename,
			values.Encode(),
		)

		// use pipe to only send data to one receiver at a time
		pipeReader, pipeWriter := io.Pipe()

		sends += 1
		go func() {
			_, err := http.Post(url, "application/data", pipeReader)

			defer func() { recover() }()
			responseChan <- err
		}()

		file, err := os.Open(filename)
		if err != nil {
			http.Error(w, err.Error(), 500)
			close(responseChan)
			return
		}
		_, err = io.Copy(pipeWriter, file)
		if err != nil {
			http.Error(w, err.Error(), 500)
			close(responseChan)
			return
		}
		pipeWriter.Close()
		file.Close()

		destinations = destinations[:destId]
	}

	for i := 0; i < sends; i++ {
		err := <-responseChan
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
