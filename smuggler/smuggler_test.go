package smuggler_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var pipeline_yml = Fixture("../fixtures/pipeline.yml")
var pipeline = NewPipeline(pipeline_yml)

var logger = log.New(GinkgoWriter, "smuggler: ", log.Lmicroseconds)

var request *ResourceRequest
var response ResourceResponse
var command *SmugglerCommand
var fixtureResourceName string
var requestType RequestType
var requestVersion json.RawMessage
var requestJson string
var err error
var dataDir string

var _ = Describe("Check Command basic tests", func() {
	Context("when given a basic config from a structure", func() {
		request := ResourceRequest{
			Source: SmugglerSource{
				Commands: []CommandDefinition{
					CommandDefinition{
						Name: "check",
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
			Type: CheckType,
		}

		It("it executes the command successfully and captures the output", func() {
			command := NewSmugglerCommand(logger)
			command.RunAction("", request)
			Ω(command.LastCommandOutput).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})

	Context("when given a basic config from a json", func() {
		requestJson := `{
			"source": {
				"commands": [
					{
						"name": "check",
						"path": "sh",
						"args": [ "-e", "-c", "echo basic echo test" ]
					}
				]
			},
			"version": {}
		}`

		BeforeEach(func() {
			request, err = NewResourceRequest(CheckType, requestJson)
			Ω(err).ShouldNot(HaveOccurred())
			command = NewSmugglerCommand(logger)
			response, err = command.RunAction("", *request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("it executes the command successfully and captures the output", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})
})

var _ = Describe("SmugglerCommand actions normal input-output", func() {
	BeforeEach(func() {
		dataDir = "/some/path"
	})
	JustBeforeEach(func() {
		runCommandFromFixture(requestType, dataDir, fixtureResourceName, "1.2.3")
	})

	Context("when calling action 'check'", func() {
		BeforeEach(func() {
			requestType = CheckType
		})

		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it gets the version id", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("version=1.2.3"))
			})
			It("it returns versions as list of strings", func() {
				vs := JsonStringToInterfaceList([]string{"1.2.3", "1.2.4"})
				Ω(response.Versions).Should(BeEquivalentTo(vs))
			})
		})
	})

	Context("When calling action 'in'", func() {
		BeforeEach(func() {
			requestType = InType
		})
		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})

			It("it gets the version id", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("version=1.2.3"))
			})
			It("it gets the destination dir", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("destinationDir=/some/path"))
			})
		})

	})
	Context("When calling action 'out'", func() {
		BeforeEach(func() {
			requestType = OutType
		})
		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it gets the sources dir", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("sourcesDir=/some/path"))
			})
		})
	})
})

var _ = Describe("SmugglerCommand actions stdin/stdout input-output", func() {
	BeforeEach(func() {
		dataDir = "/some/path"
	})
	JustBeforeEach(func() {
		runCommandFromFixture(requestType, dataDir, fixtureResourceName, "1.2.3")
	})

	Context("when calling action 'in'", func() {
		BeforeEach(func() {
			requestType = InType
		})
		Context("when a command reads and dumps the request from stdin", func() {
			BeforeEach(func() {
				dataDir, err = ioutil.TempDir("", "destination_dir")
				Ω(err).ShouldNot(HaveOccurred())

				fixtureResourceName = "dump_request_from_stdin"
			})
			AfterEach(func() {
				os.RemoveAll(dataDir)
			})

			It("the command we find the same request destiation dir", func() {
				b, err := ioutil.ReadFile(filepath.Join(dataDir, "stdin.json"))
				Ω(err).ShouldNot(HaveOccurred())

				b_orig, err := json.Marshal(&request)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(b).Should(MatchJSON(b_orig))

				r, err := NewResourceRequest(request.Type, string(b))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(r).Should(BeEquivalentTo(request))
			})
		})

		Context("when given a config with a command which writes the json response to stdout", func() {
			BeforeEach(func() {
				fixtureResourceName = "write_response_to_stdout"
			})
			It("it returns metadata as list of strings", func() {
				vs := []MetadataPair{
					MetadataPair{Name: "value_from_json", Value: "something"},
					MetadataPair{Name: "other_value_from_json", Value: "otherthing"},
				}
				Ω(response.Metadata).Should(BeEquivalentTo(vs))
			})
			It("it returns the version ID", func() {
				v := JsonStringToInterface("3.2.1")
				Ω(response.Version).Should(Equal(v))
			})
		})
	})
})

var _ = Describe("SmugglerCommand params", func() {
	BeforeEach(func() {
		dataDir = "/some/path"
	})
	JustBeforeEach(func() {
		runCommandFromFixture(requestType, dataDir, fixtureResourceName, "1.2.3")
	})

	Context("when executing a task with `smuggler_params` in the params", func() {
		BeforeEach(func() {
			requestType = InType
			fixtureResourceName = "complex_command"
		})

		It("should get those params as environment variables", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("param6=an additional param"))
		})

	})
})

func runCommandFromFixture(requestType RequestType, dataDir string, fixtureResourceName string, version string) {
	requestJson, err = pipeline.JsonRequest(requestType, fixtureResourceName, "a_job", version)
	Ω(err).ShouldNot(HaveOccurred())

	request, err = NewResourceRequest(requestType, requestJson)
	Ω(err).ShouldNot(HaveOccurred())

	command = NewSmugglerCommand(logger)
	response, err = command.RunAction(dataDir, *request)
}

func CommonSmugglerTests() func() {
	return func() {
		Context("when given a config with empty config from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "dummy_command"
			})
			It("executes without errors", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("does not execute and returns an empty response", func() {
				Ω(command.LastCommand()).Should(BeNil())
				Ω(command.LastCommandOutput).Should(BeEmpty())
				Ω(command.LastCommandSuccess()).Should(BeTrue())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})

			It("executes without errors", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("executes several lines of the script", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("Command Start"))
				Ω(command.LastCommandErr).Should(ContainSubstring("Command End"))
				Ω(command.LastCommandSuccess()).Should(BeTrue())
			})
			It("it sets the $MUGGLER_ACTION and $SMUGGLE_COMMAND variables", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("action=" + string(requestType)))
				Ω(command.LastCommandOutput).Should(ContainSubstring("command=" + string(requestType)))
			})
			It("it can sets the resource smuggler_params as environment variables", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("param1=test"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param2=true"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param3=123"))
			})
		})
		Context("when given a command which fails", func() {
			BeforeEach(func() {
				fixtureResourceName = "fail_command"
			})
			It("returns error", func() {
				Ω(err).Should(HaveOccurred())
			})
			It("captures the exit code", func() {
				Ω(command.LastCommandSuccess()).Should(BeFalse())
				Ω(command.LastCommandExitStatus()).Should(Equal(2))
			})
		})
	}
}

func InOutCommonSmugglerTests() func() {
	return func() {
		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it sets the resource smuggler_params and 'get' params as environment variables", func() {
				Ω(command.LastCommandOutput).Should(ContainSubstring("param1=test"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param2=true"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param3=123"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param4=val4"))
				Ω(command.LastCommandOutput).Should(ContainSubstring("param5=something with spaces"))
			})
			It("it returns metadata as list of strings", func() {
				vs := []MetadataPair{
					MetadataPair{Name: "value1", Value: "something quite long"},
					MetadataPair{Name: "value_2", Value: "2"},
				}
				Ω(response.Metadata).Should(BeEquivalentTo(vs))
			})
			It("it returns the version ID", func() {
				v := JsonStringToInterface("1.2.3")
				Ω(response.Version).Should(BeEquivalentTo(v))
			})
		})
	}
}
