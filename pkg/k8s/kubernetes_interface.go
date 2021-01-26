package k8s

type KubernetesInterface interface {
	CreateServiceAccount(kubeConfigPath string, namespace string) (*ServiceAccountDetails, error)
	DeleteServiceAccount(kubeConfigPath string, namespace string, serviceAccountName string) error

	CreateRole(kubeConfigPath string, namespace string) (*RoleDetails, error)
	DeleteRole(kubeConfigPath string, namespace string, roleName string) error

	CreateRoleBinding(kubeConfigPath string, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error)
	DeleteRoleBinding(kubeConfigPath string, namespace string, roleBindingName string) error
}

type ServiceAccountDetails struct {
	Namespace string
	UID       string
	Name      string
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
