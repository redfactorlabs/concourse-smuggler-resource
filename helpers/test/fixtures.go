package helpers

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"path/filepath"

	. "github.com/redfactorlabs/concourse-smuggler-resource"
)

type ResourceDefinition struct {
	Resources []Resource `json: resources`
}

type Resource struct {
	Name   string `json:name`
	Type   string `json:type`
	Source Source `json:source`
}

func ResourceSourceFromYamlManifest(yaml_manifest string, resource_name string) (*Source, error) {
	var resourceDefinition ResourceDefinition
	err := yaml.Unmarshal([]byte(yaml_manifest), &resourceDefinition)
	if err != nil {
		return nil, err
	}

	var resource *Resource
	for _, r := range resourceDefinition.Resources {
		if r.Name == resource_name {
			resource = &r
			break
		}
	}
	if resource == nil {
		return nil, fmt.Errorf("Cannot find a resource called '%s' in the yaml definition.", resource_name)
	}

	return &resource.Source, nil
}

func Fixture(filename string) string {
	path := filepath.Join("fixtures", filename)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
