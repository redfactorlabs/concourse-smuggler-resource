package check_test

import (
	"fmt"
	"github.com/ghodss/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
	. "github.com/redfactorlabs/concourse-smuggler-resource/check"
)

func checkRequestFromYaml(yaml_manifest string, resource_name string) (*CheckRequest, error) {
	var resourceDefinition ResourceDefinition
	err := yaml.Unmarshal([]byte(yaml_manifest), &resourceDefinition)
	if err != nil {
		return nil, err
	}

	var resource *Resource
	for _, r := range resourceDefinition.Resources {
		if r.Name == resource_name {
			resource = &r
		}
	}
	if resource == nil {
		return nil, fmt.Errorf("Cannot find a resource called '%s' in the yaml definition.", resource_name)
	}

	var checkRequest = CheckRequest{
		Source: resource.Source,
	}

	return &checkRequest, nil

}

var _ = Describe("Check Command", func() {
	It("executes a basic echo command", func() {
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
	})
	It("executes a basic echo command from json", func() {
		requestBasicEcho, err := NewCheckRequestFromJson(requestBasicEchoJson)
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
	})
	It("executes a basic echo command from json and returns the output", func() {
		requestBasicEcho, err := NewCheckRequestFromJson(requestBasicEchoJson)
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewCheckCommand()
		checkCommand.Run(requestBasicEcho)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})
	It("executes a basic echo command from yaml manifest and returns the output", func() {
		requestBasicEcho, err := checkRequestFromYaml(resourceDefinitionBasicEcho, "simple_echo")
		Ω(err).ShouldNot(HaveOccurred())
		checkCommand := NewCheckCommand()
		checkCommand.Run(*requestBasicEcho)
		Ω(checkCommand.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
	})
})

type ResourceDefinition struct {
	Resources []Resource `json: resources`
}

type Resource struct {
	Name   string `json:name`
	Type   string `json:type`
	Source Source `json:source`
}

var requestBasicEcho = CheckRequest{
	Source: Source{
		SmugglerConfig: SmugglerConfig{
			CheckCommand: CommandDefinition{
				Path: "bash",
				Args: []string{"-e", "-c", "echo basic echo test"},
			},
		},
	},
}

var requestBasicEchoJson = `
{
  "source": {
    "smuggler_config": {
      "check": {
	"path": "sh",
	"args": [ "-e", "-c", "echo basic echo test" ]
      }
    }
  },
  "version": {}
}
`

var resourceDefinitionBasicEcho = `
resources:
- name: simple_echo
  type: smuggler
  source:
    smuggler_config:
      check:
        path: sh
        args:
        - -e
        - -c
        - |
          echo basic echo test
`
