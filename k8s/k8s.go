package k8s

import (
	"argpatchi/helpers"
	
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

const (
	PATH_TO_SA_TOKEN = "/var/run/secrets/kubernetes.io/serviceaccount"
	K8S_URL = "https://kubernetes.default.svc"
)

type k8sConnector struct {
	url string
	cp *x509.CertPool
	token string
	api_resources map[string]string
}

func (kc *k8sConnector) EstablishConnection() (*k8sConnector, error) {
	token_path := path.Join(PATH_TO_SA_TOKEN, "token")
	token_fh, err := os.OpenFile(token_path, os.O_RDONLY, 0400)
	if err != nil {
		return nil, helpers.GenError("Unable to open token file for read: %s", err)
	}
	defer token_fh.Close()
	data, err := ioutil.ReadAll(token_fh)
	if err != nil {
		return nil, helpers.GenError("Unable to read token file: %s", err)
	}
	kc.token = strings.TrimSpace(string(data))
	
	ca_path := path.Join(PATH_TO_SA_TOKEN, "ca.crt")
	ca_fh, err := os.OpenFile(ca_path, os.O_RDONLY, 0400)
	if err != nil {
		return nil, helpers.GenError("Unable to open ca.crt file for read: %s", err)
	}
	defer ca_fh.Close()
	data, err = ioutil.ReadAll(ca_fh)
	if err != nil {
		return nil, helpers.GenError("Unable to read ca.crt file: %s", err)
	}
	kc.cp = x509.NewCertPool()
	if !kc.cp.AppendCertsFromPEM(data) {
		return nil, helpers.GenError("Unable to add ca.crt to certificate pool")
	}
	
	kc.url = K8S_URL
	
	err = kc.getApiResources()
	if err != nil {
		return nil, err
	}
	
	return kc, nil
}

func (kc *k8sConnector) GetSourceObject(source_obj SourceObjectRequest) (string, error) {
	combined_apiversion_kind := fmt.Sprintf("%s/%s", source_obj.ApiVersion, source_obj.Kind)
	plural_name, plural_name_found := kc.api_resources[combined_apiversion_kind]
	if !plural_name_found {
		return "", helpers.GenError("The API resource %s seems not to exist on this cluster", combined_apiversion_kind)
	}
	namespace_str := ""
	if source_obj.Namespace != "" {
		namespace_str = fmt.Sprintf("/namespaces/%s", source_obj.Namespace)
	}
	api_str := "api"
    if strings.Contains(source_obj.ApiVersion, "/") {
        api_str = "apis"
    }
	obj_url := fmt.Sprintf("%s/%s/%s%s/%s/%s", kc.url, api_str, source_obj.ApiVersion, namespace_str, plural_name, source_obj.Name)
	obj_result, err := kc.performGetRequest(obj_url)
	if err != nil {
		return "", helpers.GenError("%s: %s", obj_url, err)
	}
	var result_unmarshaled interface{}
	err = json.Unmarshal([]byte(obj_result), &result_unmarshaled)
	if err != nil {
		return "", helpers.GenError("Unable to parse object as interface: %s", err)
	}
	var yaml_res strings.Builder
	yaml_enc := yaml.NewEncoder(&yaml_res)
	yaml_enc.SetIndent(2)
	err = yaml_enc.Encode(&result_unmarshaled)
	yaml_enc.Close()
	if err != nil {
		return "", helpers.GenError("Unable to parse object as YAML: %s", err)
	}
	
	return yaml_res.String(), nil
}

func (kc *k8sConnector) getApiResources() error {
	api_url := fmt.Sprintf("%s/apis", kc.url)
	apis_result, err := kc.performGetRequest(api_url)
	if err != nil {
		return helpers.GenError("%s: %s", api_url, err)
	}
	var api_group_list_unmarshaled ApiGroupList
	err = json.Unmarshal([]byte(apis_result), &api_group_list_unmarshaled)
	if err != nil {
		return helpers.GenError("%s: Unable to parse ApiGroupList: %s", api_url, err)
	}
	var combined_result []string
	var combined_result_lock sync.Mutex
	var wg sync.WaitGroup
	parallel_func := func (api_group_url string) {
		api_group_result, err := kc.performGetRequest(api_group_url)
		if err != nil {
			log.Fatal(helpers.GenError("%s: %s", api_group_url, err))
		}
		combined_result_lock.Lock()
		combined_result = append(combined_result, strings.TrimSpace(api_group_result))
		combined_result_lock.Unlock()
		wg.Done()
	}
	for _, api_group := range api_group_list_unmarshaled.Groups {
		api_group_url := fmt.Sprintf("%s/apis/%s", kc.url, api_group.PreferredVersion.GroupVersion)
		wg.Add(1)
		go parallel_func(api_group_url)
	}
	main_group_url := fmt.Sprintf("%s/api/v1", kc.url)
	wg.Add(1)
	go parallel_func(main_group_url)
	wg.Wait()
	combined_result_string := fmt.Sprintf(`{"result":[%s]}`, strings.Join(combined_result, ","))
	var api_resources_unmarshaled ApiResourceListArray
	err = json.Unmarshal([]byte(combined_result_string), &api_resources_unmarshaled)
	if err != nil {
		return helpers.GenError("Unable to parse ApiResourceListArray: %s", err)
	}
	api_resources := make(map[string]string)
	for _, api_resource := range api_resources_unmarshaled.Result {
		api_res_group_version := api_resource.GroupVersion
		if strings.HasPrefix(api_res_group_version, "apps/v1") {
			api_res_group_version = "v1"
		}
		for _, api_resource_subitem := range api_resource.Resources {
			if strings.Contains(api_resource_subitem.Name, "/") {
				continue
			}
			api_resources[fmt.Sprintf("%s/%s", api_res_group_version, api_resource_subitem.Kind)] = api_resource_subitem.Name
		}
	}
	kc.api_resources = api_resources
	return nil
}

func (kc *k8sConnector) performGetRequest(url string) (string, error) {
	tls_config := &tls.Config {
		RootCAs: kc.cp,
	}
	http_transport := &http.Transport{
		TLSClientConfig: tls_config,
	}
	client := &http.Client{Transport: http_transport}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", helpers.GenError("Unable to create GET request: %s", err)
	}
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", kc.token))
	req.Header = headers
	resp, err := client.Do(req)
	if err != nil {
		return "", helpers.GenError("Unable to execute GET request: %s", err)
	}
	body_data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", helpers.GenError("Unable to read GET request body: %s", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", helpers.GenError("Got unexpected GET request response: %s: %s", resp.Status, string(body_data))
	}
	return string(body_data), nil
}
