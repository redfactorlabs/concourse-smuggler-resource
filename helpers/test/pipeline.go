package test

import (
	"fmt"

	"encoding/json"
	"github.com/ghodss/yaml"

	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

type Pipeline struct {
	Resources []Resource `json:"resources"`
	Jobs      []Job      `json:"jobs"`
}

type Resource struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Source map[string]interface{} `json:"source"`
}

type Job struct {
	Name string `json:"name"`
	Plan []Task `json:"plan"`
}

type Task struct {
	GetName string                 `json:"get"`
	PutName string                 `json:"put"`
	Params  map[string]interface{} `json:"params"`
}

func NewPipeline(yaml_manifest string) *Pipeline {
	var pipeline Pipeline
	err := yaml.Unmarshal([]byte(yaml_manifest), &pipeline)
	if err != nil {
		panic(err)
	}
	return &pipeline
}

func JsonRequestFromYaml(yaml_source string) ([]byte, error) {
	var i interface{}
	var b []byte
	err := yaml.Unmarshal([]byte(yaml_source), i)
	if err != nil {
		return b, fmt.Errorf("Failed yaml2json %q: %+v", err, i)
	}
	b, err = json.Marshal(i)
	if err != nil {
		return b, fmt.Errorf("Failed yaml2json %q: %+v", err, i)
	}
	return b, nil
}

func (pipeline *Pipeline) JsonRequest(requestType RequestType, resource_name string, job_name string, version string) (string, error) {
	var resource *Resource
	var request RawResourceRequest

	resource = nil
	for _, r := range pipeline.Resources {
		if r.Name == resource_name {
			resource = &r
			break
		}
	}
	if resource == nil {
		return "", fmt.Errorf("Cannot find a resource called '%s' in the pipeline.", resource_name)
	}
	request.Source = resource.Source

	for _, j := range pipeline.Jobs {
		if j.Name == job_name {
			for _, t := range j.Plan {
				if requestType == InType && t.GetName == resource_name {
					request.Params = t.Params
				}
				if requestType == OutType && t.PutName == resource_name {
					request.Params = t.Params
				}
			}
		}
	}

	if requestType == InType || requestType == CheckType {
		v, err := NewVersion(version)
		if err != nil {
			return "", fmt.Errorf("Failed encoding version %q: %+v", err, request)
		}
		request.Version = *v
	}

	b, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("Failed encoding request %q: %+v", err, request)
	}

	return string(b), nil
}
