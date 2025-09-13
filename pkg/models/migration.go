package models

type Migration struct {
	Project     *Project          `json:"project"`
	Changes     []*Change         `json:"changes"`
	NewFiles    map[string]string `json:"new_files"`
	DeleteFiles []string          `json:"delete_files"`
	Commands    []string          `json:"commands"`
}

type Change struct {
	Type        string `json:"type"` // "dependency", "import", "model", "config"
	Description string `json:"description"`
	File        string `json:"file"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
}
