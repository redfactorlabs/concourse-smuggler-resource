package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
)

var _ = Describe("In Command", func() {
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
		checkResponse, err := checkCommand.RunIn(request)
		Ω(err).ShouldNot(HaveOccurred())

		vs := []MetadataPair{
			MetadataPair{Name: "value1", Value: "something quite long"},
			MetadataPair{Name: "value_2", Value: "2"},
		}

		Ω(checkResponse.Metadata).Should(BeEquivalentTo(vs))
	})
})
