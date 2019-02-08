package terminal

import (
	"fmt"
	"io"
	"os"
	"time"
)

func terminalColor2terminalCode(r, g, b uint8) int {
	rr := (int(r) * 5) / 0xFF
	gg := (int(g) * 5) / 0xFF
	bb := (int(b) * 5) / 0xFF
	return 36*rr + 6*gg + bb + 16
}

// terminalColorCodeFG - returns the closest matching terminal foreground color escape sequence
func terminalColorCodeFG(r, g, b uint8) string {
	return fmt.Sprintf("\033[38;5;%dm", terminalColor2terminalCode(r, g, b))
}

// This escape code resets the terminal foreground and background colors
const terminalColorCodeReset = "\033[0;00m"

var tColorStdOut = ""
var tColorStdErr = ""
var tColorLineEnd = ""

func TurnOffColor() {
	tColorStdErr = "stderr: "
	tColorStdOut = "stdout: "
	tColorLineEnd = ""
}

func TurnOnColor() {
	tColorStdOut = terminalColorCodeReset + terminalColorCodeFG(128, 255, 128) 
	tColorStdErr = terminalColorCodeReset + terminalColorCodeFG(255, 0, 0) 
	tColorLineEnd = terminalColorCodeReset
}

var absoluteTimestamps = false

func TurnOnAbsoluteTimestamps() {
	absoluteTimestamps = true
}

var timestampRoundResolution = time.Millisecond

func TurnOnMicrosecondTimestampResolution() {
	timestampRoundResolution = time.Microsecond
}

var previousTerminalLineTime time.Time

type terminalLine struct {
	timestamp time.Time
	stream    io.Writer
	tCode     string
	s         string
}

func newTerminalLine(stream io.Writer, tCode string, s string) terminalLine {
	return terminalLine{timestamp: time.Now(), stream: stream, tCode: tCode, s: s}
}

var tLines = make(chan terminalLine, 10)
var allTerminalLinesAreFlushed = make(chan bool)

func init() {
	TurnOnColor()
	go func() {
		ResetPreviousTerminalLineTime()
		c := 0
		for line := range tLines {
			writeTerminalLine(line)
			c++
		}
		allTerminalLinesAreFlushed <- true
	}()
}

func ResetPreviousTerminalLineTime() {
	previousTerminalLineTime = time.Now()
}

func Shutdown() {
	close(tLines)
	<-allTerminalLinesAreFlushed // blocks till the tLines channel is drained
}

var combineStderrAndStdout = false

func TurnOnCombineStderrAndStdout() {
	combineStderrAndStdout = true
}

var timeRelativeToStart = false

func TurnOnTimeRelativeToStart() {
	timeRelativeToStart = true
}

func Out(s string) {
	lineOut(os.Stdout, tColorStdOut, s)
}

func Err(s string) {
	if combineStderrAndStdout {
		lineOut(os.Stdout, tColorStdErr, s)
	} else {
		lineOut(os.Stderr, tColorStdErr, s)
	}
}

func lineOut(stream io.Writer, tCode string, s string) {
	tLines <- newTerminalLine(stream, tCode, s)
}

func writeTerminalLine(tLine terminalLine) {
	if absoluteTimestamps {
		now := tLine.timestamp
		fmt.Fprintf(tLine.stream, "%s[%04d-%02d-%02d %02d:%02d:%02d.%06d]%s %s\n",
			tLine.tCode,
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			now.Nanosecond()/1000,
			tColorLineEnd,
			tLine.s)
	} else {
		fmt.Fprintf(tLine.stream, "%s[%12s]%s %s\n",
			tLine.tCode,
			time.Since(previousTerminalLineTime).Round(timestampRoundResolution).String(),
			tColorLineEnd,
			tLine.s)
		if !timeRelativeToStart {
			previousTerminalLineTime = tLine.timestamp
		}
	}
}
