package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const secretAccessKeyType = "access_keys"
const secretNamePrefix = "vault-dsa-"
const keyKubeConfig = "kubeconfig"

func secretK8sServiceAccount(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretAccessKeyType,
		Fields: map[string]*framework.FieldSchema{
			"ca.crt": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "CA Cert",
			},
			keyNamespace: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Namespace",
			},
			"token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Token",
			},
			keyUID: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "UID of the newly created secret",
			},
			keyServiceAccountName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the newly created service account",
			},
			keyKubeConfig: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Path of the kubeconfig used to connect with the k8s cluster",
			},
		},
		Revoke: b.revokeSecret,
	}
}

func (b *backend) secretAccessKeysCreate(ctx context.Context, s logical.Storage, roleName string, kubeConfigPath string, ttl int, namespace string) (*logical.Response, error) {
	b.Logger().Info(fmt.Sprintf("creating secret for role: %s with ttl: %d via kubeconfig at: %s", roleName, ttl, kubeConfigPath))

	sar, err := createServiceAccount(kubeConfigPath, namespace)
	if err != nil {
		return nil, errwrap.Wrapf("the following error occurred when querying service accounts: {{err}}", err)
	}
	if sar != nil {
		b.Logger().Info(fmt.Sprintf("created service account with name: %s in namespace: %s with uid: %s", sar.Name, sar.Namespace, sar.UID))
		resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
			"ca.crt":              roleName,
			keyNamespace:          sar.Namespace,
			"token":               "some token",
			keyUID:                sar.UID,
			keyServiceAccountName: sar.Name,
			keyKubeConfig:         kubeConfigPath,
		}, map[string]interface{}{})

		dur, err := time.ParseDuration(fmt.Sprintf("%ds", ttl))
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("ttl: %d could not be parsed due to error: {{err}}", ttl), err)
		}
		resp.Secret.TTL = dur
		resp.Secret.MaxTTL = dur
		resp.Secret.Renewable = false

		return resp, nil
	} else {
		return nil, fmt.Errorf("could not return the uid of the newly created service account")
	}
}

func createServiceAccount(kubeconfig string, namespace string) (*v1.ServiceAccount, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	sa := v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: secretNamePrefix,
		},
	}
	sar, err := clientset.CoreV1().ServiceAccounts(namespace).Create(&sa)
	if err != nil {
		return nil, err
	}
	return sar, nil
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("revoking secret")
	serviceAccountName := d.Get(keyServiceAccountName).(string)
	b.Logger().Info(fmt.Sprintf("serviceAccountName: %s", serviceAccountName))
	uid := d.Get(keyUID).(string)
	b.Logger().Info(fmt.Sprintf("uid: %s", uid))
	namespace := d.Get(keyNamespace).(string)
	b.Logger().Info(fmt.Sprintf("namespace: %s", namespace))
	kubeconfig := d.Get(keyKubeConfig).(string)
	b.Logger().Info(fmt.Sprintf("kubeconfig: %s", kubeconfig))

	b.Logger().Warn(fmt.Sprintf("deleting service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, uid))
	err := deleteServiceAccount(b, kubeconfig, namespace, serviceAccountName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, uid))
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyServiceAccountName: serviceAccountName,
	}, map[string]interface{}{})
	return resp, nil
}

func deleteServiceAccount(b *backend, kubeconfig string, namespace string, name string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return errwrap.Wrapf("error building config from kubeconfig: {{err}}", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errwrap.Wrapf("error building clientset from kubeconfig: {{err}}", err)
	}
	err = clientset.CoreV1().ServiceAccounts(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return errwrap.Wrapf("error while deleting a service account: {{err}}", err)
	}
	return nil
}
