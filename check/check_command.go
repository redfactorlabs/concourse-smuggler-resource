package check

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

type CheckCommand struct {
}

func NewCheckCommand() *CheckCommand {
	return &CheckCommand{}
}

func (command *CheckCommand) Run(request smuggler.CheckRequest) (smuggler.CheckResponse, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return smuggler.CheckResponse{}, errors.New(message)
	}

	if request.Source.CheckCommand.IsDefined() {
		fmt.Printf("Running '%s %s'",
			request.Source.CheckCommand.Path,
			strings.Join(request.Source.CheckCommand.Args, " "))
	}
	return smuggler.CheckResponse{}, nil
}
