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
}

func NewSmugglerCommand() *SmugglerCommand {
	return &SmugglerCommand{}
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

	log.Printf("[INFO] Running '%s %s' with env:\n\t",
		path, strings.Join(args, " "), strings.Join(params_env, "\n\t"))

	command.lastCommand = exec.Command(path, args...)
	command.lastCommand.Env = params_env

	output, err := command.lastCommand.CombinedOutput()
	command.lastCommandCombinedOuput = string(output)
	log.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())

	return err
}

func (command *SmugglerCommand) RunCheck(request CheckRequest) (CheckResponse, error) {
	var response = CheckResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	commandDefinition := request.Source.FindCommand("check")
	if commandDefinition == nil {
		return response, nil
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params := copyMaps(request.Source.ExtraParams)
	params["OUTPUT_DIR"] = outputDir

	err = command.Run(*commandDefinition, params)
	if err != nil {
		return response, err
	}

	response, err = readVersions(filepath.Join(outputDir, "versions"))
	if err != nil {
		return response, err
	}

	return response, nil
}

func (command *SmugglerCommand) RunIn(destinationDir string, request InRequest) (InResponse, error) {
	var response = InResponse{
		Version: request.Version,
	}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	commandDefinition := request.Source.FindCommand("in")
	if commandDefinition == nil {
		return response, nil
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params := copyMaps(request.Source.ExtraParams, request.Params)
	params["DESTINATION_DIR"] = destinationDir
	params["VERSION_ID"] = request.Version.VersionID
	params["OUTPUT_DIR"] = outputDir

	err = command.Run(*commandDefinition, params)
	if err != nil {
		return response, err
	}

	// We always use the same version that we get in the request
	response.Version = request.Version
	response.Metadata, err = readMetadata(filepath.Join(outputDir, "metadata"))
	if err != nil {
		return response, err
	}

	return response, nil
}

func (command *SmugglerCommand) RunOut(sourcesDir string, request OutRequest) (OutResponse, error) {
	var response = OutResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	commandDefinition := request.Source.FindCommand("out")
	if commandDefinition == nil {
		return response, errors.New("No out command defined. You must define one to generate versions.")
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params := copyMaps(request.Source.ExtraParams, request.Params)
	params["SOURCES_DIR"] = sourcesDir
	params["OUTPUT_DIR"] = outputDir

	err = command.Run(*commandDefinition, params)
	if err != nil {
		return response, err
	}

	versions, err := readVersions(filepath.Join(outputDir, "version"))
	if err != nil {
		return response, err
	}
	if len(versions) == 0 {
		return response, fmt.Errorf("No version found in '%s'", filepath.Join(outputDir, "version"))
	}
	response.Version = versions[0]
	response.Metadata, err = readMetadata(filepath.Join(outputDir, "metadata"))
	if err != nil {
		return response, err
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
