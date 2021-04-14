package main

import (
	"argpatchi/checkfilter"
	"argpatchi/helpers"
	"argpatchi/k8s"
	"argpatchi/parser"
	"argpatchi/patcher"

	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	REGEXP_ARGPATCHI_YAML_FILE = `^(?i)argpatchi\.ya?ml$`
)

func findArgpatchiYamlFolderSubitem(subitems []os.FileInfo) (string, error) {
	re := regexp.MustCompile(REGEXP_ARGPATCHI_YAML_FILE)
	for _, subitem := range subitems {
		if re.MatchString(subitem.Name()) {
			return subitem.Name(), nil
		}
	}
	return "", helpers.GenError("No file matching the expression \"%s\" found", REGEXP_ARGPATCHI_YAML_FILE)
}

func getArgpatchiYamlFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", helpers.GenError("Unable to get current working directory: %s", err)
	}
	cwd_fh, err := os.OpenFile(cwd, os.O_RDONLY, 0400)
	if err != nil {
		return "", helpers.GenError("Unable to open current working directory: %s", err)
	}
	defer cwd_fh.Close()
	cwd_subitems, err := cwd_fh.Readdir(-1)
	if err != nil {
		return "", helpers.GenError("Unable to get subitems of current working directory: %s", err)
	}
	argpatchi_yaml_file_name, err := findArgpatchiYamlFolderSubitem(cwd_subitems)
	if err != nil {
		return "", helpers.GenError("Argpatchi YAML file not found: %s", err)
	}
	return path.Join(cwd, argpatchi_yaml_file_name), nil
}

func generateErrorPath(array_index int, api_version, kind, namespace, name string) string {
	return fmt.Sprintf("/[%d]/%s/%s/%s/%s: ", array_index, api_version, kind, namespace, name)
}

func main() {
	debug_mode := flag.Bool("d", false, "Set this flag to activate debug messages (useful for testing patches)")
	flag.Parse()
	if *debug_mode {
		log.SetLevel(log.DebugLevel)
	}

	argpatchi_yaml_file_path, err := getArgpatchiYamlFilePath()
	if err != nil {
		log.Fatal(err)
	}
	config_parser := parser.NewConfigParser(parser.DEFAULT_CONFIG_FILE_PARSER, argpatchi_yaml_file_path)
	config, err := config_parser.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	for i, single_patch := range config.Patches {
		err_path := generateErrorPath(i, single_patch.Source.ApiVersion, single_patch.Source.Kind, single_patch.Source.Namespace, single_patch.Source.Name)

		k8s_connector, err := k8s.NewK8sConnector(k8s.K8S_CONNECTOR, single_patch.Cluster)
		if err != nil {
			log.Fatal(err_path, err)
		}

		source_obj, err := k8s_connector.GetSourceObject(single_patch.Source)
		if err != nil {
			log.Fatal(err_path, err)
		}

		if *debug_mode {
			source_obj_lines := strings.Split(source_obj, "\n")
			log.Debugln(err_path, "Source object contents:")
			for _, line := range source_obj_lines {
				log.Debugf("|%s\n", line)
			}
			log.Debugln()
			log.Debugln()
			patch_search_contents := strings.Split(single_patch.Patch.SearchFor, "\n")
			log.Debugln(err_path, "Patch 'Search For' contents:")
			for _, line := range patch_search_contents {
				log.Debugf("|%s\n", line)
			}
			log.Debugln()
		}

		patcher_instance := patcher.NewPatcher(patcher.REGEXP_PATCHER_TYPE)
		patch_result, err := patcher_instance.ExecutePatch(single_patch.Patch, source_obj)
		if err != nil {
			log.Fatal(err_path, err)
		}
		check_fitler_instance := checkfilter.NewCheckFilter(checkfilter.DEFAULT_CHECK_FILTER)
		check_result, err := check_fitler_instance.Finalize(patch_result, single_patch.Clone, single_patch.Target)
		if err != nil {
			log.Fatal(err_path, err)
		}
		fmt.Println(check_result)
		fmt.Println("---")
	}
}
