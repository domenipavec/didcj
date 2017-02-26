package daemon

import (
	"io"
	"net/http"
	"os"
	"os/user"
	"path"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Daemon struct {
	homedir string
}

func New() *Daemon {
	return &Daemon{}
}

func (d *Daemon) DistributeHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	filename := vars["filename"]

	mode := os.FileMode(0644)
	if request.URL.Query().Get("exec") != "" {
		mode = os.FileMode(0755)
	}
	file, err := os.OpenFile(path.Join(d.homedir, filename), os.O_RDWR|os.O_CREATE, mode)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	defer file.Close()

	_, err = io.Copy(file, request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (d *Daemon) Init() error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "daemon.init user")
	}

	d.homedir = usr.HomeDir

	r := mux.NewRouter()
	r.HandleFunc("/distribute/{filename}/", d.DistributeHandler)
	http.Handle("/", r)
	return nil
}
