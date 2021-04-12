package patcher

import (
	"argpatchi/parser"
)

const (
	REGEXP_PATCHER_TYPE = "REGEXP_PATCHER_TYPE"
)

type Patcher interface {
	ExecutePatch(parser.PatchDefinition, string) (string, error)
}

func NewPatcher(patcher_type string) Patcher {
	return (&regexpPatcher{})
}
