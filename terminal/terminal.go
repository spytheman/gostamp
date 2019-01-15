package terminal

import (
	"fmt"
	"io"
	"os"
	"sync"
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

// terminalColorCodeBG - returns the closest matching terminal background color escape sequence
func terminalColorCodeBG(r, g, b uint8) string {
	return fmt.Sprintf("\033[48;5%dm", terminalColor2terminalCode(r, g, b))
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
	tColorStdOut = terminalColorCodeReset + terminalColorCodeFG(128, 255, 128) + terminalColorCodeBG(0, 0, 0)
	tColorStdErr = terminalColorCodeReset + terminalColorCodeFG(255, 0, 0) + terminalColorCodeBG(0, 0, 0)
	tColorLineEnd = terminalColorCodeReset
}

var absoluteTimestamps = false

func TurnOnAbsoluteTimestamps() {
	absoluteTimestamps = true
}

var t time.Time
var mutex = &sync.Mutex{}

func init() {
	TurnOnColor()
}

var combineStderrAndStdout = false

func TurnOnCombineStderrAndStdout() {
	combineStderrAndStdout = true
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
	mutex.Lock()
	if absoluteTimestamps {
		now := time.Now()
		fmt.Fprintf(stream, "%s[%04d-%02d-%02d %02d:%02d:%02d.%06d]%s %s\n",
			tCode,
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			now.Nanosecond()/1000,
			tColorLineEnd,
			s)
	} else {
		if t.IsZero() {
			t = time.Now()
		}
		fmt.Fprintf(stream, "%s[%12s]%s %s\n",
			tCode,
			time.Since(t).Round(time.Microsecond).String(),
			tColorLineEnd,
			s)
		t = time.Now()
	}
	mutex.Unlock()
}
