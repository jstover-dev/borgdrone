package main

import (
	"fmt"
	"codeberg.org/jstover/borgdrone/internal/cmdargs"
)


func main() {
	args := cmdargs.ParseArgs()
	fmt.Println(args)

	switch {
	case args.GenerateConfig != nil:
		fmt.Println("Generate Config")

	case args.ListTargets != nil:
		fmt.Println("List Targets")
		fmt.Println(args.ListTargets)

	case args.Initialise != nil:
		fmt.Println("Initialise")
		fmt.Println(args.Initialise)

	case args.Info != nil:
		fmt.Println("Info")
		fmt.Println(args.Info)

	case args.List != nil:
		fmt.Println("List")
		fmt.Println(args.List)

	case args.Create != nil:
		fmt.Println("Create")
		fmt.Println(args.Create)

	case args.ExportKeys != nil:
		fmt.Println("ExportKeys")
		fmt.Println(args.ExportKeys)

	case args.ImportKeys != nil:
		fmt.Println("ImportKeys")
		fmt.Println(args.ImportKeys)

	case args.Clean != nil:
		fmt.Println("Clean")
		fmt.Println(args.Clean)

	}

}
