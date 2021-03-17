package servian

import (
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type KubeConfig struct {
	baseUrl             *url.URL
	versionedAPIPath    string
	clientContentConfig rest.ClientContentConfig
	rateLimiter         flowcontrol.RateLimiter
	jwt                 string
	apiServerCert       string
}

type KubernetesInterface interface {
	CreateServiceAccount(kubeConfig KubeConfig, namespace string) (*ServiceAccountDetails, error)
	DeleteServiceAccount(kubeConfig KubeConfig, namespace string, serviceAccountName string) error

	CreateRoleBinding(kubeConfig KubeConfig, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error)
	DeleteRoleBinding(kubeConfig KubeConfig, namespace string, roleBindingName string) error
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
