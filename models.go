package smuggler

import (
	"encoding/json"
	"strings"
)

type Source struct {
	SmugglerConfig SmugglerConfig    `json:"smuggler_config,omitempty"`
	ExtraParams    map[string]string `json:"extra_params,omitempty"`
}

type SmugglerConfig struct {
	CheckCommand CommandDefinition `json:"check,omitempty"`
	InCommand    CommandDefinition `json:"in,omitempty"`
	OutCommand   CommandDefinition `json:"out,omitempty"`
}

func (source Source) IsValid() (bool, string) {
	return true, ""
}

type CommandDefinition struct {
	Path string   `json:"path"`
	Args []string `json:"args,omitempty"`
}

func (commandDefinition CommandDefinition) IsDefined() bool {
	return (commandDefinition.Path != "")
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
