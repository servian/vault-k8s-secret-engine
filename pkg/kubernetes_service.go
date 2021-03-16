package servian

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const serviceAccountNamePrefix = "vault-sa-"
const roleNamePrefix = "vault-r-"
const roleBindingNamePrefix = "vault-rb-"

const serviceAccountKind = "ServiceAccount"
const roleKind = "Role"

type KubernetesService struct {
}

func (k *KubernetesService) CreateServiceAccount(kubeConfig KubeConfig, namespace string) (*ServiceAccountDetails, error) {
	clientSet, err := getClientSet(kubeConfig)
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

func (k *KubernetesService) DeleteServiceAccount(kubeConfig KubeConfig, namespace string, serviceAccountName string) error {
	clientSet, err := getClientSet(kubeConfig)
	if err != nil {
		return err
	}
	err = clientSet.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesService) CreateRoleBinding(kubeConfig KubeConfig, namespace string, serviceAccountName string, roleName string) (*RoleBindingDetails, error) {
	clientSet, err := getClientSet(kubeConfig)
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

func (k *KubernetesService) DeleteRoleBinding(kubeConfig KubeConfig, namespace string, roleBindingName string) error {
	clientSet, err := getClientSet(kubeConfig)
	if err != nil {
		return err
	}
	err = clientSet.RbacV1().RoleBindings(namespace).Delete(roleBindingName, nil)
	if err != nil {
		return err
	}
	return nil
}

func getClientSet(kubeConfig KubeConfig) (*kubernetes.Clientset, error) {
	return nil, nil
}
