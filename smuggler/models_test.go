package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var _ = Describe("ResourceRequest", func() {
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
})
