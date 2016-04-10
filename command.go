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

func (command *SmugglerCommand) Run(commandDefinition CommandDefinition, params map[string]string, outputDir string) error {
	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, len(params)+1)
	params_env = append(params_env, fmt.Sprintf("SMUGGLER_OUTPUT_DIR=%s", outputDir))
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
	if ok, message := request.Source.IsValid(); !ok {
		return CheckResponse{}, errors.New(message)
	}

	var response = CheckResponse{}

	smugglerConfig := request.Source.SmugglerConfig
	if smugglerConfig.CheckCommand.IsDefined() {
		outputDir, err := ioutil.TempDir("", "smuggler-run")
		if err != nil {
			return response, err
		}
		defer os.RemoveAll(outputDir)

		err = command.Run(smugglerConfig.CheckCommand, request.Source.ExtraParams, outputDir)
		if err != nil {
			return response, err
		}

		versionsFilename := filepath.Join(outputDir, "versions")
		if _, err := os.Stat(versionsFilename); err == nil {
			content, err := ioutil.ReadFile(versionsFilename)
			if err != nil {
				return response, err
			}
			lines := strings.Split(string(content), "\n")
			versions := make([]Version, 0, len(lines))
			for i, _ := range lines {
				trimmedLine := strings.Trim(lines[i], "\t ")
				if trimmedLine != "" {
					versions = append(versions, Version{VersionID: trimmedLine})
				}
			}
			response = versions
		}
	}
	return response, nil
}
