package commands

import (
	"encoding/json"
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func SmugglerMain(dataDir string, requestType smuggler.RequestType) {
	smugglerLogFileName := utils.GetEnvOrDefault("SMUGGLER_LOG", "/tmp/smuggler.log")
	tempFileLogger, err := utils.NewTempFileLogger(smugglerLogFileName)
	if err != nil {
		utils.Fatal("opening log '/tmp/smuggler.log'", err, 1)
	}

	request := smuggler.ResourceRequest{Type: requestType}
	inputRequest(&request)

	command := smuggler.NewSmugglerCommand(tempFileLogger.Logger)

	response, err := command.RunAction(dataDir, request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}
	os.Stderr.Write([]byte(command.LastCommandCombinedOuput()))

	if requestType == smuggler.CheckType {
		outputResponseCheck(response.Versions)
	} else {
		outputResponseInOut(response)
	}
}

func inputRequest(request *smuggler.ResourceRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err, 1)
	}
}

func outputResponseCheck(response []smuggler.Version) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}

func outputResponseInOut(response smuggler.ResourceResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err, 1)
	}
}
