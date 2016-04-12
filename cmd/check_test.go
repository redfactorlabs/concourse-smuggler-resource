package cmd_test

import (
	"bytes"
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var manifest = Fixture("pipeline.yml")

var _ = Describe("check", func() {
	var (
		command *exec.Cmd
		stdin   *bytes.Buffer
		session *gexec.Session

		expectedExitStatus int
	)

	BeforeEach(func() {
		stdin = &bytes.Buffer{}
		expectedExitStatus = 0

		command = exec.Command(checkPath)
		command.Stdin = stdin
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))
	})

	Context("when given a complex command", func() {
		var request CheckRequest

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("outputs a valid json with a version", func() {
			var response CheckResponse
			err := json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			vs := []Version{Version{VersionID: "1.2.3"}, Version{VersionID: "1.2.4"}}
			Ω(response).Should(BeEquivalentTo(vs))
		})

		It("outputs the commands output", func() {
			stderr := session.Err.Contents()

			Ω(stderr).Should(ContainSubstring("Command Start"))
			Ω(stderr).Should(ContainSubstring("Command End"))
			Ω(stderr).Should(ContainSubstring("param1=test"))
			Ω(stderr).Should(ContainSubstring("param2=true"))
			Ω(stderr).Should(ContainSubstring("param3=123"))
		})
	})

	Context("when given a dummy command", func() {
		var request CheckRequest

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "dummy_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns empty version list", func() {
			var response CheckResponse
			err := json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(response).Should(BeEmpty())
		})
	})

	Context("when given a command which fails", func() {
		var request CheckRequest

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "fail_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}

			expectedExitStatus = 2

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns an error", func() {
			Ω(session.Err).Should(gbytes.Say("error running command"))
		})
	})
})
