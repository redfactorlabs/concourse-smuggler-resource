package smuggler_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
)

var pipeline_yml = Fixture("../fixtures/pipeline.yml")
var pipeline = NewPipeline(pipeline_yml)

var logger = log.New(GinkgoWriter, "smuggler: ", log.Lmicroseconds)

var request *ResourceRequest
var response *ResourceResponse
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
				Commands: map[string]interface{}{
					"check": CommandDefinition{
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
			Type: CheckType,
		}

		It("it executes the command successfully and captures the output", func() {
			command := NewSmugglerCommand(logger)
			command.RunAction("", &request)
			Ω(command.LastCommandOutput).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})

	Context("when given a basic config from a json", func() {
		requestJson := `{
			"source": {
				"commands": {
					"check": {
						"path": "sh",
						"args": [ "-e", "-c", "echo basic echo test" ]
					}
				}
			},
			"version": {}
		}`

		BeforeEach(func() {
			request, err = NewResourceRequest(CheckType, requestJson)
			Ω(err).ShouldNot(HaveOccurred())
			command = NewSmugglerCommand(logger)
			response, err = command.RunAction("", request)
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
				vs, err := NewVersions([]string{"1.2.3", "1.2.4"})
				Ω(err).ShouldNot(HaveOccurred())
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

	Context("when calling action 'check'", func() {
		BeforeEach(func() {
			requestType = CheckType
		})
		Context("when given a config with a command which writes the json response to stdout", func() {
			BeforeEach(func() {
				fixtureResourceName = "write_response_to_stdout"
			})
			It("it returns the version IDs", func() {
				vs, err := NewVersions([]string{"3.2.1", "3.2.2"})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.Versions).Should(Equal(vs))
			})
			It("The stdout buffer is cleared", func() {
				Ω(command.LastCommandOutput).Should(BeEmpty())
			})
		})
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

			It("we find the same request destination dir", func() {
				b, err := ioutil.ReadFile(filepath.Join(dataDir, "stdin.json"))
				Ω(err).ShouldNot(HaveOccurred())

				Ω(b).Should(MatchJSON(requestJson))

				r, err := NewResourceRequest(request.Type, string(b))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(r).Should(BeEquivalentTo(request))

				Ω(r.Source.Commands).ShouldNot(BeEmpty())
				Ω(r.Source.FilterRawRequest).ShouldNot(BeTrue())
				Ω(r.Source.SmugglerParams).ShouldNot(BeEmpty())
				Ω(r.Source.ExtraParams).ShouldNot(BeEmpty())
				Ω(r.Params.SmugglerParams).ShouldNot(BeEmpty())
				Ω(r.Params.ExtraParams).ShouldNot(BeEmpty())
			})
		})

		Context("when a command reads and dumps the filtered request from stdin", func() {
			BeforeEach(func() {
				dataDir, err = ioutil.TempDir("", "destination_dir")
				Ω(err).ShouldNot(HaveOccurred())

				fixtureResourceName = "dump_filtered_request_from_stdin"
			})
			AfterEach(func() {
				os.RemoveAll(dataDir)
			})

			It("we find a filtered request in the destination dir", func() {
				b, err := ioutil.ReadFile(filepath.Join(dataDir, "stdin.json"))
				Ω(err).ShouldNot(HaveOccurred())

				Ω(b).ShouldNot(MatchJSON(requestJson))

				b_filtered, err := json.Marshal(&request.FilteredRequest)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(b).Should(MatchJSON(b_filtered))

				r, err := NewResourceRequest(request.Type, string(b))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(r.OrigRequest).Should(BeEquivalentTo(request.FilteredRequest))

				Ω(r.Source.Commands).Should(BeEmpty())
				// default if missing is false
				Ω(r.Source.FilterRawRequest).Should(BeFalse())
				Ω(r.Source.SmugglerParams).Should(BeEmpty())
				Ω(r.Source.ExtraParams).ShouldNot(BeEmpty())
				Ω(r.Params.SmugglerParams).Should(BeEmpty())
				Ω(r.Params.ExtraParams).ShouldNot(BeEmpty())
			})
			It("we the filtered request has not any of the smuggler fields", func() {
				b, err := ioutil.ReadFile(filepath.Join(dataDir, "stdin.json"))
				Ω(err).ShouldNot(HaveOccurred())

				r, err := NewRawResourceRequest(string(b))
				Ω(err).ShouldNot(HaveOccurred())

				for _, t := range utils.ListJsonTagsOfStruct(request.Source) {
					Ω(r.Source).ShouldNot(HaveKey(t))
				}
				for _, t := range utils.ListJsonTagsOfStruct(request.Params) {
					Ω(r.Params).ShouldNot(HaveKey(t))
				}
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
				v, err := NewVersion("3.2.1")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.Version).Should(Equal(*v))
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

	Context("when executing a task with a mix of `smuggler_params` in the params", func() {
		BeforeEach(func() {
			requestType = InType
			fixtureResourceName = "mix_params"
		})
		It("should get all the params as environment variables", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param1=smuggler_val1"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param2=smuggler_val2"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("non_smuggler_param1=non_smuggler_val1"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("non_smuggler_param2=non_smuggler_val2"))
		})
	})
	Context("when executing a task which handles json in params and versions", func() {
		BeforeEach(func() {
			requestType = InType
			fixtureResourceName = "json_in_params_and_versions"
		})
		It("should get the json param as a serialized json", func() {
			expectedJson := `{
				"with": "keys",
        "and": [ "other", "complex" ],
        "structures": { "like": "this" }
			}`

			m := make(map[string]string)
			for _, l := range strings.Split(string(command.LastCommandOutput), "\n") {
				l := strings.SplitN(l, "=", 2)
				if len(l) >= 2 {
					k, v := l[0], l[1]
					m[k] = v
				}
			}
			Ω(m).Should(HaveKey("complex_param"))
			Ω(m["complex_param"]).Should(MatchJSON(expectedJson))
		})
		It("should send the version param as a serialized json", func() {
			v, err := NewVersion("someid")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(response.Version).Should(BeEquivalentTo(*v))
		})
	})

	Context("when executing a task with params being overridden", func() {
		BeforeEach(func() {
			requestType = InType
			fixtureResourceName = "override_params"
		})
		It("parameters should be overridden in the order: source.smuggler_params source params.smuggler_params params ", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param1=params"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param2=params.smuggler_params"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param3=source"))
			Ω(command.LastCommandOutput).Should(ContainSubstring("smuggler_param4=source.smuggler_params"))
		})
	})

})

var _ = Describe("SmugglerCommand one line commands", func() {
	BeforeEach(func() {
		dataDir = "/some/path"
	})
	JustBeforeEach(func() {
		runCommandFromFixture(requestType, dataDir, fixtureResourceName, "1.2.3")
	})

	Context("When executing a in command", func() {
		BeforeEach(func() {
			requestType = CheckType
			fixtureResourceName = "one_line_commands"
		})
		It("should execute the one line version of the command", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("Inline command for check"))
		})
	})
	Context("When executing a in command", func() {
		BeforeEach(func() {
			requestType = InType
			fixtureResourceName = "one_line_commands"
		})
		It("should execute the one line version of the command", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("Inline command for in"))
		})
	})
	Context("When executing a out command", func() {
		BeforeEach(func() {
			requestType = OutType
			fixtureResourceName = "one_line_commands"
		})
		It("should execute the one line version of the command", func() {
			Ω(command.LastCommandOutput).Should(ContainSubstring("Inline command for out"))
		})
	})

})

var _ = Describe("SmugglerCommand named versions", func() {
	JustBeforeEach(func() {
		runCommandFromFixture(InType, "/some/path", "version_with_names", `{"foo": "foo_version", "bar": "bar_version"}`)
	})

	It("should get the versions prefixed by the key name", func() {
		Ω(command.LastCommandOutput).Should(ContainSubstring("foo=foo_version"))
		Ω(command.LastCommandOutput).Should(ContainSubstring("bar=bar_version"))
	})
})

func runCommandFromFixture(requestType RequestType, dataDir string, fixtureResourceName string, version string) {
	requestJson, err = pipeline.JsonRequest(requestType, fixtureResourceName, "a_job", version)
	Ω(err).ShouldNot(HaveOccurred())

	request, err = NewResourceRequest(requestType, requestJson)
	Ω(err).ShouldNot(HaveOccurred())

	command = NewSmugglerCommand(logger)
	response, err = command.RunAction(dataDir, request)
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
				v, err := NewVersion("1.2.3")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.Version).Should(BeEquivalentTo(*v))
			})
		})
	}
}
