package main

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	var request smuggler.CheckRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunCheck(request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.CheckRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err, 1)
	}
}

func outputResponse(response smuggler.CheckResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}
