package k8s

const (
	DEV_K8S_CONNECTOR = "DEV_K8S_CONNECTOR"
	K8S_CONNECTOR     = "K8S_CONNECTOR"
)

type SourceObjectRequest struct {
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace,omitempty"`
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

type TargetObject struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

type ApiGroupList struct {
	Groups []struct {
		PreferredVersion struct {
			GroupVersion string `yaml:"groupVersion"`
		} `yaml:"preferredVersion"`
	} `yaml:"groups"`
}

type ApiResourceListArray struct {
	Result []struct {
		GroupVersion string `yaml:"groupVersion"`
		Resources    []struct {
			Name string `yaml:"name"`
			Kind string `yaml:"kind"`
		} `yaml:"resources"`
	} `yaml:"result"`
}

type K8sConnector interface {
	GetSourceObject(SourceObjectRequest) (string, error)
}

func NewK8sConnector(k8s_connector_type string, clust_url ...string) (K8sConnector, error) {
	if k8s_connector_type == DEV_K8S_CONNECTOR {
		return (&devK8sConnector{}), nil
	}
	final_clust_url := func() string {
		if len(clust_url) > 0 {
			return clust_url[0]
		}
		return ""
	}()
	return (&k8sConnector{url: final_clust_url}).EstablishConnection()
}
