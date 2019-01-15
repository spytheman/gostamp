package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/spytheman/gostamp/terminal"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type programSettings struct {
	showVersion bool
	useColor    bool
	useAbsolute bool
	showStart   bool
	showEnd     bool
	mergeErr    bool
	useElapsed  bool
}

var (
	version  = "0.2"
	cmdline  = ""
	settings programSettings
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gostamp - Timestamp and colorize the stdout and stderr streams of CLI programs.\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] program [programoptions] \n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  The options are:\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&settings.showVersion, "version", false, "show the tool version")
	flag.BoolVar(&settings.useColor, "color", true, "colorize the output")
	flag.BoolVar(&settings.useAbsolute, "absolute", false, "use absolute timestamps")
	flag.BoolVar(&settings.showStart, "start", true, "timestamp the start of the execution")
	flag.BoolVar(&settings.showEnd, "end", true, "timestamp the end of the execution")
	flag.BoolVar(&settings.mergeErr, "merge", false, "merge stderr to stdout. Useful for later filtering with grep.")
	flag.BoolVar(&settings.useElapsed, "elapsed", false, "use timestamps, showing the elapsed time from the start of the program. Can not be used with -absolute")
	flag.Parse()
	//fmt.Println(settings)
	if settings.showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	if !settings.useColor {
		terminal.TurnOffColor()
	}
	if settings.useAbsolute {
		terminal.TurnOnAbsoluteTimestamps()
	}
	if settings.useElapsed {
		terminal.TurnOnTimeRelativeToStart()
	}
	if settings.useAbsolute && settings.useElapsed {
		fmt.Fprintf(os.Stderr, "-absolute and -elapsed can not be used together.\n")
		os.Exit(-1)
	}

	if settings.mergeErr {
		terminal.TurnOnCombineStderrAndStdout()
	}
}

func main() {
	if 0 == flag.NArg() {
		flag.Usage()
		os.Exit(1)
	}
	cmdline = strings.Join(flag.Args(), " ")
	//fmt.Printf("Running command: '%s' ...\n", cmdline)

	command := exec.Command(flag.Args()[0], flag.Args()[1:]...)

	commandIn, commandInErr := command.StdinPipe()
	if commandInErr != nil {
		log.Panic(commandInErr)
	}
	commandOut, commandOutErr := command.StdoutPipe()
	if commandOutErr != nil {
		log.Panic(commandOutErr)
	}
	commandErr, commandErrErr := command.StderrPipe()
	if commandErrErr != nil {
		log.Panic(commandErrErr)
	}

	scannerOut := bufio.NewScanner(commandOut)
	scannerErr := bufio.NewScanner(commandErr)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		_, err := io.Copy(commandIn, os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		_ = commandIn.Close()
	}()

	go func() {
		defer wg.Done()
		for scannerErr.Scan() {
			terminal.Err(scannerErr.Text())
		}
	}()

	go func() {
		defer wg.Done()
		for scannerOut.Scan() {
			terminal.Out(scannerOut.Text())
		}
	}()

	if settings.showStart {
		terminal.Out("Start of '" + cmdline + "'")
	}

	startErr := command.Start()
	if startErr != nil {
		terminal.Err("-->could not start, because of error: " + startErr.Error())
		os.Exit(1)
	}

	waitError := command.Wait()
	if waitError != nil {
		terminal.Err("-->finished with error: " + waitError.Error())
		if exitError, ok := waitError.(*exec.ExitError); ok {
			if exitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
				defer os.Exit(exitStatus.ExitStatus())
			}
		}
	}

	if settings.showEnd {
		terminal.Out("End of '" + cmdline + "'")
	}

	wg.Wait()
}
