package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("In Command", func() {
	Context("when given a config with a complex script from yaml", func() {
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
    - name: in
      path: bash
      args:
      - -e
      - -c
      - |
        echo Command Start
        echo "version=$SMUGGLER_VERSION_ID"
        echo "destinationDir=$SMUGGLER_DESTINATION_DIR"
        echo "param1=${SMUGGLER_param1}"
        echo "param2=${SMUGGLER_param2}"
        echo "param3=${SMUGGLER_param3}"
        echo "param4=${SMUGGLER_param4}"
        echo "param5=${SMUGGLER_param5}"
        echo "value1= something quite long  " > ${SMUGGLER_OUTPUT_DIR}/metadata
        echo -e "\n   " >> ${SMUGGLER_OUTPUT_DIR}/metadata
        echo -e "\t value_2=2  \n" >> ${SMUGGLER_OUTPUT_DIR}/metadata
        echo Command End
`
		var request InRequest
		var command *SmugglerCommand
		var response InResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = InRequest{
				Source:  *source,
				Version: Version{VersionID: "1.2.3"},
				Params: map[string]string{
					"param4": "val4",
					"param5": "something with spaces",
				},
			}
			command = NewSmugglerCommand()
			response, err = command.RunIn("/tmp/destination/dir", request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("executes several lines of the script", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
		})

		It("it sets the resource extra_params and 'get' params as environment variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
		})

		It("it gets the destination dir", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("destinationDir=/tmp/destination/dir"))
		})

		It("it gets the version ID", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("version=1.2.3"))
		})

		It("it returns metadata as list of strings", func() {
			vs := []MetadataPair{
				MetadataPair{Name: "value1", Value: "something quite long"},
				MetadataPair{Name: "value_2", Value: "2"},
			}
			Ω(response.Metadata).Should(BeEquivalentTo(vs))
		})
	})
})
