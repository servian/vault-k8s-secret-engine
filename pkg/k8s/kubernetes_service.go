package k8s

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const serviceAccountNamePrefix = "vault-sa-"
const roleNamePrefix = "vault-r-"
const roleBindingNamePrefix = "vault-rb-"

const serviceAccountKind = "ServiceAccount"
const roleKind = "Role"

type KubernetesService struct {
}

func (k *KubernetesService) CreateServiceAccount(kubeConfigPath string, namespace string) (*ServiceAccountDetails, error) {
	clientSet, err := getClientSet(kubeConfigPath)
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
	return &ServiceAccountDetails{
		Namespace: sar.Namespace,
		UID:       fmt.Sprintf("%s", sar.UID),
		Name:      sar.Name,
	}, nil
}

func (k *KubernetesService) DeleteServiceAccount(kubeConfigPath string, namespace string, serviceAccountName string) error {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return err
	}
	err = clientSet.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesService) CreateRole(kubeConfigPath string, namespace string) (*RoleDetails, error) {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	role := rbac.Role{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: roleNamePrefix,
			Namespace:    namespace,
		},
		Rules: nil,
	}
	r, err := clientSet.RbacV1().Roles(namespace).Create(&role)
	if err != nil {
		return nil, err
	}
	return &RoleDetails{
		Namespace: r.Namespace,
		UID:       fmt.Sprintf("%s", r.UID),
		Name:      r.Name,
	}, nil
}

func (k *KubernetesService) CreateRoleBinding(kubeConfigPath string, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error) {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	subjects := []rbac.Subject{
		{
			Kind:      serviceAccountKind,
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
			Kind: roleKind,
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

func (k *KubernetesService) DeleteRole(kubeConfigPath string, namespace string, roleName string) error {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return err
	}
	err = clientSet.RbacV1().Roles(namespace).Delete(roleName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesService) DeleteRoleBinding(kubeConfigPath string, namespace string, roleBindingName string) error {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return err
	}
	err = clientSet.RbacV1().RoleBindings(namespace).Delete(roleBindingName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesService) CreateNamespaceIfNotExists(kubeConfigPath string, namespace string) (*NamespaceDetails, error) {
	clientSet, err := getClientSet(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// TODO: Implement List Options to do paging
	namespaces, err := clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// If the namespace already exists, return its details without creating it
	for _, item := range namespaces.Items {
		if item.Name == namespace {
			return &NamespaceDetails{
				Namespace:      item.Namespace,
				UID:            fmt.Sprintf("%s", item.UID),
				Name:           item.Name,
				AlreadyExisted: true,
			}, nil
		}
	}

	namespaceResource := v1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec:   v1.NamespaceSpec{},
		Status: v1.NamespaceStatus{},
	}
	n, err := clientSet.CoreV1().Namespaces().Create(&namespaceResource)
	if err != nil {
		return nil, err
	}
	return &NamespaceDetails{
		Namespace:      n.Namespace,
		UID:            fmt.Sprintf("%s", n.UID),
		Name:           n.Name,
		AlreadyExisted: false,
	}, err
}

func getClientSet(kubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}
