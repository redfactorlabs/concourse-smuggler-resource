package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("Check Command", func() {
	It("executes a basic echo command", func() {
		request := CheckRequest{
			Source: Source{
				SmugglerConfig: SmugglerConfig{
					CheckCommand: CommandDefinition{
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
		}

		checkCommand := NewSmugglerCommand()
		checkCommand.RunCheck(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})

	It("executes a basic echo command from json", func() {
		requestJson := `
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
		request, err := NewCheckRequestFromJson(requestJson)
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewSmugglerCommand()
		checkCommand.RunCheck(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})

	It("executes a basic echo command from yaml manifest", func() {
		manifest := `
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
`
		source, err := ResourceSourceFromYamlManifest(manifest, "simple_echo")
		Ω(err).ShouldNot(HaveOccurred())
		request := CheckRequest{
			Source: *source,
		}
		checkCommand := NewSmugglerCommand()
		checkCommand.RunCheck(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})

	It("it can run multiple commands passed in multiple lines", func() {
		manifest := `
resources:
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
		source, err := ResourceSourceFromYamlManifest(manifest, "multiline_command")
		Ω(err).ShouldNot(HaveOccurred())
		request := CheckRequest{
			Source: *source,
		}
		checkCommand := NewSmugglerCommand()
		checkCommand.RunCheck(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("line1"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("line2"))
	})

	It("it can passes the resource params as environment variables", func() {
		manifest := `
resources:
- name: pass_params
  type: smuggler
  source:
    extra_params:
      param1: test
      param2: true
      param3: 123
    smuggler_config:
      check:
        path: sh
        args:
        - -e
        - -c
        - |
          echo "param1=${SMUGGLER_param1}"
          echo "param2=${SMUGGLER_param2}"
          echo "param3=${SMUGGLER_param3}"
`
		source, err := ResourceSourceFromYamlManifest(manifest, "pass_params")
		Ω(err).ShouldNot(HaveOccurred())
		request := CheckRequest{
			Source: *source,
		}
		checkCommand := NewSmugglerCommand()
		checkCommand.RunCheck(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
	})
	It("it returns versions as list of strings", func() {
		manifest := `
resources:
- name: output_versions
  type: smuggler
  source:
    smuggler_config:
      check:
        path: bash
        args:
        - -e
        - -c
        - |
          echo "1.2.3" > ${SMUGGLER_OUTPUT_DIR}/versions
          echo -e "\n   " >> ${SMUGGLER_OUTPUT_DIR}/versions
          echo -e "\t 1.2.4  \n" >> ${SMUGGLER_OUTPUT_DIR}/versions
`
		source, err := ResourceSourceFromYamlManifest(manifest, "output_versions")
		Ω(err).ShouldNot(HaveOccurred())
		request := CheckRequest{
			Source: *source,
		}
		checkCommand := NewSmugglerCommand()
		checkResponse, err := checkCommand.RunCheck(request)
		Ω(err).ShouldNot(HaveOccurred())

		vs := []Version{Version{VersionID: "1.2.3"}, Version{VersionID: "1.2.4"}}

		Ω(checkResponse).Should(BeEquivalentTo(vs))
	})
})
