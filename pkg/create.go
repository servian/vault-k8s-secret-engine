package servian

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const secretAccessKeyType = "service_account_token"

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
				Description: "Namespace in which the service account will be created",
			},
			keyServiceAccountToken: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The Service account token associated with the newly created service account",
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

func (b *backend) createSecret(ctx context.Context, s logical.Storage, saType string, namespace string, ttl int) (*logical.Response, error) {

	// reload plugin config on every call to prevent stale config
	pluginConfig, err := loadPluginConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	roleName, err := getClusterRoleName(pluginConfig, saType)
	if err != nil {
		return nil, err
	}

	if ttl <= 0 {
		ttl = pluginConfig.DefaulTTL
	}

	if ttl > pluginConfig.MaxTTL {
		ttl = pluginConfig.MaxTTL
	}

	b.Logger().Info(fmt.Sprintf("creating secret with ttl: %d for role: %s in namespace: %s", ttl, roleName, namespace))
	sa, err := b.kubernetesService.CreateServiceAccount(pluginConfig, namespace)

	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error creating Kubernetes service account: %s", err))
		return nil, err
	}

	// give the kubernetes cluster a chance to generate the secret for the SA
	time.Sleep(1 * time.Second)

	secrets, err := b.kubernetesService.GetServiceAccountSecret(pluginConfig, sa)
	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error loading secrets for service account: %s", err))
		b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name)
		return nil, err
	}

	if len(secrets) != 1 {
		b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name)
		return nil, fmt.Errorf("More than 1 secret found with the newly created service account, this is unexpected for the prupose of this plugin, please try again")
	}

	rb, err := b.kubernetesService.CreateRoleBinding(pluginConfig, namespace, sa.Name, roleName)

	if err != nil {
		b.Logger().Error(fmt.Sprintf("Error setting up Kubernetes role binding for SA %s: %s", sa.Name, err))
		b.kubernetesService.DeleteServiceAccount(pluginConfig, sa.Namespace, sa.Name)
		return nil, err
	}

	b.Logger().Info(fmt.Sprintf("Service account '%s' created with rolebinding '%s'", sa.Name, rb.Name))

	dur, err := time.ParseDuration(fmt.Sprintf("%ds", ttl))
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("ttl: %d could not be parse due to error: %s", ttl, err), err)
	}

	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyCACert:              secrets[0].CACert,
		keyNamespace:           secrets[0].Namespace,
		keyServiceAccountToken: secrets[0].Token,
		keyServiceAccountName:  sa.Name,
		keyRoleBindingName:     rb.Name,
		keyKubeConfig:          generateKubeConfig(pluginConfig, secrets[0].CACert, secrets[0].Token, sa.Name, namespace),
	}, map[string]interface{}{})

	// set up TTL for secret so it gets automatically revoked
	resp.Secret.LeaseOptions.TTL = dur
	resp.Secret.LeaseOptions.MaxTTL = dur
	resp.Secret.LeaseOptions.Renewable = false
	resp.Secret.TTL = dur
	resp.Secret.MaxTTL = dur
	resp.Secret.Renewable = false

	return resp, nil
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// reload plugin config on every call to prevent stale config
	pluginConfig, err := loadPluginConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	b.Logger().Info("revoking a service account")

	namespace := d.Get(keyNamespace).(string)
	serviceAccountName := d.Get(keyServiceAccountName).(string)
	roleBindingName := d.Get(keyRoleBindingName).(string)

	b.Logger().Info(fmt.Sprintf("deleting role binding with name: %s in namespace: %s", roleBindingName, namespace))
	err = b.kubernetesService.DeleteRoleBinding(pluginConfig, namespace, roleBindingName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted role binding with name: %s in namespace: %s", roleBindingName, namespace))

	b.Logger().Info(fmt.Sprintf("deleting service account with name: %s in namespace: %s", serviceAccountName, namespace))
	err = b.kubernetesService.DeleteServiceAccount(pluginConfig, namespace, serviceAccountName)
	if err != nil {
		return nil, err
	}
	b.Logger().Info(fmt.Sprintf("deleted service account with name: %s in namespace: %s", serviceAccountName, namespace))

	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		keyServiceAccountName: serviceAccountName,
	}, map[string]interface{}{})

	return resp, nil
}

// getCluterRoleName is a helper function to pull out the correct cluster role from the plugin configuration for the saType
func getClusterRoleName(pluginConfig *PluginConfig, saType string) (string, error) {
	switch saType {
	case "admin":
		return pluginConfig.AdminRole, nil
	case "editor":
		return pluginConfig.EditorRole, nil
	case "viewer":
		return pluginConfig.ViewerRole, nil
	}

	return "", fmt.Errorf("Service Account type '%s' is not a valid type", saType)
}

func generateKubeConfig(pluginConfig *PluginConfig, caCert string, token string, name string, namespace string) string {
	return fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
contexts:
- context:
    cluster: %s
    namespace: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    token: %s`, base64Encode(caCert), pluginConfig.Host, name, name, namespace, name, name, name, name, token)
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
