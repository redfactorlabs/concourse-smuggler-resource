package main

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	smugglerLogFileName := utils.GetEnvOrDefault("SMUGGLER_LOG", "/tmp/smuggler.log")
	tempFileLogger, err := utils.NewTempFileLogger(smugglerLogFileName)
	if err != nil {
		utils.Fatal("opening log '/tmp/smuggler.log'", err, 1)
	}

	request := smuggler.ResourceRequest{Type: smuggler.CheckType}
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand(tempFileLogger.Logger)

	response, err := command.RunAction("", request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}
	os.Stderr.Write([]byte(command.LastCommandCombinedOuput()))

	outputResponse(response.Versions)
}

func inputRequest(request *smuggler.ResourceRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err, 1)
	}
}

func outputResponse(response []smuggler.Version) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}
