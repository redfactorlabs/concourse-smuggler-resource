package main

import (
	"encoding/json"
	"os"
)

func main() {
	// no-op check
	json.NewEncoder(os.Stdout).Encode([]interface{}{})
}
