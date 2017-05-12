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

	servers := d.servers[:cfg.NumberOfNodes]

	err = utils.SendAll(servers, "/start/", cfg, nil, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	report := &RunReport{
		Status:  runner.DONE,
		Reports: make([]models.Report, len(servers)),
	}

	statuses := make([]int, len(servers))
	done := false
	for !done {
		time.Sleep(time.Millisecond * 250)

		err := utils.SendAll(servers, "/status/", nil, statuses, true)
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
			err := utils.SendAll(servers, "/stop/", nil, nil, true)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
		}
	}

	err = utils.SendAll(servers, "/report/", nil, report.Reports, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	err = json.NewEncoder(w).Encode(report)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
