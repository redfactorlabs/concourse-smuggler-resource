package smuggler

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
)

type SmugglerSource struct {
	CheckCommand     string                 `json:"check_command,omitempty"`
	InCommand        string                 `json:"in_command,omitempty"`
	OutCommand       string                 `json:"out_command,omitempty"`
	Commands         []CommandDefinition    `json:"commands,omitempty"`
	FilterRawRequest bool                   `json:"filter_raw_request,omitempty"`
	SmugglerParams   map[string]interface{} `json:"smuggler_params,omitempty"`
	ExtraParams      map[string]interface{} `json:"-"`
}

func WrapCommandWithShell(name string, commandLine string) *CommandDefinition {
	// Try to find bash
	shellPath, err := exec.LookPath("bash")
	if err == nil {
		return &CommandDefinition{
			Name: name,
			Path: shellPath,
			Args: []string{"-e", "-u", "-o", "pipefail", "-c", commandLine},
		}
	}
	// Try to find sh
	shellPath, err = exec.LookPath("sh")
	if err == nil {
		return &CommandDefinition{
			Name: name,
			Path: shellPath,
			Args: []string{"-e", "-u", "-o", "pipefail", "-c", commandLine},
		}
	}

	// In last case, use the command itself
	l := strings.Split(commandLine, ",")
	return &CommandDefinition{
		Name: name,
		Path: l[0],
		Args: l[1:],
	}
}

func (source SmugglerSource) FindCommand(name string) *CommandDefinition {
	if name == "check" && source.CheckCommand != "" {
		return WrapCommandWithShell(name, source.CheckCommand)
	}
	if name == "in" && source.InCommand != "" {
		return WrapCommandWithShell(name, source.InCommand)
	}
	if name == "out" && source.OutCommand != "" {
		return WrapCommandWithShell(name, source.OutCommand)
	}
	for _, command := range source.Commands {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

// Merges two configuration Sources
// * Commands: get merged by key 'name'. sourceB overrides sourceA
// * SmugglerParams: gets merged by key. sourceB overrides sourceA
func MergeSource(sourceA, sourceB *SmugglerSource) *SmugglerSource {
	var newSource SmugglerSource

	newSource.Commands = make([]CommandDefinition, 0, 6)
	for _, command := range sourceB.Commands {
		newSource.Commands = append(newSource.Commands, command)
	}
	for _, command := range sourceA.Commands {
		if newSource.FindCommand(command.Name) == nil {
			newSource.Commands = append(newSource.Commands, command)
		}
	}
	newSource.SmugglerParams = make(map[string]interface{})
	for k, v := range sourceA.SmugglerParams {
		newSource.SmugglerParams[k] = v
	}
	for k, v := range sourceB.SmugglerParams {
		newSource.SmugglerParams[k] = v
	}
	newSource.ExtraParams = make(map[string]interface{})
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

type RawResourceRequest struct {
	Source  map[string]interface{} `json:"source,omitempty"`
	Version interface{}            `json:"version,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

func NewRawResourceRequest(jsonString string) (*RawResourceRequest, error) {
	request := RawResourceRequest{}
	err := json.Unmarshal([]byte(jsonString), &request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

type ResourceRequest struct {
	Type            RequestType         `json:"-"`
	Source          SmugglerSource      `json:"source,omitempty"`
	Version         interface{}         `json:"version,omitempty"`
	Params          TaskParams          `json:"params,omitempty"`
	OrigRequest     *RawResourceRequest `json:"-"`
	FilteredRequest *RawResourceRequest `json:"-"`
}

type TaskParams struct {
	SmugglerParams map[string]interface{} `json:"smuggler_params,omitempty"`
	ExtraParams    map[string]interface{} `json:"-"`
}

func NewResourceRequest(requestType RequestType, jsonString string) (*ResourceRequest, error) {
	request := ResourceRequest{
		Type: requestType,
	}

	// Parse the request
	err := json.Unmarshal([]byte(jsonString), &request)
	if err != nil {
		return nil, err
	}

	// Populate the raw original request
	request.OrigRequest, err = NewRawResourceRequest(jsonString)
	if err != nil {
		return nil, err
	}

	// Populate a filtered version of the request without the smuggler config
	request.FilteredRequest, err = NewRawResourceRequest(jsonString)
	if err != nil {
		return nil, err
	}

	filterMapFromJsonStruct(request.FilteredRequest.Source, request.Source)
	filterMapFromJsonStruct(request.FilteredRequest.Params, request.Params)

	// The filtered request source is the extra params for smuggler source
	request.Source.ExtraParams = make(map[string]interface{})
	for k, v := range request.FilteredRequest.Source {
		request.Source.ExtraParams[k] = v
	}

	// The filtered request params is the extra params for smuggler params
	request.Params.ExtraParams = make(map[string]interface{})
	for k, v := range request.FilteredRequest.Params {
		request.Params.ExtraParams[k] = v
	}

	return &request, nil
}

// Removes the keys in a map that match the json tag names (`json:"name,opts"`)
// for the given struct value
func filterMapFromJsonStruct(m map[string]interface{}, x interface{}) {
	for _, t := range utils.ListJsonTagsOfStruct(x) {
		delete(m, t)
	}
}

func (request *ResourceRequest) ToJson() ([]byte, error) {
	return json.Marshal(request)
}

type ResourceResponse struct {
	Version  interface{}    `json:"version,omitempty"`
	Versions []interface{}  `json:"versions,omitempty"`
	Metadata []MetadataPair `json:"metadata,omitempty"`
	Type     RequestType    `json:"-"`
}

func (r *ResourceResponse) IsEmpty() bool {
	return r.Version == interface{}(nil) &&
		len(r.Versions) == 0 &&
		len(r.Metadata) == 0
}
