package main

import (
	"flag"
	"fmt"
	"os"
)

type programsettings struct {
	showVersion bool
}

var (
	version  = "0.1"
	settings programsettings
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
