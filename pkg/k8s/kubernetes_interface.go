package k8s

type KubernetesInterface interface {
	CreateServiceAccount(kubeConfigPath string, namespace string) (*ServiceAccountDetails, error)
	DeleteServiceAccount(kubeConfigPath string, namespace string, serviceAccountName string) error

	CreateRole(kubeConfigPath string, namespace string) (*RoleDetails, error)
	DeleteRole(kubeConfigPath string, namespace string, roleName string) error

	CreateRoleBinding(kubeConfigPath string, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error)
	DeleteRoleBinding(kubeConfigPath string, namespace string, roleBindingName string) error

	GetServiceAccountSecrets(kubeConfigPath string, namespace string, serviceAccountName string) ([]*ServiceAccountSecret, error)

	CreateNamespaceIfNotExists(kubeConfigPath string, namespace string) (*NamespaceDetails, error)
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
