package servian

// Contents of the kubeconfig file used to communicate with the cluster
// Refer to: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/ for more details about the kubeconfig file
type KubeConfig string

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
