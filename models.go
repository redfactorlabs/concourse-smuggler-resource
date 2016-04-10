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

type StringParams map[string]string

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

func NewCheckRequestFromJson(s string) (CheckRequest, error) {
	checkRequest := CheckRequest{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(&checkRequest)
	return checkRequest, err
}

type CheckResponse []Version
