package check

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

func main() {
	var request smuggler.CheckRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunCheck(request)
	if err != nil {
		smuggler.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.CheckRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		smuggler.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response smuggler.CheckResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		smuggler.Fatal("writing response to stdout", err)
	}
}
