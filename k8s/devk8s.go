package k8s

import (
	"argpatchi/helpers"

	"io/ioutil"
	"os"
)

const ()

type devK8sConnector struct {
}

func (dkc *devK8sConnector) GetSourceObject(source_obj SourceObjectRequest) (string, error) {
	fh, err := os.OpenFile("test-source-obj.yaml", os.O_RDONLY, 0400)
	if err != nil {
		return "", helpers.GenError("Unable to open test source object file: %s", err)
	}
	defer fh.Close()
	data, err := ioutil.ReadAll(fh)
	if err != nil {
		return "", helpers.GenError("Unable to read test source object file: %s", err)
	}
	return string(data), nil
}
