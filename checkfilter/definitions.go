package checkfilter

import (
	"argpatchi/k8s"
)

const (
	DEFAULT_CHECK_FILTER = "DEFAULT_CHECK_FILTER"
)

type CheckFilter interface {
	Finalize(string, bool, k8s.TargetObject) (string, error)
}

func NewCheckFilter(check_filter_type string) CheckFilter {
	return (&defaultCheckFilter{})
}
