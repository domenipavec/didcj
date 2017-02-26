package compile

import (
	"os"
	"os/exec"
)

func Compile(file string) error {
	gppCmd := exec.Command("g++", "-std=gnu++0x", "-O2", "-static", "-lm", "-I.", "-o", file+".app", file+".cpp")
	gppCmd.Stdout = os.Stdout
	gppCmd.Stderr = os.Stderr
	return gppCmd.Run()
}
