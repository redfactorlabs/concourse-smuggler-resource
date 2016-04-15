package smuggler

import (
	"encoding/json"
	"strings"
)

type Source struct {
	Commands    []CommandDefinition `json:"commands,omitempty"`
	ExtraParams map[string]string   `json:"extra_params,omitempty"`
}

func (source Source) IsValid() (bool, string) {
	return true, ""
}

func (source Source) FindCommand(name string) *CommandDefinition {
	for _, command := range source.Commands {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

// Merges two configuration Source.
// * Commands: get merged by key 'name'. sourceB overrides sourceA
// * ExtraParams: gets merged by key. sourceB overrides sourceA
func MergeSource(sourceA, sourceB *Source) *Source {
	var newSource Source

	newSource.Commands = make([]CommandDefinition, 0, 6)
	for _, command := range sourceB.Commands {
		newSource.Commands = append(newSource.Commands, command)
	}
	for _, command := range sourceA.Commands {
		if newSource.FindCommand(command.Name) == nil {
			newSource.Commands = append(newSource.Commands, command)
		}
	}
	newSource.ExtraParams = make(map[string]string)
	for k, v := range sourceA.ExtraParams {
		newSource.ExtraParams[k] = v
	}
	for k, v := range sourceB.ExtraParams {
		newSource.ExtraParams[k] = v
	}
	return &newSource
}

type CommandDefinition struct {
	Name string   `json:"name"`
	Path string   `json:"path"`
	Args []string `json:"args,omitempty"`
}

func (commandDefinition CommandDefinition) IsDefined() bool {
	return (commandDefinition.Name != "")
}

type Version struct {
	VersionID string `json:"version_id,omitempty"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RequestType string

func (t RequestType) Name() string {
	return string(t)
}

const (
	CheckType RequestType = "check"
	InType    RequestType = "in"
	OutType   RequestType = "out"
)

type ResourceRequest struct {
	Source  Source            `json:"source"`
	Version Version           `json:"version"`
	Params  map[string]string `json:"params,omitempty"`
	Type    RequestType       `json:-`
}

func NewResourceRequestFromJson(jsonString string, requestType RequestType) (ResourceRequest, error) {
	request := ResourceRequest{}
	err := json.NewDecoder(strings.NewReader(jsonString)).Decode(&request)
	request.Type = requestType
	return request, err
}

type ResourceResponse struct {
	Version  Version        `json:"version,omitempty"`
	Versions []Version      `json:"versions,omitempty"`
	Metadata []MetadataPair `json:"metadata,omitempty"`
	Type     RequestType    `json:"-"`
}

func (r *ResourceResponse) IsEmpty() bool {
	return r.Version.VersionID == "" &&
		len(r.Versions) == 0 &&
		len(r.Metadata) == 0
}
