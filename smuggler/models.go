package smuggler

import (
	"encoding/json"
	"fmt"
)

type SmugglerSource struct {
	Commands       []CommandDefinition    `json:"commands,omitempty"`
	SmugglerParams map[string]interface{} `json:"smuggler_params,omitempty"`
	ExtraParams    map[string]interface{} `json:"-"`
}

func (source SmugglerSource) FindCommand(name string) *CommandDefinition {
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

type ResourceRequest struct {
	Type            RequestType            `json:"-"`
	Source          SmugglerSource         `json:"source,omitempty"`
	Version         interface{}            `json:"version,omitempty"`
	Params          map[string]interface{} `json:"params,omitempty"`
	OrigRequest     RawResourceRequest     `json:"-"`
	FilteredRequest RawResourceRequest     `json:"-"`
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
	// Flatten the task params
	m := request.Params["smuggler_params"]
	if m != nil {
		switch m := m.(type) {
		case map[string]interface{}:
			for k, v := range m {
				request.Params[k] = v
			}
		default:
			return nil, fmt.Errorf("Invalid format in task params 'smuggler_params', expected hash of key-value")
		}
		delete(request.Params, "smuggler_params")
	}

	// Populate the raw original request
	rawRequest := RawResourceRequest{}
	err = json.Unmarshal([]byte(jsonString), &rawRequest)
	if err != nil {
		return nil, err
	}
	request.OrigRequest = rawRequest

	// Populate a filtered version of the request without the smuggler config
	filteredRequest := RawResourceRequest{}
	err = json.Unmarshal([]byte(jsonString), &filteredRequest)
	delete(filteredRequest.Source, "commands")
	delete(filteredRequest.Source, "smuggler_params")
	delete(filteredRequest.Params, "smuggler_params")

	request.FilteredRequest = filteredRequest

	// The fistered request source is the extra params for smuggler
	request.Source.ExtraParams = request.FilteredRequest.Source

	return &request, nil
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
