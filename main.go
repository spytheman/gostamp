package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type programsettings struct {
	showVersion bool
}

var (
	version  = "0.1"
	cmdline = ""
	settings programsettings
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "gostamp - Timestamp and colorize the stdout and stderr streams of CLI programs.\n", )
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] program [programoptions] \n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  The options are:\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&settings.showVersion, "version", false, "show the tool version")
	flag.Parse()
	if settings.showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	if 0 == flag.NArg() {
		flag.Usage()
		os.Exit(1)
	}
	cmdline = strings.Join(flag.Args(), " ")
	fmt.Printf("Running command: '%s' ...\n", cmdline)
}
