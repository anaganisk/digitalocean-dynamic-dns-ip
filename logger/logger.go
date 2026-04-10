package logger

import (
	"io"
	"log"
	"os"
)

var stdErrLogger = log.New(os.Stderr, "", 0) // Logger for errors and warnings to stderr
var stdOutLogger = log.New(os.Stdout, "", 0) // Logger for debug output to stdout

// ExitFunc is the function called when exiting. Can be overridden for testing.
var ExitFunc = os.Exit

// SetDebugOutput sets the destination for debug output. Call with io.Discard to disable debug output.
func SetDebugOutput(w io.Writer) {
	stdOutLogger = log.New(w, "", 0)
}

// CheckError checks if an error occurred and logs it to stderr before exiting the program
func CheckError(err error) {
	if err != nil {
		ErrorAndExit(err.Error())
	}
}

// ErrorAndExit logs an error message to stderr and exits the program
func ErrorAndExit(msg string) {
	stdErrLogger.Println(msg)
	ExitFunc(1)
}

// Warning logs a warning message to stderr without exiting the program
func Warning(msg string) {
	stdErrLogger.Println(msg)
}

// Warningf logs a formatted warning message to stderr without exiting the program
func Warningf(format string, v ...interface{}) {
	stdErrLogger.Printf(format, v...)
}

// Debug logs a debug message to stdout without exiting the program
func Debug(format string, v ...interface{}) {
	stdOutLogger.Printf(format, v...)
}
