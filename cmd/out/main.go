package main

import (
	"encoding/json"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <sourceDirectory>")
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode([]interface{}{})
}
