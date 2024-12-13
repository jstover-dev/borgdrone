package main

import (
	"fmt"
	"log"
	"codeberg.org/jstover/borgdrone"
)

func main(){


    cfg, err := borgdrone.ReadConfigFile("/home/josh/borg-drone.new.yml")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", cfg)

}


