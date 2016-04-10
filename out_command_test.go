package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("Out Command", func() {
	It("it fails if it cannot find the version ID", func() {
		manifest := `
resources:
- name: out_command
  type: smuggler
  source:
    smuggler_config:
      out:
        path: bash
        args:
        - -e
        - -c
        - |
          true
`
		source, err := ResourceSourceFromYamlManifest(manifest, "out_command")
		Ω(err).ShouldNot(HaveOccurred())
		request := OutRequest{
			Source: *source,
		}
		command := NewSmugglerCommand()
		_, err = command.RunOut(request)
		Ω(err).Should(HaveOccurred())
	})
	It("it exports the version ID", func() {
		manifest := `
resources:
- name: out_command
  type: smuggler
  source:
    smuggler_config:
      out:
        path: bash
        args:
        - -e
        - -c
        - |
          echo "1.2.3" > $SMUGGLER_OUTPUT_DIR/version
          echo -e "\t 1.2.3   " >> $SMUGGLER_OUTPUT_DIR/version
          echo -e "\t 1.2.4   " >> $SMUGGLER_OUTPUT_DIR/version
`
		source, err := ResourceSourceFromYamlManifest(manifest, "out_command")
		Ω(err).ShouldNot(HaveOccurred())
		request := OutRequest{
			Source: *source,
		}
		command := NewSmugglerCommand()
		outResponse, err := command.RunOut(request)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(outResponse.Version).Should(BeEquivalentTo(Version{VersionID: "1.2.3"}))
	})
	It("passes the resource params as environment variables", func() {
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
      out:
        path: sh
        args:
        - -e
        - -c
        - |
          echo "param1=${SMUGGLER_param1}"
          echo "param2=${SMUGGLER_param2}"
          echo "param3=${SMUGGLER_param3}"
          echo "param4=${SMUGGLER_param4}"
          echo "param5=${SMUGGLER_param5}"
          echo "1.2.3" > ${SMUGGLER_OUTPUT_DIR}/version
`
		source, err := ResourceSourceFromYamlManifest(manifest, "pass_params")
		Ω(err).ShouldNot(HaveOccurred())
		request := OutRequest{
			Source: *source,
			Params: map[string]string{
				"param4": "val4",
				"param5": "something with spaces",
			},
		}
		checkCommand := NewSmugglerCommand()
		checkCommand.RunOut(request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
	})
})
