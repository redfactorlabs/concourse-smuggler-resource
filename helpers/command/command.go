package command

import (
	"log"
	"os/exec"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

type SmugglerCommand struct {
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
}

func (command *SmugglerCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
}

func (command *SmugglerCommand) Run(commandDefinition smuggler.CommandDefinition) error {
	path := commandDefinition.Path
	args := commandDefinition.Args
	log.Printf("[INFO] Running '%s %s'", path, strings.Join(args, " "))
	command.lastCommand = exec.Command(path, args...)
	output, err := command.lastCommand.CombinedOutput()
	if err != nil {
		return err
	}
	command.lastCommandCombinedOuput = string(output)
	log.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())
	return nil
}
