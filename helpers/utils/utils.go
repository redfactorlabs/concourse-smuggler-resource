package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/colorstring"
)

type TempFileLogger struct {
	logFile *os.File
	Logger  *log.Logger
}

func NewTempFileLogger(path string) (*TempFileLogger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	l := log.New(f, "", log.Ldate|log.Ltime)
	t := &TempFileLogger{
		logFile: f,
		Logger:  l,
	}
	return t, nil
}

func (t *TempFileLogger) Close() {
	t.logFile.Close()
}

func GetEnvOrDefault(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	} else {
		return defaultValue
	}
}

func Fatal(doing string, err error, exitCode int) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	os.Exit(exitCode)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}
