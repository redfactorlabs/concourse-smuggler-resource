package in

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

func main() {
	if len(os.Args) < 2 {
		smuggler.Sayf("usage: %s <sources directory>\n", os.Args[0])
		os.Exit(1)
	}

	sourceDir := os.Args[1]

	var request smuggler.OutRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunOut(sourceDir, request)
	if err != nil {
		smuggler.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.OutRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		smuggler.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response smuggler.OutResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		smuggler.Fatal("writing response to stdout", err)
	}
}
