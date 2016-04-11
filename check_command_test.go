package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("Check Command", func() {
	Context("when given a basic config from a structure", func() {
		request := CheckRequest{
			Source: Source{
				Commands: []CommandDefinition{
					CommandDefinition{
						Name: "check",
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
		}

		It("it executes the command and captures the output", func() {
			command := NewSmugglerCommand()
			command.RunCheck(request)
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
		})
	})

	Context("when given a basic config from a json", func() {
		var command *SmugglerCommand
		var response CheckResponse
		requestJson := `{
			"source": {
				"commands": [
					{
			"name": "check",
			"path": "sh",
			"args": [ "-e", "-c", "echo basic echo test" ]
					}
				]
			},
			"version": {}
		}`

		BeforeEach(func() {
			request, err := NewCheckRequestFromJson(requestJson)
			Ω(err).ShouldNot(HaveOccurred())
			command = NewSmugglerCommand()
			response, err = command.RunCheck(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("it executes the command and captures the output", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
		})
	})

	Context("when given a config with a complex script from yaml", func() {
		var request CheckRequest
		var command *SmugglerCommand
		var response CheckResponse
		manifest := `
resources:
- name: complex_command
  type: smuggler
  source:
    extra_params:
      param1: test
      param2: true
      param3: 123
    commands:
    - name: check
      path: bash
      args:
      - -e
      - -c
      - |
        echo Command Start
        echo "param1=${SMUGGLER_param1}"
        echo "param2=${SMUGGLER_param2}"
        echo "param3=${SMUGGLER_param3}"
        echo "1.2.3" > ${SMUGGLER_OUTPUT_DIR}/versions
        echo -e "\n   " >> ${SMUGGLER_OUTPUT_DIR}/versions
        echo -e "\t 1.2.4  \n" >> ${SMUGGLER_OUTPUT_DIR}/versions
        echo Command End
`

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}
			command = NewSmugglerCommand()
			response, err = command.RunCheck(request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("executes several lines of the script", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
		})
		It("it can sets the resource extra_params as environment variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
		})
		It("it returns versions as list of strings", func() {
			vs := []Version{Version{VersionID: "1.2.3"}, Version{VersionID: "1.2.4"}}
			Ω(response).Should(BeEquivalentTo(vs))
		})
	})
})
