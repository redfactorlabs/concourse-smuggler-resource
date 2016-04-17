package test

import (
	"io/ioutil"
)

func Fixture(path string) string {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
