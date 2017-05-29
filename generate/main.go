package generate

import (
	"fmt"
	"os"
	"strings"

	"github.com/matematik7/didcj/templates"
	"github.com/pkg/errors"
)

func MainDcj(filename, getn string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "generate.MainDcj file create")
	}
	defer f.Close()

	basename := strings.TrimSuffix(filename, ".dcj")

	_, err = fmt.Fprintf(f, templates.Box.String("main.dcj"), basename, getn)
	if err != nil {
		return errors.Wrap(err, "generate.MainDcj fprintf")
	}

	return nil
}
