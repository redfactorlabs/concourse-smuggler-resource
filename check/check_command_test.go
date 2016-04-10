package check_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/check"
)

var _ = Describe("Check Command", func() {
	It("executes a basic echo command", func() {
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
	})
	It("executes a basic echo command from json", func() {
		requestBasicEcho, err := NewCheckRequestFromJson(requestBasicEchoJson)
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
	})
	It("executes a basic echo command from json and returns the output", func() {
		requestBasicEcho, err := NewCheckRequestFromJson(requestBasicEchoJson)
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})
})

var requestBasicEcho = CheckRequest{
	Source: Source{
		SmugglerConfig: SmugglerConfig{
			CheckCommand: CommandDefinition{
				Path: "bash",
				Args: []string{"-e", "-c", "echo basic echo test"},
			},
		},
	},
}

var requestBasicEchoJson = `
{
  "source": {
    "smuggler_config": {
      "check": {
	"path": "sh",
	"args": [ "-e", "-c", "echo basic echo test" ]
      }
    }
  },
  "version": {}
}
`
