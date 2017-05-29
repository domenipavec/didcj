package compile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDcj(t *testing.T) {
	files, err := filepath.Glob("../templates/tests/*.dcj")
	assert.NoError(t, err)

	for _, file := range files {
		file := strings.TrimSuffix(file, ".dcj")
		t.Logf("Testing %s", file)
		t.Run(file, func(t *testing.T) {
			err := Transpile(file)
			assert.NoError(t, err, "could not transpile")

			err = Compile(file)
			assert.NoError(t, err, "could not compile")

			cmd := exec.Command(file + ".app")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			assert.NoError(t, err, "test failed")

			err = os.Remove(file + ".app")
			assert.NoError(t, err, "could not remove app file")

			err = os.Remove(file + ".cpp")
			assert.NoError(t, err, "could not remove cpp file")
		})
	}
}
