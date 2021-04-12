package parser

import (
	"argpatchi/k8s"
)

const (
	DEFAULT_CONFIG_FILE_PARSER = "DEFAULT_CONFIG_FILE_PARSER"
)

type Config struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Patches    []struct {
		Source k8s.SourceObjectRequest `yaml:"sourceObj"`
		Patch  PatchDefinition         `yaml:"patch"`
		Type   string                  `yaml:"type"`
		Clone  bool                    `yaml:"clone"`
		Target k8s.TargetObject        `yaml:"targetObj,omitempty"`
	} `yaml:"patches"`
}

type PatchDefinition struct {
	SearchFor   string `yaml:"searchFor,omitempty"`
	Replacement string `yaml:"replacement"`
}

type ConfigParserInterface interface {
	GetConfig() (Config, error)
}

func NewConfigParser(config_parser_type string, config_file_path string) ConfigParserInterface {
	return (&defaultParser{filepath: ""}).SetConfigFilePath(config_file_path)
}
