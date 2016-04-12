package utils

import (
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
)

func Fatal(doing string, err error, exitCode int) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	os.Exit(exitCode)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}
