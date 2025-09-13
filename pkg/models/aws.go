package models

type AWSConfig struct {
	Region        string            `json:"region"`
	Profile       string            `json:"profile"`
	BedrockModels []string          `json:"bedrock_models"`
	CloudWatch    *CloudWatchConfig `json:"cloudwatch"`
	Environment   map[string]string `json:"environment"`
}

type CloudWatchConfig struct {
	Namespace string `json:"namespace"`
	Enabled   bool   `json:"enabled"`
}

type TerraformConfig struct {
	Provider    string            `json:"provider"`
	Region      string            `json:"region"`
	Variables   map[string]string `json:"variables"`
	Resources   []string          `json:"resources"`
	ProjectName string            `json:"project_name"`
}
