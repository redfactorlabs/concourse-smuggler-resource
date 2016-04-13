package smuggler

import (
	"errors"
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
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
	logger                   *log.Logger
}

func NewSmugglerCommand(logger *log.Logger) *SmugglerCommand {
	return &SmugglerCommand{logger: logger}
}

func (command *SmugglerCommand) LastCommand() *exec.Cmd {
	return command.lastCommand
}

func (command *SmugglerCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
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

func (command *SmugglerCommand) Run(commandDefinition CommandDefinition, params map[string]string) error {
	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, 0, len(params))
	for k, v := range params {
		params_env = append(params_env, fmt.Sprintf("SMUGGLER_%s=%s", k, v))
	}
	params_env = append(params_env, os.Environ()...)

	command.logger.Printf("[INFO] Running command:\n\tPath: '%s'\n\tArgs: '%s'\n\tEnv:\n\t'%s",
		path, strings.Join(args, "' '"), strings.Join(params_env, "',\n\t'"))

	command.lastCommand = exec.Command(path, args...)
	command.lastCommand.Env = params_env

	output, err := command.lastCommand.CombinedOutput()
	command.lastCommandCombinedOuput = string(output)
	command.logger.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())

	return err
}

func (command *SmugglerCommand) RunAction(dataDir string, request ResourceRequest) (ResourceResponse, error) {
	command.logger.Printf("[INFO] Running %s action", string(request.Type))

	var response = ResourceResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
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

	params := copyMaps(request.Source.ExtraParams, request.Params)
	params["ACTION"] = string(request.Type)
	params["COMMAND"] = string(request.Type)
	params["OUTPUT_DIR"] = outputDir
	switch request.Type {
	case "in":
		params["DESTINATION_DIR"] = dataDir
		params["VERSION_ID"] = request.Version.VersionID
	case "out":
		params["SOURCES_DIR"] = dataDir
	}

	err = command.Run(*commandDefinition, params)
	if err != nil {
		return response, err
	}

	versions, err := readVersions(filepath.Join(outputDir, "versions"))
	if err != nil {
		return response, err
	}
	command.logger.Printf("[INFO] command reports versions '%q'", versions)

	metadata, err := readMetadata(filepath.Join(outputDir, "metadata"))
	if err != nil {
		return response, err
	}
	command.logger.Printf("[INFO] command reports metadata '%q'", metadata)

	switch request.Type {
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

	return response, nil
}

func copyMaps(maps ...map[string]string) map[string]string {
	total_len := 0
	for _, m := range maps {
		total_len += len(m)
	}
	result := make(map[string]string, total_len)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func readVersions(versionsFile string) ([]Version, error) {
	result := []Version{}
	if versionLines, err := readAndTrimAllLines(versionsFile); err != nil {
		return result, err
	} else {
		for _, l := range versionLines {
			result = append(result, Version{VersionID: l})
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
