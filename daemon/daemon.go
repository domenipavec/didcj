package daemon

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/matematik7/didcj/config"
	"github.com/matematik7/didcj/runner"
	"github.com/pkg/errors"
)

type Daemon struct {
	runner *runner.Runner
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

	r := mux.NewRouter()
	r.HandleFunc("/run/", d.RunHandler)
	r.HandleFunc("/start/", d.StartHandler)
	r.HandleFunc("/stop/", d.StopHandler)
	r.HandleFunc("/status/", d.StatusHandler)
	r.HandleFunc("/report/", d.ReportHandler)
	r.HandleFunc("/delete/{filename}/", d.DeleteHandler)
	http.Handle("/", r)
	return nil
}

func (d *Daemon) DeleteHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	filename := vars["filename"]

	err := os.Remove(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (d *Daemon) StartHandler(w http.ResponseWriter, request *http.Request) {
	cfg := &config.Config{}
	err := json.NewDecoder(request.Body).Decode(cfg)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	request.Body.Close()

	d.runner.Start(cfg)
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

func (d *Daemon) ReportHandler(w http.ResponseWriter, request *http.Request) {
	err := json.NewEncoder(w).Encode(d.runner.Report())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
