package servian

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const keyRoleName = "role_name"
const keyClusterRoleName = "cluster_role_name"
const keyKubeConfigPath = "kube_config_path"
const keyTTLSeconds = "ttl"
const keyNamespace = "namespace"
const keyServiceAccountToken = "service_account_token"
const keyServiceAccountUID = "service_account_uid"
const keyServiceAccountName = "service_account_name"
const keyRoleBindingName = "role_binding_name"
const keySAType = "type"

func getAllowedSATypes() []string {
	return []string{"admin", "editor", "viewer"}
}

func readSecret(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account/" + framework.GenericNameRegex(keyNamespace) + "/" + framework.GenericNameRegex(keySAType),
		Fields: map[string]*framework.FieldSchema{
			keySAType: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Type of the service account to be created. Accepted types: %s", strings.Join(getAllowedSATypes(), ", ")),
				Required:    true,
			},
			keyNamespace: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The namespace under which the service account should be created",
				Required:    true,
			},
			keyTTLSeconds: &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: "The time to live for the token in seconds",
				Default:     600,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleReadForRole,
				Summary:  "Create new service account credentials",
			},
		},
	}
}

func (b *backend) handleReadForRole(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if d != nil {
		saType := strings.ToLower(d.Get(keySAType).(string))
		saTypeFound := false
		for _, allowedType := range getAllowedSATypes() {
			if saType == allowedType {
				saTypeFound = true
				break
			}
		}

		if saTypeFound == false {
			return nil, fmt.Errorf("Service account type '%s' not one of the allowed types: %s", saType, strings.Join(getAllowedSATypes(), ", "))
		}

		namespace := d.Get(keyNamespace).(string)
		ttl := d.Get(keyTTLSeconds).(int)
		return b.createSecret(ctx, req.Storage, saType, namespace, ttl)
	}

	return nil, fmt.Errorf("could not find a role name to associate with the service account")

}
