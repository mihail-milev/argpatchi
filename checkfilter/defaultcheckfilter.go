package checkfilter

import (
	"argpatchi/helpers"
	"argpatchi/k8s"

	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	yaml "gopkg.in/yaml.v3"
)

const (
	JSON_PATCH = `[
  { "op": "remove", "path": "/metadata/creationTimestamp" },
  { "op": "remove", "path": "/metadata/managedFields" }
]`
    JSON_PATCH_SELF_LINK_ADDITION = `[
  { "op": "remove", "path": "/metadata/selfLink" }
]`
	JSON_PATCH_CLONING_BASE = `[
  { "op": "remove", "path": "/metadata/resourceVersion" },
  { "op": "remove", "path": "/metadata/uid" },
  { "op": "replace", "path": "/metadata/name", "value": "%s" }%s
]`
	JSON_PATCH_CLONING_NAMESPACE = `,
  { "op": "replace", "path": "/metadata/namespace", "value": "%s" }`
)

type defaultCheckFilter struct {
}

func (dcf *defaultCheckFilter) Finalize(patched_data string, cloning bool, target k8s.TargetObject) (string, error) {
	var unmarshaled interface{}
	err := yaml.Unmarshal([]byte(patched_data), &unmarshaled)
	if err != nil {
		return "", helpers.GenError("Patched data is not valid YAML: %s", err)
	}
	json_marshaled, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", helpers.GenError("Unable to convert patched YAML to JSON: %s", err)
	}
	patcher, err := jsonpatch.DecodePatch([]byte(JSON_PATCH))
	if err != nil {
		return "", helpers.GenError("Unable to generate JSON base patch: %s", err)
	}
	json_patched_data, err := patcher.Apply(json_marshaled)
	if err != nil {
		return "", helpers.GenError("Unable to apply JSON patch: %s", err)
	}
	
	self_link_patcher, err := jsonpatch.DecodePatch([]byte(JSON_PATCH_SELF_LINK_ADDITION))
	if err != nil {
		return "", helpers.GenError("Unable to generate JSON selfLink patch: %s", err)
	}
	json_patched_data_new, err := self_link_patcher.Apply(json_patched_data)
    if err == nil {
        json_patched_data = json_patched_data_new
    }
	
	if cloning {
		json_patched_data, err = dcf.performCloningPatch(json_patched_data, target)
		if err != nil {
			return "", helpers.GenError("Unable to execute cloning patch: %s", err)
		}
	}
	var filtered_structure interface{}
	err = json.Unmarshal(json_patched_data, &filtered_structure)
	if err != nil {
		return "", helpers.GenError("Unable to convert patched JSON to structure: %s", err)
	}
	final_yaml, err := yaml.Marshal(filtered_structure)
	if err != nil {
		return "", helpers.GenError("Unable to convert patched JSON structure to YAML: %s", err)
	}
	return string(final_yaml), nil
}

func (dcf *defaultCheckFilter) performCloningPatch(patched_data []byte, target k8s.TargetObject) ([]byte, error) {
	var patch_ns string
	if target.Namespace != "" {
		patch_ns = fmt.Sprintf(JSON_PATCH_CLONING_NAMESPACE, target.Namespace)
	}
	cloning_json_patch := fmt.Sprintf(JSON_PATCH_CLONING_BASE, target.Name, patch_ns)
	patcher, err := jsonpatch.DecodePatch([]byte(cloning_json_patch))
	if err != nil {
		return []byte(""), helpers.GenError("Unable to generate cloning JSON patch: %s", err)
	}
	json_patched_data, err := patcher.Apply(patched_data)
	if err != nil {
		return []byte(""), helpers.GenError("Unable to apply JSON patch: %s", err)
	}
	return json_patched_data, nil
}
