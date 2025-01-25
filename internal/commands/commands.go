package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"codeberg.org/jstover/borgdrone/internal/borg"
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/types"
	"gopkg.in/yaml.v3"
)

func ListTargets(cfg config.Config, format string) {
	fmt.Println("Running ListTargets")
	fmt.Printf("Format = %s\n", format)

	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg.TargetMap, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))

	case "yaml":
		data, err := yaml.Marshal(cfg.TargetMap)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))

	case "text":
	}

}

func Initialise(target types.BorgTarget) {
	fmt.Println("Runnning Initialise")
	fmt.Printf("Target = %+v\n", target)
	borg.Run([]string{})
}

func Info(target types.BorgTarget) {
	fmt.Println("Running Info")
	fmt.Printf("Target = %+v\n", target)
}

func List(target types.BorgTarget) {
	fmt.Println("Running List")
	fmt.Printf("Target = %+v\n", target)
}

func Create(target types.BorgTarget) {
	fmt.Println("Running Create")
	fmt.Printf("Target = %+v\n", target)
}

func ExportKey(target types.BorgTarget) {
	fmt.Println("Running ExportKey")
	fmt.Printf("Target = %+v\n", target)
}

func ImportKey(target types.BorgTarget, keyFile string, passwordFile string) {
	fmt.Println("Running ImportKey")
	fmt.Printf("Target = %+v\n", target)
	fmt.Printf("Key File = %s\n", keyFile)
	fmt.Printf("Password File = %s\n", passwordFile)
}

func Clean() {
	fmt.Println("Running Clean...")
}
