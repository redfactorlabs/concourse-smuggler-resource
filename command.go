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
)

type SmugglerCommand struct {
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
}

func NewSmugglerCommand() *SmugglerCommand {
	return &SmugglerCommand{}
}

func (command *SmugglerCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
}

func (command *SmugglerCommand) Run(commandDefinition CommandDefinition, params map[string]string) error {
	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, len(params)+1)
	for k, v := range params {
		params_env = append(params_env, fmt.Sprintf("SMUGGLER_%s=%s", k, v))
	}
	params_env = append(params_env, os.Environ()...)

	log.Printf("[INFO] Running '%s %s' with env:\n\t",
		path, strings.Join(args, " "), strings.Join(params_env, "\n\t"))

	command.lastCommand = exec.Command(path, args...)
	command.lastCommand.Env = params_env
	output, err := command.lastCommand.CombinedOutput()
	if err != nil {
		return err
	}
	command.lastCommandCombinedOuput = string(output)
	log.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())
	return nil
}

func (command *SmugglerCommand) RunCheck(request CheckRequest) (CheckResponse, error) {
	var response = CheckResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	smugglerConfig := request.Source.SmugglerConfig
	if !smugglerConfig.CheckCommand.IsDefined() {
		return response, nil
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params := copyMaps(request.Source.ExtraParams)
	params["OUTPUT_DIR"] = outputDir

	err = command.Run(smugglerConfig.CheckCommand, params)
	if err != nil {
		return response, err
	}

	response, err = readVersions(filepath.Join(outputDir, "versions"))
	if err != nil {
		return response, err
	}

	return response, nil
}

func (command *SmugglerCommand) RunIn(request InRequest) (InResponse, error) {
	var response = InResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	smugglerConfig := request.Source.SmugglerConfig
	if !smugglerConfig.InCommand.IsDefined() {
		return response, nil
	}

	outputDir, err := ioutil.TempDir("", "smuggler-run")
	if err != nil {
		return response, err
	}
	defer os.RemoveAll(outputDir)

	params := copyMaps(request.Source.ExtraParams, request.Params)
	params["VERSION_ID"] = request.Version.VersionID
	params["OUTPUT_DIR"] = outputDir

	err = command.Run(smugglerConfig.InCommand, params)
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
