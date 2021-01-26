package k8s

type KubernetesInterface interface {
	CreateServiceAccount(kubeConfigPath string, namespace string) (*ServiceAccountDetails, error)
	DeleteServiceAccount(kubeConfigPath string, namespace string, serviceAccountName string) error
}

type ServiceAccountDetails struct {
	Namespace string
	UID       string
	Name      string
}
