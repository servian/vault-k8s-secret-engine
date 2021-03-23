package servian

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
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
		},
		Revoke: b.revokeSecret,
	}
}

func (b *backend) createSecret(ctx context.Context, s logical.Storage, namespace string, roleName string, roleType RoleType, ttl int) (*logical.Response, error) {
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
	sa, err := b.kubernetesService.CreateServiceAccount(pluginConfig, namespace, b.Logger())

	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error creating Kubernetes service account: %s", err))
		return nil, err
	}

	// give the kubernetes cluster a chance to generate the secret for the SA
	time.Sleep(1 * time.Second)

	secrets, err := b.kubernetesService.GetServiceAccountSecret(pluginConfig, sa, b.Logger())
	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error loading secrets for service account: %s", err))
		b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name, b.Logger())
		return nil, err
	}

	if len(secrets) != 1 {
		//b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name, b.Logger())
		b.Logger().Error(fmt.Sprint("Found %d Secrets for service account, expected 1", len(secrets)))
		for secret := range secrets {

			b.Logger().Info(fmt.Sprint("Secret found: %s", secret))
		}
		return nil, fmt.Errorf("More than 1 secret found with the newly created service account, this is unexpected for the prupose of this plugin, please try again.")
	}

	rb, err := b.kubernetesService.CreateRoleBinding(pluginConfig, namespace, sa.Name, roleName, b.Logger())

	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error setting up Kubernetes role binding for SA %s: %s", sa.Name, err))
		//b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name, b.Logger())
		return nil, err
	}

	b.Logger().Info(fmt.Sprintf("Service account '%s' created with rolebinding '%s'", sa.Name, rb.Name))

	dur, err := time.ParseDuration(fmt.Sprintf("%ds", ttl))
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("ttl: %d could not be parse due to error: {{err}}", ttl, err), err)
	}

	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyCACert:              secrets[0].CACert,
		keyNamespace:           secrets[0].Namespace,
		keyServiceAccountToken: secrets[0].Token,
		keyServiceAccountName:  sa.Name,
		keyServiceAccountUID:   sa.UID,
		keyRoleName:            roleName,
		keyRoleBindingName:     rb.Name,
		keyTtlSeconds:          dur,
		keyMaxTTL:              dur,
	}, map[string]interface{}{})

	resp.Secret.Renewable = false

	return resp, nil
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	se, err := req.Storage.Get(ctx, configPath)
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

	b.Logger().Info("revoking a service account")

	namespace := d.Get(keyNamespace).(string)

	serviceAccountName := d.Get(keyServiceAccountName).(string)
	serviceAccountUID := d.Get(keyServiceAccountUID).(string)

	roleBindingName := d.Get(keyRoleBindingName).(string)

	b.Logger().Info(fmt.Sprintf("deleting role binding with name: %s in namespace: %s", roleBindingName, namespace))
	err = b.kubernetesService.DeleteRoleBinding(pluginConfig, namespace, roleBindingName, b.Logger())
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted role binding with name: %s in namespace: %s", roleBindingName, namespace))

	b.Logger().Info(fmt.Sprintf("deleting service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, serviceAccountUID))
	err = b.kubernetesService.DeleteServiceAccount(pluginConfig, namespace, serviceAccountName, b.Logger())
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted service account with name: %s in namespace: %s with uid: %s", serviceAccountName, namespace, serviceAccountUID))

	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyServiceAccountName: serviceAccountName,
	}, map[string]interface{}{})

	return resp, nil
}
