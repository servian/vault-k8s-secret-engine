package servian

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const serviceAccountNamePrefix = "vault-sa-"
const roleNamePrefix = "vault-r-"
const roleBindingNamePrefix = "vault-rb-"

const serviceAccountKind = "ServiceAccount"
const roleKind = "Role"

// KubernetesService is an empty struct to wrap the Kubernetes service functions
type KubernetesService struct{}

// CheckConnection checks connectivity with a cluster
func (k *KubernetesService) CheckConnection(pluginConfig *PluginConfig) error {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return err
	}

	_, err = clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateServiceAccount creates a new service account
func (k *KubernetesService) CreateServiceAccount(pluginConfig *PluginConfig, namespace string) (*ServiceAccountDetails, error) {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return nil, err
	}
	sa := v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: serviceAccountNamePrefix,
			Namespace:    namespace,
		},
	}
	sar, err := clientSet.CoreV1().ServiceAccounts(namespace).Create(&sa)
	if err != nil {
		return nil, err
	}

	var secrets []*string
	for _, item := range sar.Secrets {
		s := item.String()
		secrets = append(secrets, &s)
	}

	return &ServiceAccountDetails{
		Namespace: sar.Namespace,
		UID:       fmt.Sprintf("%s", sar.UID),
		Name:      sar.Name,
	}, nil
}

// GetServiceAccountSecret retrieves the secrets for a newly created service account
func (k *KubernetesService) GetServiceAccountSecret(pluginConfig *PluginConfig, sa *ServiceAccountDetails) ([]*ServiceAccountSecret, error) {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return nil, err
	}

	ksa, err := clientSet.CoreV1().ServiceAccounts(sa.Namespace).Get(sa.Name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	var secrets []*ServiceAccountSecret
	for _, secret := range ksa.Secrets {
		secretName := secret.Name
		token, err := clientSet.CoreV1().Secrets(sa.Namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if token != nil {
			caCert := string(token.Data["ca.crt"])
			tokenNamespace := string(token.Data["namespace"])
			tokenValue := string(token.Data["token"])
			secretValue := ServiceAccountSecret{
				CACert:    caCert,
				Namespace: tokenNamespace,
				Token:     tokenValue,
			}
			secrets = append(secrets, &secretValue)
		}
	}

	return secrets, nil
}

// DeleteServiceAccount removes a services account from the Kubernetes server
func (k *KubernetesService) DeleteServiceAccount(pluginConfig *PluginConfig, namespace string, serviceAccountName string) error {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return err
	}
	err = clientSet.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountName, nil)
	if err != nil {
		return err
	}
	return nil
}

// CreateRoleBinding creates a new rolebinding for a service account in a specific namespace
func (k *KubernetesService) CreateRoleBinding(pluginConfig *PluginConfig, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error) {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return nil, err
	}
	subjects := []rbac.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}

	roleBinding := rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: roleBindingNamePrefix,
			Namespace:    namespace,
		},
		Subjects: subjects,
		RoleRef: rbac.RoleRef{
			Kind: "ClusterRole",
			Name: roleName,
		},
	}

	rb, err := clientSet.RbacV1().RoleBindings(namespace).Create(&roleBinding)
	if err != nil {
		return nil, err
	}
	return &RoleBindingDetails{
		Namespace: rb.Namespace,
		UID:       fmt.Sprintf("%s", rb.UID),
		Name:      rb.Name,
	}, nil
}

// DeleteRoleBinding removes an existing role binding
func (k *KubernetesService) DeleteRoleBinding(pluginConfig *PluginConfig, namespace string, roleBindingName string) error {
	clientSet, err := getClientSet(pluginConfig)
	if err != nil {
		return err
	}
	err = clientSet.RbacV1().RoleBindings(namespace).Delete(roleBindingName, nil)
	if err != nil {
		return err
	}
	return nil
}

// getClientSet sets up a new client for accessing the kubernetes API using a bearer token and a CACert
func getClientSet(pluginConfig *PluginConfig) (*kubernetes.Clientset, error) {

	tlsConfig := rest.TLSClientConfig{
		CAData: []byte(pluginConfig.CACert),
	}

	conf := &rest.Config{
		Host:            pluginConfig.Host,
		TLSClientConfig: tlsConfig,
		BearerToken:     pluginConfig.ServiceAccountJWT,
	}

	return kubernetes.NewForConfig(conf)
}
