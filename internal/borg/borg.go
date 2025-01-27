package borg

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/go-cmd/cmd"
)

func assertExists() {
	_, err := exec.LookPath("borg")
	if err != nil {
		log.Fatal("borg command was not found or is not installed.")
	}
}

func Run(args []string, env []string) {
	assertExists()
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	command := cmd.NewCmdOptions(cmdOptions, "borg", args...)
	command.Env = env

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
				fmt.Println(line)
			case line, open := <-command.Stderr:
				if !open {
					command.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	<-command.Start()
	<-doneChan

	status := command.Status()
	if status.Error != nil {
		log.Fatal(status.Error)
	}
}
