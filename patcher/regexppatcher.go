package patcher

import (
	"argpatchi/helpers"
	"argpatchi/parser"

	"regexp"
)

const ()

type regexpPatcher struct {
}

func (rp *regexpPatcher) ExecutePatch(patch parser.PatchDefinition, source string) (string, error) {
	re, err := regexp.Compile(patch.SearchFor)
	if err != nil {
		return "", helpers.GenError("Unable to compile patch to regular expression: %s", err)
	}
	return re.ReplaceAllString(source, patch.Replacement), nil
}
