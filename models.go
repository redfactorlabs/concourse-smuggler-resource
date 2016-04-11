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

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

func NewCheckRequestFromJson(s string) (CheckRequest, error) {
	request := CheckRequest{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(&request)
	return request, err
}

type CheckResponse []Version

type InRequest struct {
	Source  Source            `json:"source"`
	Version Version           `json:"version"`
	Params  map[string]string `json:"params,omitempty"`
}

func NewInRequestFromJson(s string) (InRequest, error) {
	request := InRequest{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(&request)
	return request, err
}

type InResponse struct {
	Version  Version        `json:"version"`
	Metadata []MetadataPair `json:"metadata"`
}

type OutRequest struct {
	Source Source            `json:"source"`
	Params map[string]string `json:"params,omitempty"`
}

func NewOutRequestFromJson(s string) (OutRequest, error) {
	request := OutRequest{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(&request)
	return request, err
}

type OutResponse struct {
	Version  Version        `json:"version"`
	Metadata []MetadataPair `json:"metadata"`
}
