package servian

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const keyMaxTTL = "max_ttl"
const keyTokenReviewerJWT = "token_reviewer_jwt"
const keyKubernetesHost = "kubernetes_host"
const keyKubernetesCACert = "kubernetes_ca_cert"
const keyAllowedRoles = "allowed_roles"
const keyAllowedClusterRoles = "allowed_cluster_roles"

const configPath = "config"

type PluginConfig struct {
	MaxTTL              int      `json:"max_ttl"`
	TokenReviewerJWT    string   `json:"token_reviewer_jwt"`
	KubernetesHost      string   `json:"kubernetes_host"`
	KubernetesCACert    string   `json:"kubernetes_ca_cert"`
	AllowedRoles        []string `json:"allowed_roles"`
	AllowedClusterRoles []string `json:"allowed_cluster_roles"`
}

func configK8sServiceAccount(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			keyMaxTTL: {
				Type:        framework.TypeInt,
				Description: "Time to live for the credentials returned.",
			},
			keyTokenReviewerJWT: {
				Type: framework.TypeString,
				Description: `A service account JWT used to access the
TokenReview API to validate other JWTs during login. If not set
the JWT used for login will be used to access the API.`,
			},
			keyKubernetesHost: {
				Type:        framework.TypeString,
				Description: "Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server.",
			},
			keyKubernetesCACert: {
				Type:        framework.TypeString,
				Description: "PEM encoded CA cert for use by the TLS client used to talk with the API.",
			},
			keyAllowedRoles: {
				Type:        framework.TypeStringSlice,
				Description: "Kubernetes roles that can be assigned to service accounts created by this plugin.",
			},
			keyAllowedClusterRoles: {
				Type:        framework.TypeStringSlice,
				Description: "Kubernetes cluster roles that can be assigned to service accounts created by this plugin.",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.handleConfigWrite,
				Summary:  "Configure the plugin",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.handleConfigWrite,
				Summary:  "Configure the plugin",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleConfigRead,
				Summary:  "Read plugin configuration",
			},
		},
	}
}

func (b *backend) handleConfigWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	ttl := d.Get(keyMaxTTL).(int)
	tokenReviewerJWT := d.Get(keyTokenReviewerJWT).(string)
	kubernetesHost := d.Get(keyKubernetesHost).(string)
	kubernetesCACert := d.Get(keyKubernetesCACert).(string)
	allowedRoles := d.Get(keyAllowedRoles).([]string)
	allowedClusterRoles := d.Get(keyAllowedClusterRoles).([]string)
	b.Logger().Info(fmt.Sprintf("TTL specified is: %d", ttl))
	config := PluginConfig{
		MaxTTL:              ttl,
		TokenReviewerJWT:    tokenReviewerJWT,
		KubernetesHost:      kubernetesHost,
		KubernetesCACert:    kubernetesCACert,
		AllowedRoles:        allowedRoles,
		AllowedClusterRoles: allowedClusterRoles,
	}
	entry, err := logical.StorageEntryJSON(configPath, config)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *backend) handleConfigRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if config, err := b.config(ctx, req.Storage); err != nil {
		return nil, err
	} else if config == nil {
		return nil, nil
	} else {
		resp := &logical.Response{
			Data: map[string]interface{}{
				keyMaxTTL:              config.MaxTTL,
				keyTokenReviewerJWT:    config.TokenReviewerJWT,
				keyKubernetesHost:      config.KubernetesHost,
				keyKubernetesCACert:    config.KubernetesCACert,
				keyAllowedRoles:        config.AllowedRoles,
				keyAllowedClusterRoles: config.AllowedClusterRoles,
			},
		}
		return resp, nil
	}
}

func (b *backend) config(ctx context.Context, s logical.Storage) (*PluginConfig, error) {
	raw, err := s.Get(ctx, configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	conf := &PluginConfig{}
	if err := json.Unmarshal(raw.Value, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
