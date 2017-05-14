package generate

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func MainCpp(filename, getn string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "generate.MainCpp file create")
	}
	defer f.Close()

	basename := strings.TrimSuffix(filename, ".cpp")

	_, err = fmt.Fprintf(f, templates.String("main.cpp"), basename, getn)
	if err != nil {
		return errors.Wrap(err, "generate.MainCpp fprintf")
	}

	return nil
}
