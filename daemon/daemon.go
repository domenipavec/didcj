package daemon

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/matematik7/didcj/inventory/server"
	"github.com/matematik7/didcj/runner"
	"github.com/matematik7/didcj/utils"
	"github.com/pkg/errors"
)

type Daemon struct {
	servers []*server.Server
	nodeid  int
	runner  *runner.Runner
}

func New() *Daemon {
	return &Daemon{
		runner: runner.New(),
	}
}

func (d *Daemon) Init() error {
	err := d.runner.Init()
	if err != nil {
		return errors.Wrap(err, "daemon.init")
	}

	d.nodeid, err = utils.Nodeid()
	if err != nil {
		return errors.Wrap(err, "daemon.init")
	}

	d.servers, err = utils.ServerList()
	if err != nil {
		return errors.Wrap(err, "daemon.init")
	}

	r := mux.NewRouter()
	r.HandleFunc("/distribute/{filename}/", d.DistributeHandler)
	r.HandleFunc("/run/", d.RunHandler)
	r.HandleFunc("/start/", d.StartHandler)
	r.HandleFunc("/stop/", d.StopHandler)
	r.HandleFunc("/status/", d.StatusHandler)
	r.HandleFunc("/messages/", d.MessagesHandler)
	http.Handle("/", r)
	return nil
}

func (d *Daemon) StartHandler(w http.ResponseWriter, request *http.Request) {
	d.runner.Start()
}

func (d *Daemon) StopHandler(w http.ResponseWriter, request *http.Request) {
	d.runner.Stop()
}

func (d *Daemon) StatusHandler(w http.ResponseWriter, request *http.Request) {
	err := json.NewEncoder(w).Encode(d.runner.Status())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (d *Daemon) MessagesHandler(w http.ResponseWriter, request *http.Request) {
	err := json.NewEncoder(w).Encode(d.runner.Messages())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
