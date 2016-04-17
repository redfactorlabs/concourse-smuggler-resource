package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var _ = Describe("ResourceRequest", func() {
	var pipeline_yml = Fixture("../fixtures/pipeline.yml")
	var pipeline = NewPipeline(pipeline_yml)

	It("Decoding and encoding a string with json results in the same string", func() {
		var r *ResourceRequest
		s := `{"source":{"commands":[{"name":"in","path":"env"}]},"version":"1.2.3"}`

		r, err := NewResourceRequest(InType, s)
		Ω(err).ShouldNot(HaveOccurred())

		b, err := r.ToJson()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(MatchJSON(s))
	})
	It("Adding a srting version with JsonStringToInterface encodes without escaping it", func() {
		r := ResourceRequest{}
		r.Version = JsonStringToInterface("1.2.3")
		b, err := r.ToJson()

		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(BeEquivalentTo(`{"source":{},"version":"1.2.3"}`))
	})
	It("populates the Source.ExtraParams with any non predefined parameter", func() {
		json, err := pipeline.JsonRequest(InType, "mix_params", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		request, err := NewResourceRequest(InType, json)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request.Source.ExtraParams).Should(HaveKey("non_smuggler_param1"))
	})
})
