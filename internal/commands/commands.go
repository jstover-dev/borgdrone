package commands

import (
	"fmt"

	"codeberg.org/jstover/borgdrone/internal/types"
)

func ListTargets(format string) {
	fmt.Println("Running ListTargets")
	fmt.Printf("Format = %s\n", format)
}

func Initialise(target types.BorgTarget) {
	fmt.Println("Runnning Initialise")
	fmt.Printf("Target = %+v\n", target)
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
