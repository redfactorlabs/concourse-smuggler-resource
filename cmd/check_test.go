package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var manifest = Fixture("pipeline.yml")
var err error

var _ = Describe("check", func() {
	var (
		command *exec.Cmd
		stdin   *bytes.Buffer
		session *gexec.Session
		logFile *os.File

		expectedExitStatus int
	)

	BeforeEach(func() {
		var err error
		stdin = &bytes.Buffer{}
		expectedExitStatus = 0

		command = exec.Command(checkPath)
		command.Stdin = stdin

		// Point log file to a temporary location
		logFile, err = ioutil.TempFile("", "smuggler.log")
		Ω(err).ShouldNot(HaveOccurred())
		command.Env = append(os.Environ(), fmt.Sprintf("SMUGGLER_LOG=%s", logFile.Name()))
	})

	AfterEach(func() {
		stat, err := logFile.Stat()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(stat.Size()).Should(BeNumerically(">", 0))
		os.Remove(logFile.Name())
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))
	})

	Context("when given a complex command", func() {
		var request ResourceRequest

		BeforeEach(func() {
			request, err = GetResourceRequestFromYamlManifest(CheckType, manifest, "complex_command", "a_job")
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("outputs a valid json with a version", func() {
			var response []Version
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
		var request ResourceRequest

		BeforeEach(func() {
			request, err = GetResourceRequestFromYamlManifest(CheckType, manifest, "dummy_command", "a_job")

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns empty version list", func() {
			var response []Version
			err := json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(response).Should(BeEmpty())
		})
	})

	Context("when given a command which fails", func() {
		var request ResourceRequest

		BeforeEach(func() {
			request, err = GetResourceRequestFromYamlManifest(CheckType, manifest, "fail_command", "a_job")

			expectedExitStatus = 2

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns an error", func() {
			Ω(session.Err).Should(gbytes.Say("error running command"))
		})
	})
})
