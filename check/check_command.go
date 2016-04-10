package check

import (
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

type CheckCommand struct {
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
}

func NewCheckCommand() *CheckCommand {
	return &CheckCommand{}
}

func (command *CheckCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
}

func (command *CheckCommand) Run(request smuggler.CheckRequest) (smuggler.CheckResponse, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return smuggler.CheckResponse{}, errors.New(message)
	}
	smugglerConfig := request.Source.SmugglerConfig
	if smugglerConfig.CheckCommand.IsDefined() {
		path := smugglerConfig.CheckCommand.Path
		args := smugglerConfig.CheckCommand.Args
		log.Printf("[INFO] Running '%s %s'", path, strings.Join(args, " "))
		command.lastCommand = exec.Command(path, args...)
		output, err := command.lastCommand.CombinedOutput()
		if err != nil {
			return nil, err
		}
		command.lastCommandCombinedOuput = string(output)
		log.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())
	}
	return smuggler.CheckResponse{}, nil
}
