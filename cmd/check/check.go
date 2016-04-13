package main

import (
	"github.com/redfactorlabs/concourse-smuggler-resource/cmd"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	commands.SmugglerMain("", smuggler.CheckType)
}
