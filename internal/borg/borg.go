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

type Runner struct {
	Env    []string
	Stdout []string
	Stderr []string
}

func (r *Runner) Run(args ...string) bool {
	assertExists()
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	command := cmd.NewCmdOptions(cmdOptions, "borg", args...)
	command.Env = r.Env

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for command.Stdout != nil || command.Stderr != nil {
			select {
			case line, open := <-command.Stdout:
				if !open {
					command.Stdout = nil
					continue
				}
				logger.Debug(line)
				r.Stdout = append(r.Stdout, line)
			case line, open := <-command.Stderr:
				if !open {
					command.Stderr = nil
					continue
				}
				logger.Debug(line)
				fmt.Fprintln(os.Stderr, line)
				r.Stderr = append(r.Stderr, line)
			}
		}
	}()

	<-command.Start()
	<-doneChan

	status := command.Status()
	if status.Error != nil {
		logger.Fatal(status.Error.Error(), 2)
	}

	return status.Exit == 0

}
