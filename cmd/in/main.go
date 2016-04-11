package in

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

func main() {
	if len(os.Args) < 3 {
		smuggler.Sayf("usage: %s <dest directory>\n", os.Args[0])
		os.Exit(1)
	}

	destinationDir := os.Args[1]

	var request smuggler.InRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunIn(destinationDir, request)
	if err != nil {
		smuggler.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.InRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		smuggler.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response smuggler.InResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		smuggler.Fatal("writing response to stdout", err)
	}
}
