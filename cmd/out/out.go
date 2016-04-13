package main

import (
	"os"

	"github.com/redfactorlabs/concourse-smuggler-resource/cmd"
	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
	"github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

func main() {
	if len(os.Args) < 2 {
		utils.Sayf("usage: %s <sources directory>\n", os.Args[0])
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	commands.SmugglerMain(sourceDir, smuggler.OutType)
}
