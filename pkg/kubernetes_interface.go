package servian

import (
	"net/url"
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
	CreateServiceAccount(pluginConfig *PluginConfig, namespace string) (*ServiceAccountDetails, error)
	GetServiceAccountSecret(pluginConfig *PluginConfig, sa *ServiceAccountDetails) ([]*ServiceAccountSecret, error)
	DeleteServiceAccount(pluginConfig *PluginConfig, namespace string, serviceAccountName string) error

	CreateRoleBinding(pluginConfig *PluginConfig, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error)
	DeleteRoleBinding(pluginConfig *PluginConfig, namespace string, roleBindingName string) error
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
