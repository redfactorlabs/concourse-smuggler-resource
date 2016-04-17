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
		s := `{"source":{"commands":{"in":{"path":"env"}}},"version":{"ref":"1.2.3"},"params":{}}`

		r, err := NewResourceRequest(InType, s)
		Ω(err).ShouldNot(HaveOccurred())

		b, err := r.ToJson()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(MatchJSON(s))
	})
	It("Adding a string version with NewVersion encodes without escaping it", func() {
		r := ResourceRequest{}
		v, err := NewVersion("1.2.3")
		Ω(err).ShouldNot(HaveOccurred())
		r.Version = *v
		b, err := r.ToJson()

		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(MatchJSON(`{"source":{},"version":{"ref": "1.2.3"},"params":{}}`))
	})
	It("populates the Source.ExtraParams with any additional parameter", func() {
		json, err := pipeline.JsonRequest(InType, "mix_params", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		request, err := NewResourceRequest(InType, json)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request.Source.ExtraParams).Should(HaveKey("non_smuggler_param1"))
	})
	It("populates the Params.ExtraParams with any additional parameter", func() {
		json, err := pipeline.JsonRequest(InType, "mix_params", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		request, err := NewResourceRequest(InType, json)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request.Params.ExtraParams).Should(HaveKey("non_smuggler_param2"))
	})
	It("populates the OrigRequest attribute", func() {
		json, err := pipeline.JsonRequest(InType, "mix_params", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		request, err := NewResourceRequest(InType, json)
		Ω(err).ShouldNot(HaveOccurred())

		rawRequest, err := NewRawResourceRequest(json)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request.OrigRequest).Should(BeEquivalentTo(rawRequest))
	})
	It("populates the FilteredRequest attribute with a RawRequest without smuggler config", func() {
		json, err := pipeline.JsonRequest(InType, "mix_params", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		request, err := NewResourceRequest(InType, json)
		Ω(err).ShouldNot(HaveOccurred())

		rawJson, err := pipeline.JsonRequest(InType, "mix_params_filtered", "a_job", "1.2.3")
		Ω(err).ShouldNot(HaveOccurred())

		rawRequest, err := NewRawResourceRequest(rawJson)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request.FilteredRequest).Should(BeEquivalentTo(rawRequest))
	})
})
