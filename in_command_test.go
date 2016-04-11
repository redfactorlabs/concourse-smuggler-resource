package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("In Command", func() {
	It("it gets the destination dir", func() {
		manifest := `
resources:
- name: output_dir
  type: smuggler
  source:
    smuggler_config:
      in:
        path: bash
        args:
        - -e
        - -c
        - |
          echo "destinationDir=$SMUGGLER_DESTINATION_DIR"
`
		source, err := ResourceSourceFromYamlManifest(manifest, "output_dir")
		Ω(err).ShouldNot(HaveOccurred())
		request := InRequest{
			Source:  *source,
			Version: Version{VersionID: "1.2.3"},
		}
		command := NewSmugglerCommand()
		_, err = command.RunIn("/tmp/destination/dir", request)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("destinationDir=/tmp/destination/dir"))
	})

	It("it gets the version ID", func() {
		manifest := `
resources:
- name: output_versions
  type: smuggler
  source:
    smuggler_config:
      in:
        path: bash
        args:
        - -e
        - -c
        - |
          echo "version=$SMUGGLER_VERSION_ID"
`
		source, err := ResourceSourceFromYamlManifest(manifest, "output_versions")
		Ω(err).ShouldNot(HaveOccurred())
		request := InRequest{
			Source:  *source,
			Version: Version{VersionID: "1.2.3"},
		}
		checkCommand := NewSmugglerCommand()
		_, err = checkCommand.RunIn("", request)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("version=1.2.3"))
	})

	It("it returns metadata as list of strings", func() {
		manifest := `
resources:
- name: output_versions
  type: smuggler
  source:
    smuggler_config:
      in:
        path: bash
        args:
        - -e
        - -c
        - |
          echo "value1= something quite long  " > ${SMUGGLER_OUTPUT_DIR}/metadata
          echo -e "\n   " >> ${SMUGGLER_OUTPUT_DIR}/metadata
          echo -e "\t value_2=2  \n" >> ${SMUGGLER_OUTPUT_DIR}/metadata
`
		source, err := ResourceSourceFromYamlManifest(manifest, "output_versions")
		Ω(err).ShouldNot(HaveOccurred())
		request := InRequest{
			Source:  *source,
			Version: Version{VersionID: "1.2.3"},
		}
		checkCommand := NewSmugglerCommand()
		checkResponse, err := checkCommand.RunIn("", request)
		Ω(err).ShouldNot(HaveOccurred())

		vs := []MetadataPair{
			MetadataPair{Name: "value1", Value: "something quite long"},
			MetadataPair{Name: "value_2", Value: "2"},
		}

		Ω(checkResponse.Metadata).Should(BeEquivalentTo(vs))
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
      in:
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
`
		source, err := ResourceSourceFromYamlManifest(manifest, "pass_params")
		Ω(err).ShouldNot(HaveOccurred())
		request := InRequest{
			Source:  *source,
			Version: Version{VersionID: "1.2.3"},
			Params: map[string]string{
				"param4": "val4",
				"param5": "something with spaces",
			},
		}
		checkCommand := NewSmugglerCommand()
		checkCommand.RunIn("", request)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
	})
})
