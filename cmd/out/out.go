package in

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	if len(os.Args) < 2 {
		utils.Sayf("usage: %s <sources directory>\n", os.Args[0])
		os.Exit(1)
	}

	sourceDir := os.Args[1]

	var request smuggler.OutRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunOut(sourceDir, request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.OutRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err, 1)
	}
}

func outputResponse(response smuggler.OutResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}
