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

	if request.Source.CheckCommand.IsDefined() {
		path := request.Source.CheckCommand.Path
		args := request.Source.CheckCommand.Args
		log.Printf("[INFO] Running '%s %s'", path, strings.Join(args, " "))
		command.lastCommand = exec.Command(path, args...)
		output, err := command.lastCommand.CombinedOutput()
		if err != nil {
			return nil, err
		}
		command.lastCommandCombinedOuput = string(output)
		log.Printf("[INFO] Output '%s'", command.lastCommandCombinedOuput)
	}
	return smuggler.CheckResponse{}, nil
}
