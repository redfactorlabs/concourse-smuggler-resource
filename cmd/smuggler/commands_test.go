package main_test

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

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
)

var pipeline_yml = Fixture("../../fixtures/pipeline.yml")
var pipeline = NewPipeline(pipeline_yml)
var err error

var _ = Describe("smuggler commands", func() {
	var (
		session *gexec.Session
		logFile *os.File

		commandPath        string
		dataDir            string
		expectedExitStatus int
		jsonRequest        string
	)

	BeforeEach(func() {
		expectedExitStatus = 0
		dataDir = ""
	})

	JustBeforeEach(func() {
		var err error
		var command *exec.Cmd

		RegisterFailHandler(Fail)

		if dataDir == "" {
			command = exec.Command(commandPath)
		} else {
			command = exec.Command(commandPath, dataDir)
		}
		command.Stdin = bytes.NewBuffer([]byte(jsonRequest))

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

				commandPath, jsonRequest = prepareCommandCheck("complex_command")
			})
			It("outputs a valid json with a version", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs, err := NewVersions([]string{"1.2.3", "1.2.4"})
				Ω(err).ShouldNot(HaveOccurred())
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
				commandPath, dataDir, jsonRequest = prepareCommandIn("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, jsonRequest = prepareCommandOut("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
	})

	Context("when given a dummy command", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("dummy_command")
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
				commandPath, dataDir, jsonRequest = prepareCommandIn("dummy_command")
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
				commandPath, dataDir, jsonRequest = prepareCommandOut("dummy_command")
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
				commandPath, jsonRequest = prepareCommandCheck("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, jsonRequest = prepareCommandIn("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, jsonRequest = prepareCommandOut("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
	})

	Context("when there is local config file 'smuggler.yml' that is empty", func() {
		BeforeEach(func() {
			err := utils.Copy("../../fixtures/empty_smuggler.yml",
				filepath.Join(filepath.Dir(checkPath), "smuggler.yml"))
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("when running 'check' with a dummy definition", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("dummy_command")
			})

			It("returns empty version list", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEmpty())
			})
		})

		Context("when running 'check' with a complex_command definition", func() {
			BeforeEach(func() {
				commandPath = checkPath
				commandPath, jsonRequest = prepareCommandCheck("complex_command")
			})
			It("outputs a valid json with a version", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs, err := NewVersions([]string{"1.2.3", "1.2.4"})
				Ω(err).ShouldNot(HaveOccurred())
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
	})

	Context("when there is local config file 'smuggler.yml' with config", func() {
		BeforeEach(func() {
			err := utils.Copy("../../fixtures/full_smuggler.yml",
				filepath.Join(filepath.Dir(checkPath), "smuggler.yml"))
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("when running 'check' with a empty command definition", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("dummy_command")
			})

			It("returns versions of the config file", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs, err := NewVersions([]string{"4.5.6", "4.5.7"})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output from the command definition", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("config_param1=param_in_config"))
				Ω(stderr).Should(ContainSubstring("param1=undef"))
				Ω(stderr).Should(ContainSubstring("from config file"))
				Ω(stderr).ShouldNot(ContainSubstring("Command Start"))
			})
		})

		Context("when running 'check' with a complex command definition", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("complex_command")
			})

			It("returns versions of the definition", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs, err := NewVersions([]string{"1.2.3", "1.2.4"})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output from the definition", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("Command Start"))
				Ω(stderr).Should(ContainSubstring("Command End"))
				Ω(stderr).Should(ContainSubstring("param1=test"))
				Ω(stderr).Should(ContainSubstring("param2=true"))
				Ω(stderr).Should(ContainSubstring("param3=123"))
				Ω(stderr).ShouldNot(ContainSubstring("from config file"))
			})
		})

		Context("when running 'check' with a definition with only 'in' command defined", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("test_merge_with_smuggler_yml")
			})
			It("returns versions of the config file", func() {
				var response []Version
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs, err := NewVersions([]string{"4.5.6", "4.5.7"})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output from the config file definition", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("config_param1=param_in_config"))
				Ω(stderr).Should(ContainSubstring("param1=param1"))
				Ω(stderr).Should(ContainSubstring("from config file"))
				Ω(stderr).ShouldNot(ContainSubstring("Command Start"))
			})
		})
		Context("when running 'in' with a definition with only 'in' command defined", func() {
			BeforeEach(func() {
				commandPath, dataDir, jsonRequest = prepareCommandIn("test_merge_with_smuggler_yml")
			})
			It("outputs the commands output from the pipeline definition", func() {
				stderr := session.Err.Contents()
				Ω(stderr).Should(ContainSubstring("from pipeline"))
			})
		})

	})

	Context("when running a quiet command", func() {
		Context("when running 'check'", func() {
			BeforeEach(func() {
				commandPath, jsonRequest = prepareCommandCheck("a_quiet_command")
			})
			It("There are no messages in Stderr", func() {
				Ω(session.Err.Contents()).To(BeEmpty())
			})
		})
	})

})

func getJsonRequest(t RequestType, resourceName string) string {
	jsonRequest, err := pipeline.JsonRequest(t, resourceName, "a_job", "1.2.3")
	Ω(err).ShouldNot(HaveOccurred())

	return jsonRequest
}

func prepareCommandCheck(resourceName string) (string, string) {
	commandPath := checkPath

	jsonRequest := getJsonRequest(CheckType, resourceName)

	return commandPath, jsonRequest
}

func prepareCommandIn(resourceName string) (string, string, string) {
	commandPath := inPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	jsonRequest := getJsonRequest(InType, resourceName)

	return commandPath, dataDir, jsonRequest
}

func prepareCommandOut(resourceName string) (string, string, string) {
	commandPath := outPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	jsonRequest := getJsonRequest(OutType, resourceName)

	return commandPath, dataDir, jsonRequest
}

func InOutCommonSmugglerTests(session **gexec.Session) func() {
	return func() {
		It("outputs a valid json with a version", func() {
			var response ResourceResponse
			err := json.Unmarshal((*session).Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			v, err := NewVersion("1.2.3")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(response.Version).Should(BeEquivalentTo(*v))
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
