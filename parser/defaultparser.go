package parser

import (
	"argpatchi/helpers"

	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

const (
	CORRECT_CONFIG_API_VERSION = "mmilev.io/v1alpha1"
	CORRECT_CONFIG_KIND        = "Argpatchi"
	CORRECT_TYPES              = "regex"
)

type defaultParser struct {
	filepath string
}

func (cp *defaultParser) SetConfigFilePath(filepath string) *defaultParser {
	cp.filepath = filepath
	return cp
}

func (cp *defaultParser) GetConfig() (Config, error) {
	fh, err := os.OpenFile(cp.filepath, os.O_RDONLY, 0400)
	if err != nil {
		return Config{}, helpers.GenError("Unable to open Argpatchi YAML file (\"%s\"): %s", cp.filepath, err)
	}
	defer fh.Close()
	dec := yaml.NewDecoder(fh)
	var cfg Config
	err = dec.Decode(&cfg)
	if err != nil {
		return Config{}, helpers.GenError("Unable to decode Argpatchi YAML file (\"%s\"): %s", cp.filepath, err)
	}
	err = cp.checkConfigCorrectness(&cfg)
	if err != nil {
		return Config{}, helpers.GenError("Incorrect Argpatchi YAML file (\"%s\"): %s", cp.filepath, err)
	}
	return cfg, nil
}

func (cp *defaultParser) checkConfigCorrectness(cfg *Config) error {
	if cfg.ApiVersion != CORRECT_CONFIG_API_VERSION {
		return helpers.GenError("Wrong apiVersion, expected: %s", CORRECT_CONFIG_API_VERSION)
	}
	if cfg.Kind != CORRECT_CONFIG_KIND {
		return helpers.GenError("Wrong kind, expected: %s", CORRECT_CONFIG_KIND)
	}
	type_list := strings.Split(CORRECT_TYPES, ",")
	for _, single_patch := range cfg.Patches {
		found := false
		for _, type_from_list := range type_list {
			if type_from_list == single_patch.Type {
				found = true
				break
			}
		}
		if !found {
			return helpers.GenError("Wrong type: %s, allowed types are: %q", single_patch.Type, type_list)
		}
		if single_patch.Clone && single_patch.Target.Name == "" {
			return helpers.GenError("When the patch is set to Clone, then targetObj may not be empty!")
		}
	}
	return nil
}
