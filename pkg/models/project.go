package models

import "go/token"

type Project struct {
	Path           string                 `json:"path"`
	SourceProvider string                 `json:"source_provider"`
	TargetProvider string                 `json:"target_provider"`
	Files          map[string]*SourceFile `json:"files"`
	Dependencies   map[string]string      `json:"dependencies"`
	Flows          []*Flow                `json:"flows"`
	Models         []*Model               `json:"models"`
	Configuration  map[string]interface{} `json:"configuration"`
}

type SourceFile struct {
	Path        string   `json:"path"`
	PackageName string   `json:"package_name"`
	Imports     []string `json:"imports"`
	Flows       []*Flow  `json:"flows"`
	Models      []*Model `json:"models"`
	HasGenKit   bool     `json:"has_genkit"`
}

type Flow struct {
	Name        string         `json:"name"`
	Position    token.Position `json:"position"`
	InputType   string         `json:"input_type,omitempty"`
	OutputType  string         `json:"output_type,omitempty"`
	Description string         `json:"description,omitempty"`
}

type Model struct {
	Name     string         `json:"name"`
	Provider string         `json:"provider"`
	Position token.Position `json:"position"`
}
