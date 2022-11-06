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
	useCsv      bool
	showStart   bool
	showEnd     bool
	mergeErr    bool
	useElapsed  bool
	microSecond bool
	nobuffering bool
}

var (
	version  string
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
	flag.BoolVar(&settings.useCsv, "csv", false, "do not format the output at all, just show the time in ns, followed by ',' then the output")
	flag.BoolVar(&settings.useAbsolute, "absolute", false, "use absolute timestamps")
	flag.BoolVar(&settings.showStart, "start", true, "timestamp the start of the execution")
	flag.BoolVar(&settings.showEnd, "end", true, "timestamp the end of the execution")
	flag.BoolVar(&settings.mergeErr, "merge", false, "merge stderr to stdout. Useful for later filtering with grep.")
	flag.BoolVar(&settings.useElapsed, "elapsed", false, "use timestamps, showing the elapsed time from the start of the program. Can not be used with -absolute")
	flag.BoolVar(&settings.microSecond, "micro", false, "round timestamps to microseconds, instead of milliseconds. Can not be used with -absolute")
	flag.BoolVar(&settings.nobuffering, "nobuf", false, "run the program with stdbuf -i0 -oL -eL, i.e. with *buffering off* for the std streams")
	flag.Parse()
	//fmt.Println(settings)
	if settings.showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	if !settings.useColor {
		terminal.TurnOffColor()
	}
	if settings.useCsv {
		terminal.TurnOnCsv()
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
	if settings.useAbsolute && settings.microSecond {
		fmt.Fprintf(os.Stderr, "-absolute and -micro can not be used together.\n")
		os.Exit(-1)
	}

	if settings.mergeErr {
		terminal.TurnOnCombineStderrAndStdout()
	}
	if settings.microSecond {
		terminal.TurnOnMicrosecondTimestampResolution()
	}
}

func main() {
	if 0 == flag.NArg() {
		flag.Usage()
		os.Exit(1)
	}

	var cmd_args []string
	if settings.nobuffering {
		cmd_args = append(cmd_args, []string{"stdbuf", "-i0", "-oL", "-eL"}...)
	}
	cmd_args = append(cmd_args, flag.Args()...)

	cmdline = strings.Join(cmd_args, " ")
	//fmt.Printf("Running command: '%s' ...\n", cmdline)

	command := exec.Command(cmd_args[0], cmd_args[1:]...)

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
		terminal.ResetPreviousTerminalLineTime()
		terminal.Out("Start of '" + cmdline + "'")
	}

	// Setup is finished at this point. Run the command and process the results:
	startErr := command.Start()
	if startErr != nil {
		terminal.Err("-->could not start, because of error: " + startErr.Error())
		defer os.Exit(1)
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

	wg.Wait()

	if settings.showEnd {
		terminal.Out("End of '" + cmdline + "'")
	}

	terminal.Shutdown()
}
