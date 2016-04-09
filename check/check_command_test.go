package check_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/check"
)

var requestBasicEcho = CheckRequest{
	Source: Source{
		CheckCommand: CommandDefinition{
			Path: "bash",
			Args: []string{"-e", "-c", "echo basic echo test"},
		},
	},
}

var requestBasicEchoJson = `
{
  "source": {
    "check": {
		"path": "sh",
		"args": [ "-e", "-c", "echo basic echo test" ]
	}
  },
  "version": {}
}
`

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
})
