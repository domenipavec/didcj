package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/matematik7/didcj/inventory/server"
	"github.com/pkg/errors"
)

const Port = "3333"

type Daemon struct {
	homedir string
	servers []*server.Server
	nodeid  int
}

func New() *Daemon {
	return &Daemon{}
}

func (d *Daemon) DistributeHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	filename := vars["filename"]

	query := request.URL.Query()

	mode := os.FileMode(0644)
	if query.Get("exec") != "" {
		mode = os.FileMode(0755)
	}
	file, err := os.OpenFile(path.Join(d.homedir, filename), os.O_RDWR|os.O_CREATE, mode)
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
			Port,
			filename,
			values.Encode(),
		)

		// use pipe to only send data to one receiver at a time
		pipeReader, pipeWriter := io.Pipe()

		sends += 1
		go func() {
			_, err := http.Post(url, "application/data", pipeReader)
			pipeReader.Close()
			responseChan <- err
		}()

		file, err := os.Open(path.Join(d.homedir, filename))
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

func (d *Daemon) Init() error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "daemon.init user")
	}

	d.homedir = usr.HomeDir

	nodeidFile, err := os.Open(path.Join(d.homedir, "nodeid"))
	if err != nil {
		return errors.Wrap(err, "daemon.init nodeid")
	}
	_, err = fmt.Fscanf(nodeidFile, "%d", &d.nodeid)
	nodeidFile.Close()

	serversFile, err := os.Open(path.Join(d.homedir, "servers.json"))
	if err != nil {
		return errors.Wrap(err, "daemon.init servers.json open")
	}
	serversDecoder := json.NewDecoder(serversFile)
	err = serversDecoder.Decode(&d.servers)
	if err != nil {
		return errors.Wrap(err, "daemon.init servers.json decode")
	}
	serversFile.Close()

	r := mux.NewRouter()
	r.HandleFunc("/distribute/{filename}/", d.DistributeHandler)
	http.Handle("/", r)
	return nil
}
