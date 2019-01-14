package main

import (
	"flag"
	"os"
	"fmt"
)

type Settings struct {
	showVersion bool
}

var (
	version  string = "0.1"
	settings Settings
)

func init() {
	flag.BoolVar(&settings.showVersion, "version", false, "show the tool version")
	flag.Parse()
	if settings.showVersion {
		fmt.Println(version)
		os.Exit(0)		
	}
		
}

func main() {
	fmt.Println("Bye")
}
