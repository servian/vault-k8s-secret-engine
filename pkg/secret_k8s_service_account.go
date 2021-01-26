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

const secretAccessKeyType = "service_account_token"
const keyKubeConfig = "kubeconfig"

func secretK8sServiceAccount(b *backend) *framework.Secret {
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

// TODO: delete service account, role etc. if any step errors out
func (b *backend) createSecret(ctx context.Context, s logical.Storage, kubeConfigPath string, ttl int, namespace string) (*logical.Response, error) {
	b.Logger().Info(fmt.Sprintf("creating secret with ttl: %d via kubeconfig at: %s", ttl, kubeConfigPath))

	n, err := b.kubernetesService.CreateNamespaceIfNotExists(kubeConfigPath, namespace)
	if err != nil {
		return nil, errwrap.Wrapf("the following error occurred when creating a namespace: {{err}}", err)
	}
	if n.AlreadyExisted {
		b.Logger().Info(fmt.Sprintf("namespace: %s already exists", n.Name))
	} else {
		b.Logger().Warn(fmt.Sprintf("namespace: %s was created by this vault plugin, it will *not* be deleted automatically as that could have unintended consequences", n.Name))
	}

	sa, err := b.kubernetesService.CreateServiceAccount(kubeConfigPath, n.Name)
	if err != nil {
		return nil, errwrap.Wrapf("the following error occurred when creating a service account: {{err}}", err)
	}

	secrets, err := b.kubernetesService.GetServiceAccountSecrets(kubeConfigPath, namespace, sa.Name)
	if err != nil {
		return nil, errwrap.Wrapf("the following error occurred when getting tokens for a service account: {{err}}", err)
	}
	if len(secrets) != 1 {
		return nil, fmt.Errorf("more than 1 secret found with the newly created service account, this is unexpected for the purposes of this plugin")
	}
	b.Logger().Info(fmt.Sprintf("%d tokens available", len(secrets)))

	r, err := b.kubernetesService.CreateRole(kubeConfigPath, n.Name)
	if err != nil {
		// TODO: delete service account
		return nil, errwrap.Wrapf("the following error occurred when creating a role: {{err}}", err)
	}

	rb, err := b.kubernetesService.CreateRoleBinding(kubeConfigPath, n.Name, sa.Name, r.Name)
	if err != nil {
		// TODO: delete service account and role
		return nil, errwrap.Wrapf("the following error occurred when creating a role binding: {{err}}", err)
	}

	if sa != nil {
		resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
			keyCACert:              secrets[0].CACert,
			keyNamespace:           secrets[0].Namespace,
			keyServiceAccountToken: secrets[0].Token,
			keyKubeConfig:          kubeConfigPath,
			keyServiceAccountUID:   sa.UID,
			keyServiceAccountName:  sa.Name,
			keyRoleName:            r.Name,
			keyRoleBindingName:     rb.Name,
		}, map[string]interface{}{})

		dur, err := time.ParseDuration(fmt.Sprintf("%ds", ttl))
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("ttl: %d could not be parsed due to error: {{err}}", ttl), err)
		}
		resp.Secret.TTL = dur
		resp.Secret.MaxTTL = dur
		resp.Secret.Renewable = false
		b.Logger().Info(fmt.Sprintf("namespace uid: %s service account uid: %s, role uid: %s, role binding uid: %s", n.UID, sa.UID, r.UID, rb.UID))
		b.Logger().Info(fmt.Sprintf("created secret with ttl: %d via kubeconfig at: %s", ttl, kubeConfigPath))
		return resp, nil
	} else {
		return nil, fmt.Errorf("could not return the uid of the newly created service account")
	}
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("revoking a service account")

	namespace := d.Get(keyNamespace).(string)
	kubeconfig := d.Get(keyKubeConfig).(string)

	serviceAccountName := d.Get(keyServiceAccountName).(string)
	serviceAccountUID := d.Get(keyServiceAccountUID).(string)

	roleName := d.Get(keyRoleName).(string)
	roleBindingName := d.Get(keyRoleBindingName).(string)

	b.Logger().Info(fmt.Sprintf("deleting role binding with name: %s in namespace: %s", roleBindingName, namespace))
	err := b.kubernetesService.DeleteRoleBinding(kubeconfig, namespace, roleBindingName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted role binding with name: %s in namespace: %s", roleBindingName, namespace))

	b.Logger().Info(fmt.Sprintf("deleting role with name: %s in namespace: %s", roleName, namespace))
	err = b.kubernetesService.DeleteRole(kubeconfig, namespace, roleName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted role with name: %s in namespace: %s", roleName, namespace))

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
