package borg

import (
	"fmt"
	"os"
	"os/exec"

	"codeberg.org/jstover/borgdrone/internal/logger"
	"github.com/go-cmd/cmd"
)

func assertExists() {
	_, err := exec.LookPath("borg")
	if err != nil {
		logger.Fatal("borg command was not found or is not installed.", 1)
	}
}

func Run(env []string, args ...string) bool {
	assertExists()
	opts := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	c := cmd.NewCmdOptions(opts, "borg", args...)
	c.Env = env

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for c.Stdout != nil || c.Stderr != nil {
			select {
			case line, open := <-c.Stdout:
				if !open {
					c.Stdout = nil
					continue
				}
				logger.Debug(line)
			case line, open := <-c.Stderr:
				if !open {
					c.Stderr = nil
					continue
				}
				logger.Debug(line)
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	<-c.Start()
	<-doneChan

	status := c.Status()
	if status.Error != nil {
		logger.Fatal(status.Error.Error(), 2)
	}

	return status.Exit == 0

}
