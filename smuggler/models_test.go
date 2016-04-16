package smuggler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var _ = Describe("JsonStringToInterface", func() {
	It("It returns a plain json as json.RawMessage", func() {
		s := `{"a":1,"b":2,"c":{"x":true,"y":false},"d":"ABC"}`
		v := JsonStringToInterface(s)
		Ω(v).ShouldNot(BeEquivalentTo(s))
		m := v.(map[string]interface{})
		Ω(m["a"]).Should(BeEquivalentTo(1))
	})
	It("It returns a non string as a quoted string in a json.RawMessage", func() {
		s := `hello world`
		v := JsonStringToInterface(s)
		Ω(v).Should(BeEquivalentTo(s))
	})
	It("It returns a non valid json as a quoted string in a json.RawMessage", func() {
		s := `{"a":1,"b":2,"c":{"x":true,"y":false},"d":"ABC" ... invalid`
		v := JsonStringToInterface(s)
		Ω(v).Should(BeEquivalentTo(s))
	})
})

var _ = Describe("InterfaceToJsonString", func() {
	It("It returns a plain json as json.RawMessage", func() {
		s := `{"a":1,"b":2,"c":{"x":true,"y":false},"d":"ABC"}`
		v := JsonStringToInterface(s)
		s2 := InterfaceToJsonString(v)
		Ω(s2).Should(MatchJSON(s))
	})
	It("It returns a non string as a quoted string in a json.RawMessage", func() {
		s := `hello world`
		v := JsonStringToInterface(s)
		s2 := InterfaceToJsonString(v)
		Ω(s2).Should(Equal(s))
	})
	It("It returns a non valid json as a quoted string in a json.RawMessage", func() {
		s := `{"a":1,"b":2,"c":{"x":true,"y":false},"d":"ABC" ... invalid`
		v := JsonStringToInterface(s)
		s2 := InterfaceToJsonString(v)
		Ω(s2).Should(Equal(s))
	})
})

var _ = Describe("ResourceRequest", func() {
	It("Decoding and encoding a string with json results in the same string", func() {
		var r ResourceRequest
		s := `{"source":{"commands":[{"name":"in","path":"env"}]},"version":"1.2.3"}`
		err := json.Unmarshal([]byte(s), &r)
		Ω(err).ShouldNot(HaveOccurred())

		b, err := json.Marshal(r)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(MatchJSON(s))
	})
	It("Adding a srting version with JsonStringToInterface encodes without escaping it", func() {
		r := ResourceRequest{}
		r.Version = JsonStringToInterface("1.2.3")
		b, err := json.Marshal(r)

		Ω(err).ShouldNot(HaveOccurred())
		Ω(b).Should(BeEquivalentTo(`{"source":{},"version":"1.2.3"}`))
	})
})
