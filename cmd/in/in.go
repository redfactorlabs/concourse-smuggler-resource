package in

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	if len(os.Args) < 3 {
		utils.Sayf("usage: %s <dest directory>\n", os.Args[0])
		os.Exit(1)
	}

	destinationDir := os.Args[1]

	var request smuggler.InRequest
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand()

	response, err := command.RunIn(destinationDir, request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}

	outputResponse(response)
}

func inputRequest(request *smuggler.InRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err, 1)
	}
}

func outputResponse(response smuggler.InResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}
