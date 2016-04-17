package smuggler

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
)

type SmugglerSource struct {
	Commands         map[string]interface{} `json:"commands,omitempty"`
	FilterRawRequest bool                   `json:"filter_raw_request,omitempty"`
	SmugglerParams   map[string]interface{} `json:"smuggler_params,omitempty"`
	ExtraParams      map[string]interface{} `json:"-"`
}

func WrapCommandWithShell(name string, commandLine string) *CommandDefinition {
	// Try to find bash
	shellPath, err := exec.LookPath("bash")
	if err == nil {
		return &CommandDefinition{
			Path: shellPath,
			Args: []string{"-e", "-u", "-o", "pipefail", "-c", commandLine},
		}
	}
	// Try to find sh
	shellPath, err = exec.LookPath("sh")
	if err == nil {
		return &CommandDefinition{
			Path: shellPath,
			Args: []string{"-e", "-u", "-o", "pipefail", "-c", commandLine},
		}
	}

	// In last case, use the command itself
	l := strings.Split(commandLine, ",")
	return &CommandDefinition{
		Path: l[0],
		Args: l[1:],
	}
}

func (source SmugglerSource) FindCommand(name string) (*CommandDefinition, error) {
	cmd, ok := source.Commands[name]
	if !ok {
		return nil, nil
	}
	switch cmd := cmd.(type) {
	case string:
		return WrapCommandWithShell(name, cmd), nil
	default:
		c, err := NewCommandDefinition(cmd)
		return c, err
	}
}

type CommandDefinition struct {
	Path string   `json:"path"`
	Args []string `json:"args,omitempty"`
}

func NewCommandDefinition(i interface{}) (*CommandDefinition, error) {
	// Small hack to remap a interface{} to a struct
	// Convert interface{} => json => struct
	var b []byte
	var c CommandDefinition
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (commandDefinition CommandDefinition) IsDefined() bool {
	return (commandDefinition.Path != "")
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
