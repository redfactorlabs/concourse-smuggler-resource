package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	var dataDir string
	var requestType smuggler.RequestType

	// Determine which command is being called by the name
	commandName := filepath.Base(os.Args[0])
	switch {
	case strings.Contains(commandName, "check"):
		dataDir = ""
		requestType = smuggler.CheckType
	case strings.Contains(commandName, "in"):
		if len(os.Args) < 2 {
			utils.Sayf("usage: %s <dest directory>\n", os.Args[0])
			os.Exit(1)
		}
		dataDir = os.Args[1]
		requestType = smuggler.InType
	case strings.Contains(commandName, "out"):
		if len(os.Args) < 2 {
			utils.Sayf("usage: %s <sources directory>\n", os.Args[0])
			os.Exit(1)
		}
		dataDir = os.Args[1]
		requestType = smuggler.OutType
	default:
		utils.Fatal("identifying resource type",
			fmt.Errorf("command name '%s' does not contain check/in/out", commandName),
			1)
	}

	// Open Log file
	smugglerLogFileName := utils.GetEnvOrDefault("SMUGGLER_LOG", "/tmp/smuggler.log")
	tempFileLogger, err := utils.NewTempFileLogger(smugglerLogFileName)
	if err != nil {
		utils.Fatal("opening log '/tmp/smuggler.log'", err, 1)
	}

	// Read request
	request := smuggler.ResourceRequest{Type: requestType}
	inputRequest(&request)

	// Execute command
	command := smuggler.NewSmugglerCommand(tempFileLogger.Logger)

	response, err := command.RunAction(dataDir, request)
	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}
	os.Stderr.Write([]byte(command.LastCommandCombinedOuput()))

	// Send back response
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
