package servian

import (
	"net/url"

	"github.com/hashicorp/go-hclog"
)

type KubeConfig struct {
	baseUrl *url.URL
	//versionedAPIPath    string
	//clientContentConfig rest.ClientContentConfig
	//rateLimiter         flowcontrol.RateLimiter
	jwt    string
	CACert string
}

type KubernetesInterface interface {
	CreateServiceAccount(pluginConfig PluginConfig, namespace string, l hclog.Logger) (*ServiceAccountDetails, error)
	GetServiceAccountSecret(pluginConfig PluginConfig, sa *ServiceAccountDetails, l hclog.Logger) ([]*ServiceAccountSecret, error)
	DeleteServiceAccount(pluginConfig PluginConfig, namespace string, serviceAccountName string, l hclog.Logger) error

	CreateRoleBinding(pluginConfig PluginConfig, namespace string, serviceAccountName string, roleName string, l hclog.Logger) (*RoleBindingDetails, error)
	DeleteRoleBinding(pluginConfig PluginConfig, namespace string, roleBindingName string, l hclog.Logger) error
}

type ServiceAccountDetails struct {
	Namespace string
	UID       string
	Name      string
}

type RoleBindingDetails struct {
	Namespace string
	UID       string
	Name      string
}

type ServiceAccountSecret struct {
	CACert    string
	Namespace string
	Token     string
}
