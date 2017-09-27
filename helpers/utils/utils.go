package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mitchellh/colorstring"
)

type TempFileLogger struct {
	logFile *os.File
	Logger  *log.Logger
}

func NewTempFileLogger(path string) (*TempFileLogger, error) {
	var l *log.Logger
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	l = log.New(f, "", log.Ldate|log.Ltime)
	t := &TempFileLogger{
		logFile: f,
		Logger:  l,
	}
	return t, nil
}

func (t *TempFileLogger) DupToStderr() {
	t.Logger = log.New(io.MultiWriter(os.Stderr, t.logFile), "", log.Ldate|log.Ltime)
}

func (t *TempFileLogger) SendToStderr() {
	t.logFile.Close()
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

// List the json tag names (`json:"name,opts"`)   of a struct
func ListJsonTagsOfStruct(x interface{}) []string {
	v := reflect.TypeOf(x)
	tags := make([]string, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		t := v.Field(i).Tag.Get("json")
		t = strings.SplitN(t, ",", 2)[0]
		tags = append(tags, t)
	}
	return tags
}

// Removes the keys in a map that match the json tag names (`json:"name,opts"`)
// for the given struct value
func FilterMapFromJsonStruct(m map[string]interface{}, x interface{}) {
	for _, t := range ListJsonTagsOfStruct(x) {
		delete(m, t)
	}
}

func InterfaceToMap(v interface{}) (map[string]interface{}, error) {
	switch v.(type) {
	case map[string]interface{}:
		return v.(map[string]interface{}), nil
	default:
		return nil, fmt.Errorf("The value %+v is not a map", v)
	}
}

// If a and b are a maps, merge a into b, not overriding
func MergeMaps(a, b interface{}) (interface{}, error) {
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}
	ma, err := InterfaceToMap(a)
	if err != nil {
		return nil, err
	}
	mb, err := InterfaceToMap(b)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for k, v := range ma {
		m[k] = v
	}
	for k, v := range mb {
		if m[k] == nil {
			m[k] = v
		}
	}
	return m, nil
}

func JsonPrettyPrint(in []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "  ")
	if err != nil {
		return in
	}
	return out.Bytes()
}
