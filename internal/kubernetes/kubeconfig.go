package kubernetes

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type KubeConfig struct {
	Clusters       []clusterEntry         `yaml:"clusters"`
	Contexts       []contextEntry         `yaml:"contexts,omitempty"`
	Users          []userEntry            `yaml:"users"`
	CurrentContext string                 `yaml:"current-context,omitempty"`
	Kind           string                 `yaml:"kind,omitempty"`
	Preferences    map[string]interface{} `yaml:"preferences,omitempty"`
}

type contextEntry struct {
	Name    string  `yaml:"name"`
	Context context `yaml:"context"`
}

type context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

type clusterEntry struct {
	Name    string  `yaml:"name"`
	Cluster cluster `yaml:"cluster"`
}

type cluster struct {
	ClusterAuthorityData string `yaml:"certificate-authority-data"`
	Server               string `yaml:"server"`
}

type userEntry struct {
	Name string `yaml:"name"`
	User user   `yaml:"user"`
}

type user struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	Token                 string `yaml:"token"`
	ClientKeyData         string `yaml:"client-key-data"`
}

func ParseKubeConfig(config string) (*KubeConfig, error) {
	var kubeConfig KubeConfig

	if err := yaml.Unmarshal([]byte(config), &kubeConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML config with error %+v", err)
	}

	return &kubeConfig, nil
}
