package generate

import (
	"fmt"
	"os"

	"github.com/matematik7/didcj/templates"
	"github.com/pkg/errors"
)

func MessageH(numberOfNodes int) error {
	f, err := os.Create("message.h")
	if err != nil {
		return errors.Wrap(err, "GenerateMessageH file create")
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, templates.Box.String("message.h"), numberOfNodes)
	if err != nil {
		return errors.Wrap(err, "GenerateMessageH fprintf")
	}

	return nil
}
