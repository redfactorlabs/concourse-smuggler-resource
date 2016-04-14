package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var manifest = Fixture("pipeline.yml")
var err error

var _ = Describe("smuggler commands", func() {
	var (
		session *gexec.Session
		logFile *os.File

		commandPath        string
		dataDir            string
		expectedExitStatus int
		request            ResourceRequest
	)

	BeforeEach(func() {
		expectedExitStatus = 0
		dataDir = ""
	})

	JustBeforeEach(func() {
		var err error
		var command *exec.Cmd

		stdin := &bytes.Buffer{}
		err = json.NewEncoder(stdin).Encode(request)
		Ω(err).ShouldNot(HaveOccurred())

		if dataDir == "" {
			command = exec.Command(commandPath)
		} else {
			command = exec.Command(commandPath, dataDir)
		}
		command.Stdin = stdin

		// Point log file to a temporary location
		logFile, err = ioutil.TempFile("", "smuggler.log")
		Ω(err).ShouldNot(HaveOccurred())
		command.Env = append(os.Environ(), fmt.Sprintf("SMUGGLER_LOG=%s", logFile.Name()))

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))

	})

	AfterEach(func() {
		stat, err := logFile.Stat()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(stat.Size()).Should(BeNumerically(">", 0))
		os.Remove(logFile.Name())
	})

	Context("when given a complex definition", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				commandPath = checkPath

				commandPath, request = prepareCommandCheck("complex_command")
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
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandIn("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandOut("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
	})

	Context("when given a dummy command", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				commandPath, request = prepareCommandCheck("dummy_command")
			})

			It("returns empty version list", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEmpty())
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandIn("dummy_command")
			})
			It("returns empty response", func() {
				var response ResourceResponse
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandOut("dummy_command")
			})
			It("returns empty response", func() {
				var response ResourceResponse
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
	})

	Context("when given a command which fails", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, request = prepareCommandCheck("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, request = prepareCommandIn("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, request = prepareCommandOut("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
	})

})

func prepareCommandCheck(manifestDefinitionName string) (string, ResourceRequest) {
	commandPath := checkPath

	request, err := GetResourceRequestFromYamlManifest(CheckType, manifest, manifestDefinitionName, "a_job")
	Ω(err).ShouldNot(HaveOccurred())

	return commandPath, request
}

func prepareCommandIn(manifestDefinitionName string) (string, string, ResourceRequest) {
	commandPath := inPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	request, err := GetResourceRequestFromYamlManifest(InType, manifest, manifestDefinitionName, "a_job")
	Ω(err).ShouldNot(HaveOccurred())

	return commandPath, dataDir, request
}

func prepareCommandOut(manifestDefinitionName string) (string, string, ResourceRequest) {
	commandPath := outPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	request, err := GetResourceRequestFromYamlManifest(OutType, manifest, manifestDefinitionName, "a_job")
	Ω(err).ShouldNot(HaveOccurred())

	return commandPath, dataDir, request
}

func InOutCommonSmugglerTests(session **gexec.Session) func() {
	return func() {
		It("outputs a valid json with a version", func() {
			var response ResourceResponse
			err := json.Unmarshal((*session).Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(response.Version).Should(Equal(Version{VersionID: "1.2.3"}))
		})
		It("outputs a valid json with a version", func() {
			var response ResourceResponse
			err := json.Unmarshal((*session).Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			expectedMetadata := []MetadataPair{
				MetadataPair{Name: "value1", Value: "something quite long"},
				MetadataPair{Name: "value_2", Value: "2"},
			}
			Ω(response.Metadata).Should(Equal(expectedMetadata))
		})
		It("outputs the commands output", func() {
			stderr := (*session).Err.Contents()

			Ω(stderr).Should(ContainSubstring("Command Start"))
			Ω(stderr).Should(ContainSubstring("Command End"))
			Ω(stderr).Should(ContainSubstring("param1=test"))
			Ω(stderr).Should(ContainSubstring("param2=true"))
			Ω(stderr).Should(ContainSubstring("param3=123"))
			Ω(stderr).Should(ContainSubstring("param4=val4"))
			Ω(stderr).Should(ContainSubstring("param5=something with spaces"))
		})
	}
}
