package check

import (
	"errors"

	"github.com/redfactorlabs/concourse-smuggler-resource"
	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/command"
)

type CheckCommand struct {
	SmugglerCommand command.SmugglerCommand
}

func NewCheckCommand() *CheckCommand {
	return &CheckCommand{}
}

func (command *CheckCommand) Run(request smuggler.CheckRequest) (smuggler.CheckResponse, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return smuggler.CheckResponse{}, errors.New(message)
	}
	smugglerConfig := request.Source.SmugglerConfig
	if smugglerConfig.CheckCommand.IsDefined() {
		err := command.SmugglerCommand.Run(smugglerConfig.CheckCommand, request.Source.ExtraParams)
		if err != nil {
			return nil, err
		}
	}
	return smuggler.CheckResponse{}, nil
}
