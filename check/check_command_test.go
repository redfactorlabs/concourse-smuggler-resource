package check_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/check"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
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
		Ω(checkCommand.SmugglerCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})
	It("executes a basic echo command from yaml manifest and returns the output", func() {
		source, err := ResourceSourceFromYamlManifest(manifestResourceDefinitions, "simple_echo")
		Ω(err).ShouldNot(HaveOccurred())
		requestBasicEcho := CheckRequest{
			Source: *source,
		}
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
		Ω(checkCommand.SmugglerCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})
	It("it can run multiple commands passed in multiple lines", func() {
		source, err := ResourceSourceFromYamlManifest(manifestResourceDefinitions, "multiline_command")
		Ω(err).ShouldNot(HaveOccurred())
		requestBasicEcho := CheckRequest{
			Source: *source,
		}
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
		Ω(checkCommand.SmugglerCommand.LastCommandCombinedOuput()).Should(ContainSubstring("line1"))
		Ω(checkCommand.SmugglerCommand.LastCommandCombinedOuput()).Should(ContainSubstring("line2"))
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

var manifestResourceDefinitions = `
resources:
- name: simple_echo
  type: smuggler
  source:
    smuggler_config:
      check:
        path: sh
        args:
        - -e
        - -c
        - |
          echo basic echo test
- name: multiline_command
  type: smuggler
  source:
    smuggler_config:
      check:
        path: sh
        args:
        - -e
        - -c
        - |
          echo line1
          echo line2
`
