package test

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"

	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

type Manifest struct {
	Resources []Resource `json: resources`
	Jobs      []Job      `json: jobs`
}

type Resource struct {
	Name   string `json:name`
	Type   string `json:type`
	Source Source `json:source`
}

type Job struct {
	Name string `json:name`
	Plan []Task `json:plan`
}

type Task struct {
	GetName string                 `json:"get"`
	PutName string                 `json:"put"`
	Params  map[string]interface{} `json:params`
}

func ManifestFromYaml(yaml_manifest string) (Manifest, error) {
	var manifest Manifest
	err := yaml.Unmarshal([]byte(yaml_manifest), &manifest)
	return manifest, err
}

func GetResourceRequestFromYamlManifest(requestType RequestType, yaml_manifest string, resource_name string, job_name string) (ResourceRequest, error) {
	resourceRequest := ResourceRequest{Type: requestType}

	manifest, err := ManifestFromYaml(yaml_manifest)
	if err != nil {
		return resourceRequest, err
	}
	var resource *Resource
	for _, r := range manifest.Resources {
		if r.Name == resource_name {
			resource = &r
			break
		}
	}
	if resource == nil {
		return resourceRequest, fmt.Errorf("Cannot find a resource called '%s' in the yaml definition.", resource_name)
	}
	resourceRequest.Source = resource.Source

	for _, j := range manifest.Jobs {
		if j.Name == job_name {
			for _, t := range j.Plan {
				if requestType == InType && t.GetName == resource_name {
					resourceRequest.Params = t.Params
				}
				if requestType == OutType && t.PutName == resource_name {
					resourceRequest.Params = t.Params
				}
			}
		}
	}

	if requestType == InType || requestType == CheckType {
		resourceRequest.Version = JsonStringToInterface("1.2.3")
	}

	return resourceRequest, nil
}

func ParamsFromYamlManifest(yaml_manifest string, resource_name string) (*Source, error) {
	manifest, err := ManifestFromYaml(yaml_manifest)
	if err != nil {
		return nil, err
	}
	var resource *Resource
	for _, r := range manifest.Resources {
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

func Fixture(path string) string {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
