package utils

import (
	"fmt"
	"io"
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

func Panic(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}

func PrintRecover() {
	if r := recover(); r != nil {
		Sayf(colorstring.Color("[red]%s\n"), r)
		os.Exit(1)
	}
}

func Fatal(doing string, err error, exitCode int) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	os.Exit(exitCode)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

// Copy files
// via	http://stackoverflow.com/a/21061062/395686
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}
