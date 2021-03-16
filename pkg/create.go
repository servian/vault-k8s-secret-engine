package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const secretAccessKeyType = "service_account_token"

type RoleType string

const (
	RoleTypeRole        RoleType = "Role"
	RoleTypeClusterRole RoleType = "ClusterRole"
)

func secret(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretAccessKeyType,
		Fields: map[string]*framework.FieldSchema{
			keyCACert: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "CA Cert to use with the service account",
			},
			keyNamespace: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Namespace in which the service account was created",
			},
			keyServiceAccountToken: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The Service account token associated with the newly created service account",
			},
			keyServiceAccountUID: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "UID of the newly created secret",
			},
			keyServiceAccountName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the newly created service account",
			},
			keyRoleName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the newly created role",
			},
			keyRoleBindingName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the newly created role binding",
			},
			keyKubeConfig: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Path of the kubeconfig used to connect with the k8s cluster",
			},
		},
		Revoke: b.revokeSecret,
	}
}

// TODO: delete any resources created in this method if any step errors out
func (b *backend) createSecret(ctx context.Context, s logical.Storage, namespace string, roleName string, roleType RoleType) (*logical.Response, error) {
	se, err := s.Get(ctx, configPath)
	if err != nil {
		return nil, err
	}
	if se == nil {
		return nil, fmt.Errorf("the plugin is not configured correctly")
	}
	var pluginConfig PluginConfig
	err = se.DecodeJSON(&pluginConfig)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("creating secret with ttl: %d for role: %s in namespace: %s", pluginConfig.MaxTTL, roleName, namespace))
	return nil, nil
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("revoking a service account")

	namespace := d.Get(keyNamespace).(string)
	kubeconfig := d.Get(keyKubeConfig).(KubeConfig)

	serviceAccountName := d.Get(keyServiceAccountName).(string)
	serviceAccountUID := d.Get(keyServiceAccountUID).(string)

	roleBindingName := d.Get(keyRoleBindingName).(string)

	b.Logger().Info(fmt.Sprintf("deleting role binding with name: %s in namespace: %s", roleBindingName, namespace))
	err := b.kubernetesService.DeleteRoleBinding(kubeconfig, namespace, roleBindingName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted role binding with name: %s in namespace: %s", roleBindingName, namespace))

	b.Logger().Info(fmt.Sprintf("deleting service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, serviceAccountUID))
	err = b.kubernetesService.DeleteServiceAccount(kubeconfig, namespace, serviceAccountName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, serviceAccountUID))

	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyServiceAccountName: serviceAccountName,
	}, map[string]interface{}{})

	return resp, nil
}
