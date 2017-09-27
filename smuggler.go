package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)

func main() {
	defer utils.PrintRecover()

	dataDir, requestType := processArguments()

	// Open Logger
	tempFileLogger := openSmugglerLog()
	logger = tempFileLogger.Logger

	// Read request
	request, jsonRequest := inputRequest(requestType)

	// Dump logs to stderr if required
	if request.Source.SmugglerDebug {
		tempFileLogger.DupToStderr()
		logger = tempFileLogger.Logger
	}

	// Execute command
	command := smuggler.NewSmugglerCommand(tempFileLogger.Logger)

	logger.Printf(
		"[INFO] Smuggler command called as:\n%s <<\"EOF\"\n%sEOF",
		strings.Join(os.Args, " "),
		utils.JsonPrettyPrint(jsonRequest),
	)

	response, err := command.RunAction(dataDir, request)

	// Print output to stderr
	if len(command.LastCommandErr) > 0 {
		fmt.Fprintf(os.Stderr, "Stderr:")
		os.Stderr.Write(command.LastCommandErr)
	}
	if len(command.LastCommandOutput) > 0 {
		fmt.Fprintf(os.Stderr, "Stdout:")
		os.Stderr.Write(command.LastCommandOutput)
	}

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
func inputRequest(requestType smuggler.RequestType) (*smuggler.ResourceRequest, []byte) {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		utils.Panic("reading request from stdin", err)
	}

	smugglerConfig := findAndReadSmugglerConfig()

	r := ParseInputAndConfig(requestType, input, smugglerConfig)

	return r, input
}

func ParseInputAndConfig(requestType smuggler.RequestType, input []byte, config []byte) *smuggler.ResourceRequest {
	if len(config) > 0 {
		var requestCatchAll struct {
			Source  map[string]interface{} `json:"source,omitempty"`
			Version map[string]interface{} `json:"version,omitempty"`
			Params  map[string]interface{} `json:"params,omitempty"`
		}
		var configCatchAll map[string]interface{}

		err := json.Unmarshal(input, &requestCatchAll)
		if err != nil {
			utils.Panic("Error parsing request: %s", err)
		}
		err = yaml.Unmarshal(config, &configCatchAll)
		if err != nil {
			utils.Panic("Error parsing 'smuggler.yml': %s", err)
		}

		commands, err := utils.MergeMaps(requestCatchAll.Source["commands"], configCatchAll["commands"])
		if err != nil {
			utils.Panic("Format error in 'commands', is not a map: %s", err)
		}
		smuggler_params, err := utils.MergeMaps(requestCatchAll.Source["smuggler_params"], configCatchAll["smuggler_params"])
		if err != nil {
			utils.Panic("Format error in 'smuggler_params', is not a map: %s", err)
		}

		if requestCatchAll.Source == nil {
			requestCatchAll.Source = make(map[string]interface{})
		}
		for k, v := range configCatchAll {
			requestCatchAll.Source[k] = v
		}
		requestCatchAll.Source["commands"] = commands
		requestCatchAll.Source["smuggler_params"] = smuggler_params

		input, err = json.Marshal(&requestCatchAll)
		if err != nil {
			utils.Panic("Error merging 'smuggler.yml': %s", err)
		}
	}
	request, err := smuggler.NewResourceRequest(requestType, string(input))
	if err != nil {
		utils.Panic("Error parsing request from stdin: %s", err)
	}
	return request
}

func findAndReadSmugglerConfig() []byte {
	smugglerYmlPaths := []string{
		filepath.Join(filepath.Dir(os.Args[0]), "smuggler.yml"),
		utils.GetEnvOrDefault("SMUGGLER_CONFIG", "/opt/resource/smuggler.yml"),
	}

	smugglerConfigFile := ""
OuterLoop:
	for _, f := range smugglerYmlPaths {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			smugglerConfigFile = f
			break OuterLoop
		}
	}
	if smugglerConfigFile == "" {
		logger.Printf("[INFO] No config file in any of: %s", strings.Join(smugglerYmlPaths, ", "))
		return []byte{}
	}
	logger.Printf("[INFO] Found config file %s", smugglerConfigFile)

	content, err := ioutil.ReadFile(smugglerConfigFile)
	if err != nil {
		utils.Panic("Error reading '%s': %s", smugglerConfigFile, err)
	}

	return content
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
