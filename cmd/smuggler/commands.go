package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	defer utils.PrintRecover()

	dataDir, requestType := processArguments()

	tempFileLogger := openSmugglerLog()

	// Read request
	request := inputRequest(requestType)

	// Execute command
	command := smuggler.NewSmugglerCommand(tempFileLogger.Logger)

	response, err := command.RunAction(dataDir, request)

	// Print output to stderr
	os.Stderr.Write(command.LastCommandOutput)
	os.Stderr.Write(command.LastCommandErr)

	if err != nil {
		utils.Fatal("running command", err, command.LastCommandExitStatus())
	}

	outputResponse(response)
}

// Determine which command is being called by the name
func processArguments() (string, smuggler.RequestType) {
	var dataDir string
	var requestType smuggler.RequestType

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
		utils.Panic("identifying resource type: command name '%s' does not contain check/in/out", commandName)
	}

	return dataDir, requestType
}

func openSmugglerLog() *utils.TempFileLogger {
	// Open Log file
	smugglerLogFileName := utils.GetEnvOrDefault("SMUGGLER_LOG", "/tmp/smuggler.log")
	tempFileLogger, err := utils.NewTempFileLogger(smugglerLogFileName)
	if err != nil {
		utils.Panic("opening log '%s': %s", smugglerLogFileName, err)
	}
	return tempFileLogger
}

// Read input request, merged with the configuration file
func inputRequest(requestType smuggler.RequestType) *smuggler.ResourceRequest {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		utils.Panic("reading request from stdin", err)
	}

	// If there is a config in filesystem, merge it with the request
	smugglerConfig := readSmugglerConfig()
	if smugglerConfig != nil {
		// Create jsonRequest merged with the config
		var rawRequest interface{}
		err := json.Unmarshal(input, &rawRequest)
		if err != nil {
			utils.Panic("parsing request from stdin", err)
		}
		err = mergo.MergeWithOverwrite(smugglerConfig, rawRequest)
		if err != nil {
			utils.Panic("merging config 'smuggler.yml' with request", err)
		}
		input, err = json.Marshal(smugglerConfig)
		if err != nil {
			utils.Panic("merging config 'smuggler.yml' with request (to json)", err)
		}
	}

	request, err := smuggler.NewResourceRequest(requestType, string(input))
	if err != nil {
		utils.Panic("parsing request from stdin", err)
	}

	return request
}

// Load the local config file
func readSmugglerConfig() interface{} {
	var source smuggler.SmugglerSource
	var config interface{}

	smugglerConfigFile := filepath.Join(filepath.Dir(os.Args[0]), "smuggler.yml")
	if _, err := os.Stat(smugglerConfigFile); os.IsNotExist(err) {
		return nil
	}
	content, err := ioutil.ReadFile(smugglerConfigFile)
	if err != nil {
		utils.Panic("Error reading '%s': %s", smugglerConfigFile, err)
	}

	// Check the syntax of the config
	yaml.Unmarshal(content, &source)
	if err != nil {
		utils.Panic("Error parsing '%s': %s", smugglerConfigFile, err)
	}
	// But unmarshall in raw
	yaml.Unmarshal(content, &config)

	return &map[string]interface{}{"source": config}
}

// Send back response
func outputResponse(response *smuggler.ResourceResponse) {
	if response.Type == smuggler.CheckType {
		outputResponseCheck(response.Versions)
	} else {
		outputResponseInOut(response)
	}
}

func outputResponseCheck(response []smuggler.Version) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Panic("writing response to stdout: %s", err)
	}
}

func outputResponseInOut(response *smuggler.ResourceResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Panic("writing response to stdout: %s", err)
	}
}
