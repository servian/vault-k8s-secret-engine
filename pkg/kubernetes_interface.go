package servian

// Contents of the kubeconfig file used to communicate with the cluster
// Refer to: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/ for more details about the kubeconfig file
type KubeConfig string

type KubernetesInterface interface {
	CreateServiceAccount(kubeConfig KubeConfig, namespace string) (*ServiceAccountDetails, error)
	DeleteServiceAccount(kubeConfig KubeConfig, namespace string, serviceAccountName string) error

	CreateRole(kubeConfig KubeConfig, namespace string) (*RoleDetails, error)
	DeleteRole(kubeConfig KubeConfig, namespace string, roleName string) error

	CreateRoleBinding(kubeConfig KubeConfig, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error)
	DeleteRoleBinding(kubeConfig KubeConfig, namespace string, roleBindingName string) error

	GetServiceAccountSecrets(kubeConfig KubeConfig, namespace string, serviceAccountName string) ([]*ServiceAccountSecret, error)

	CreateNamespaceIfNotExists(kubeConfig KubeConfig, namespace string) (*NamespaceDetails, error)
}

type ServiceAccountDetails struct {
	Namespace string
	UID       string
	Name      string
}

type ServiceAccountSecret struct {
	CACert    string
	Namespace string
	Token     string
}

type RoleDetails struct {
	Namespace string
	UID       string
	Name      string
}

type RoleBindingDetails struct {
	Namespace string
	UID       string
	Name      string
}

type NamespaceDetails struct {
	Namespace      string
	UID            string
	Name           string
	AlreadyExisted bool
}
