package util

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultErrorExitCode = 1
)

// fatal prints the message (if provided) and then exits. If V(99) or greater,
// klog.Fatal is invoked for extended information. This is intended for maintainer
// debugging and out of a reasonable range for users.
func fatalErrHandler(msg string, code int) {
	// nolint:logcheck // Not using the result of klog.V(99) inside the if
	// branch is okay, we just use it to determine how to terminate.
	//if klog.V(99).Enabled() {
	//	klog.FatalDepth(2, msg)
	//}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}

		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(code)
}

var ErrExit = fmt.Errorf("exit")

func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

// checkErr formats a given error as a string and calls the passed handleErr
// func with that string and an kubectl exit code.
func checkErr(err error, handleErr func(string, int)) {

	if err == nil {
		return
	}

	switch {
	case err == ErrExit:
		handleErr("", DefaultErrorExitCode)
	default:
		switch err := err.(type) {
		default: // for any other error type
			msg, ok := StandardErrorMessage(err)
			if !ok {
				msg = err.Error()
				if !strings.HasPrefix(msg, "error: ") {
					msg = fmt.Sprintf("error: %s", msg)
				}
			}
			handleErr(msg, DefaultErrorExitCode)
		}
	}
}

func StandardErrorMessage(err error) (string, bool) {
	return fmt.Sprintf("error: %s", err), false
}
