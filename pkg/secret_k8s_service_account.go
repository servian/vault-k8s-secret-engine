package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const secretAccessKeyType = "access_keys"
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
	sa, err := b.kubernetesService.CreateServiceAccount(kubeConfigPath, namespace)
	if err != nil {
		return nil, errwrap.Wrapf("the following error occurred when querying service accounts: {{err}}", err)
	}
	if sa != nil {
		b.Logger().Info(fmt.Sprintf("created service account with name: %s in namespace: %s with uid: %s", sa.Name, sa.Namespace, sa.UID))
		resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
			"ca.crt":              roleName,
			keyNamespace:          sa.Namespace,
			"token":               "some token",
			keyUID:                sa.UID,
			keyServiceAccountName: sa.Name,
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

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("revoking a service account")
	serviceAccountName := d.Get(keyServiceAccountName).(string)
	uid := d.Get(keyUID).(string)
	namespace := d.Get(keyNamespace).(string)
	kubeconfig := d.Get(keyKubeConfig).(string)
	b.Logger().Info(fmt.Sprintf("deleting service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, uid))
	err := b.kubernetesService.DeleteServiceAccount(kubeconfig, namespace, serviceAccountName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, uid))
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyServiceAccountName: serviceAccountName,
	}, map[string]interface{}{})
	return resp, nil
}
