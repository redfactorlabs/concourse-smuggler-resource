package smuggler

import (
	"encoding/json"
	//	"fmt"
	//	"os"
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
	Source  Source            `json:"source,omitempty"`
	Version interface{}       `json:"version,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Type    RequestType       `json:"-"`
}

// Check if the string is json itself, in which case is parsed and
// return as interface{}. If not, returns the string itself
func JsonStringToInterface(s string) interface{} {
	var r interface{}
	b := []byte(s)

	err := json.Unmarshal(b, &r)
	if err == nil {
		return r
	}
	return s
}

func JsonStringToInterfaceList(sl []string) []interface{} {
	vl := []interface{}{}
	for _, s := range sl {
		vl = append(vl, JsonStringToInterface(s))
	}
	return vl
}

func InterfaceToJsonString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	default:
	}
	s, err := json.Marshal(v)
	if err != nil {
		panic("Error converting version to json. Shouldn't happen :(")
	}
	return string(s)
}

func NewResourceRequestFromJson(jsonString string, requestType RequestType) (ResourceRequest, error) {
	request := ResourceRequest{}
	err := json.NewDecoder(strings.NewReader(jsonString)).Decode(&request)
	request.Type = requestType
	return request, err
}

type ResourceResponse struct {
	Version  interface{}    `json:"version,omitempty"`
	Versions []interface{}  `json:"versions,omitempty"`
	Metadata []MetadataPair `json:"metadata,omitempty"`
	Type     RequestType    `json:-`
}

func (r *ResourceResponse) IsEmpty() bool {
	return r.Version == interface{}(nil) &&
		len(r.Versions) == 0 &&
		len(r.Metadata) == 0
}
