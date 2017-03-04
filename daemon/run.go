package daemon

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/matematik7/didcj/config"
	"github.com/matematik7/didcj/models"
	"github.com/matematik7/didcj/runner"
	"github.com/matematik7/didcj/utils"
)

type RunReport struct {
	Status  int
	Reports []models.Report
}

func (d *Daemon) RunHandler(w http.ResponseWriter, request *http.Request) {
	cfg := &config.Config{}
	err := json.NewDecoder(request.Body).Decode(cfg)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	request.Body.Close()

	err = utils.SendAll(d.servers, "/start/", cfg, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	report := &RunReport{
		Status:  runner.DONE,
		Reports: make([]models.Report, len(d.servers)),
	}

	statuses := make([]int, len(d.servers))
	done := false
	for !done {
		time.Sleep(time.Millisecond * 250)

		err := utils.SendAll(d.servers, "/status/", nil, statuses)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		done = true
		report.Status = runner.DONE
		for _, status := range statuses {
			if status == runner.RUNNING {
				done = false
			} else if status == runner.ERROR {
				report.Status = runner.ERROR
			}
		}

		if !done && report.Status == runner.ERROR {
			err := utils.SendAll(d.servers, "/stop/", nil, nil)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
		}
	}

	err = utils.SendAll(d.servers, "/report/", nil, report.Reports)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	err = json.NewEncoder(w).Encode(report)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
