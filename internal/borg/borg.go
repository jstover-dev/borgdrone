package borg

import (
	"fmt"
	"os"

	"github.com/go-cmd/cmd"
)

type Environment struct {
	PassCommand             string
	RelocatedRepoAccessIsOk bool
	Repo                    string
	Rsh                     string
}

func Run(args []string) {
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	envCmd := cmd.NewCmdOptions(cmdOptions, "./counter", args...)

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for envCmd.Stdout != nil || envCmd.Stderr != nil {
			select {
			case line, open := <-envCmd.Stdout:
				if !open {
					envCmd.Stdout = nil
					continue
				}
				fmt.Println(line)
			case line, open := <-envCmd.Stderr:
				if !open {
					envCmd.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	<-envCmd.Start()
	<-doneChan
}
