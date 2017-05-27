package daemon

import (
	"encoding/json"
	"fmt"
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
		http.Error(w, fmt.Sprintf("could not decode config: %v", err), 500)
	}
	request.Body.Close()

	err = utils.SendAll(cfg.Servers, "/start/", cfg, nil, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not start: %v", err), 500)
		return
	}

	report := &RunReport{
		Status:  runner.DONE,
		Reports: make([]models.Report, cfg.NumberOfNodes),
	}

	statuses := make([]int, cfg.NumberOfNodes)
	done := false
	for !done {
		time.Sleep(time.Millisecond * 250)

		err := utils.SendAll(cfg.Servers, "/status/", nil, statuses, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get status: %v", err), 500)
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
			err := utils.SendAll(cfg.Servers, "/stop/", nil, nil, true)
			if err != nil {
				http.Error(w, fmt.Sprintf("could not stop: %v", err), 500)
			}
		}
	}

	err = utils.SendAll(cfg.Servers, "/report/", nil, report.Reports, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get report: %v", err), 500)
	}

	file, err := utils.FindFileBasename("app")
	if err != nil {
		http.Error(w, fmt.Sprintf("could not find app file: %v", err), 500)
	}
	err = utils.SendAll(cfg.Servers, fmt.Sprintf("/delete/%s.app/", file), nil, nil, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not delete app: %v", err), 500)
	}

	err = json.NewEncoder(w).Encode(report)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not json encode report: %v", err), 500)
		return
	}
}
