package smuggler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type SmugglerCommand struct {
	lastCommand       *exec.Cmd
	logger            *log.Logger
	LastCommandOutput []byte
	LastCommandErr    []byte
}

func NewSmugglerCommand(logger *log.Logger) *SmugglerCommand {
	return &SmugglerCommand{logger: logger}
}

func (command *SmugglerCommand) LastCommand() *exec.Cmd {
	return command.lastCommand
}

func (command *SmugglerCommand) LastCommandSuccess() bool {
	if command.lastCommand == nil || command.lastCommand.ProcessState == nil {
		return true
	}
	return command.lastCommand.ProcessState.Success()
}

func (command *SmugglerCommand) LastCommandExitStatus() int {
	waitStatus := command.lastCommand.ProcessState.Sys().(syscall.WaitStatus)
	return waitStatus.ExitStatus()
}

func (command *SmugglerCommand) Run(commandDefinition CommandDefinition, params map[string]interface{}, jsonRequest []byte) error {

	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, 0, len(params))
	for k, v := range params {
		string_val := InterfaceToJsonString(v)
		env_key_val := fmt.Sprintf("SMUGGLER_%s=%s", k, string_val)
		params_env = append(params_env, env_key_val)
	}
	params_env = append(params_env, os.Environ()...)

	command.logger.Printf("[INFO] Running command:\n\tPath: '%s'\n\tArgs: '%s'\n\tEnv:\n\t'%s",
		path, strings.Join(args, "' '"), strings.Join(params_env, "',\n\t'"))

	command.lastCommand = exec.Command(path, args...)
	command.lastCommand.Env = params_env

	command.lastCommand.Stdin = bytes.NewBuffer(jsonRequest)
	stdout := new(bytes.Buffer)
	command.lastCommand.Stdout = stdout
	stderr := new(bytes.Buffer)
	command.lastCommand.Stderr = stderr

	err := command.lastCommand.Run()
	command.LastCommandOutput, _ = ioutil.ReadAll(stdout)
	command.LastCommandErr, _ = ioutil.ReadAll(stderr)
	command.logger.Printf("[INFO] Output '%s'", command.LastCommandOutput)
	command.logger.Printf("[INFO] Stderr '%s'", command.LastCommandErr)
	command.logger.Printf("[INFO] Return error '%v'", err)

	return err
}

func (command *SmugglerCommand) RunAction(dataDir string, request ResourceRequest) (ResourceResponse, error) {
	command.logger.Printf("[INFO] Running %s action", string(request.Type))

	var response = ResourceResponse{
		Type: request.Type,
	}

	commandDefinition := request.Source.FindCommand(string(request.Type))
	if commandDefinition == nil {
		command.logger.Printf("[INFO] No command definition, skipping")
		return response, nil
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params, err := prepareParams(dataDir, outputDir, request)
	if err != nil {
		return response, err
	}

	jsonRequest, err := prepareJsonRequest(request)
	if err != nil {
		return response, err
	}

	err = command.Run(*commandDefinition, params, jsonRequest)
	if err != nil {
		return response, err
	}

	// Try to get the response from a valid json from Stdout.
	// If not, as files from the output directory
	err = populateResponseFromStdoutAsJson(command.LastCommandOutput, &request, &response)
	if err != nil {
		err = populateResponseFromOutputDir(outputDir, &request, &response)
		if err != nil {
			return response, err
		}
	}

	command.logger.Printf("[INFO] command reports versions '%q'", response.Versions)
	command.logger.Printf("[INFO] command reports metadata '%q'", response.Metadata)

	return response, nil
}

func copyMaps(maps ...map[string]interface{}) map[string]interface{} {
	total_len := 0
	for _, m := range maps {
		total_len += len(m)
	}
	result := make(map[string]interface{}, total_len)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func prepareParams(dataDir string, outputDir string, request ResourceRequest) (map[string]interface{}, error) {
	// Prepare the params to send to the commands
	params := copyMaps(
		request.Source.SmugglerParams,
		request.Source.ExtraParams,
		request.Params.SmugglerParams,
		request.Params.ExtraParams,
	)
	params["ACTION"] = string(request.Type)
	params["COMMAND"] = string(request.Type)
	params["OUTPUT_DIR"] = outputDir
	switch request.Type {
	case "check":
		params["VERSION_ID"] = InterfaceToJsonString(request.Version)
	case "in":
		params["DESTINATION_DIR"] = dataDir
		params["VERSION_ID"] = InterfaceToJsonString(request.Version)
	case "out":
		params["SOURCES_DIR"] = dataDir
	}

	return params, nil
}

func prepareJsonRequest(request ResourceRequest) ([]byte, error) {
	jsonRequest, err := json.Marshal(request.OrigRequest)
	return jsonRequest, err
}

//
// Tries to populate the response from the stdout
//
func populateResponseFromStdoutAsJson(stdout []byte, request *ResourceRequest, response *ResourceResponse) error {
	var r ResourceResponse
	err := json.Unmarshal(stdout, &r)
	if err != nil {
		return err
	}
	// Copy all the new values but the type to the response
	r.Type = response.Type
	*response = r
	return nil
}

//
// Tries to get the Request from the filesystem
//
func populateResponseFromOutputDir(outputDir string, request *ResourceRequest, response *ResourceResponse) error {
	versions, err := readVersions(filepath.Join(outputDir, "versions"))
	if err != nil {
		return err
	}

	metadata, err := readMetadata(filepath.Join(outputDir, "metadata"))
	if err != nil {
		return err
	}

	switch response.Type {
	case "check":
		response.Versions = versions
	case "in", "out":
		if len(versions) > 0 {
			response.Version = versions[0]
		} else {
			response.Version = request.Version
		}
		response.Metadata = metadata
	}

	return nil
}

func readVersions(versionsFile string) ([]interface{}, error) {
	result := []interface{}{}
	if versionLines, err := readAndTrimAllLines(versionsFile); err != nil {
		return result, err
	} else {
		for _, l := range versionLines {
			result = append(result, JsonStringToInterface(l))
		}
	}
	return result, nil
}

func readMetadata(metadataFile string) ([]MetadataPair, error) {
	result := []MetadataPair{}
	if metadataLines, err := readAndTrimAllLines(metadataFile); err != nil {
		return result, err
	} else {
		for _, l := range metadataLines {
			s := strings.SplitN(l, "=", 2)
			k, v := "", ""
			k = strings.Trim(s[0], " \t")
			if len(s) > 1 {
				v = strings.Trim(s[1], " \t")
			}
			result = append(result, MetadataPair{Name: k, Value: v})
		}
	}
	return result, nil
}

func readAndTrimAllLines(filename string) ([]string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []string{}, nil
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}
	fileLines := strings.Split(string(content), "\n")
	validLines := make([]string, 0, len(fileLines))
	for _, l := range fileLines {
		trimmedLine := strings.Trim(l, "\t ")
		if trimmedLine != "" {
			validLines = append(validLines, trimmedLine)
		}
	}
	return validLines, nil
}
