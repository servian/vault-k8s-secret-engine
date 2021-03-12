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
const keyPemKeys = "pem_keys"
const keyIssuer = "issuer"
const keyDisableISSValidation = "disable_iss_validation"
const keyDisableLocalCaJWT = "disable_local_ca_jwt"
const keyAllowedRoles = "allowed_roles"
const keyAllowedClusterRoles = "allowed_cluster_roles"

const configPath = "config"

type PluginConfig struct {
	MaxTTL               int      `json:"max_ttl"`
	TokenReviewerJWT     string   `json:"token_reviewer_jwt"`
	KubernetesHost       string   `json:"kubernetes_host"`
	KubernetesCaCert     string   `json:"kubernetes_ca_cert"`
	PemKeys              []string `json:"pem_keys"`
	Issuer               string   `json:"issuer"`
	DisableISSValidation bool     `json:"disable_iss_validation"`
	DisableLocalCaJWT    bool     `json:"disable_local_ca_jwt"`
	AllowedRoles         []string `json:"allowed_roles"`
	AllowedClusterRoles  []string `json:"allowed_cluster_roles"`
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
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Token Reviewer JWT",
				},
			},
			keyKubernetesHost: {
				Type:        framework.TypeString,
				Description: "Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server.",
			},
			keyKubernetesCACert: {
				Type:        framework.TypeString,
				Description: "PEM encoded CA cert for use by the TLS client used to talk with the API.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Kubernetes CA Certificate",
				},
			},
			keyPemKeys: {
				Type: framework.TypeCommaStringSlice,
				Description: `Optional list of PEM-formated public keys or certificates
used to verify the signatures of kubernetes service account
JWTs. If a certificate is given, its public key will be
extracted. Not every installation of Kuberentes exposes these keys.`,
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Service account verification keys",
				},
			},
			keyIssuer: {
				Type:        framework.TypeString,
				Description: "Optional JWT issuer. If no issuer is specified, then this plugin will use kubernetes.io/serviceaccount as the default issuer.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "JWT Issuer",
				},
			},
			keyDisableISSValidation: {
				Type:        framework.TypeBool,
				Description: "Disable JWT issuer validation. Allows to skip ISS validation.",
				Default:     false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Disable JWT Issuer Validation",
				},
			},
			keyDisableLocalCaJWT: {
				Type:        framework.TypeBool,
				Description: "Disable defaulting to the local CA cert and service account JWT when running in a Kubernetes pod",
				Default:     false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Disable use of local CA and service account JWT",
				},
			},
			keyAllowedRoles: {
				Type:        framework.TypeCommaStringSlice,
				Description: "Kubernetes roles that can be assigned to service accounts created by this plugin.",
			},
			keyAllowedClusterRoles: {
				Type:        framework.TypeCommaStringSlice,
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
	pemKeys := d.Get(keyPemKeys).([]string)
	issuer := d.Get(keyIssuer).(string)
	disableISSValidation := d.Get(keyDisableISSValidation).(bool)
	disableLocalCaJWT := d.Get(keyDisableLocalCaJWT).(bool)
	allowedRoles := d.Get(keyAllowedRoles).([]string)
	allowedClusterRoles := d.Get(keyAllowedClusterRoles).([]string)
	b.Logger().Info(fmt.Sprintf("TTL specified is: %d", ttl))
	config := PluginConfig{
		MaxTTL:               ttl,
		TokenReviewerJWT:     tokenReviewerJWT,
		KubernetesHost:       kubernetesHost,
		KubernetesCaCert:     kubernetesCACert,
		PemKeys:              pemKeys,
		Issuer:               issuer,
		DisableISSValidation: disableISSValidation,
		DisableLocalCaJWT:    disableLocalCaJWT,
		AllowedRoles:         allowedRoles,
		AllowedClusterRoles:  allowedClusterRoles,
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
				keyMaxTTL:               config.MaxTTL,
				keyTokenReviewerJWT:     config.TokenReviewerJWT,
				keyKubernetesHost:       config.KubernetesHost,
				keyKubernetesCACert:     config.KubernetesCaCert,
				keyPemKeys:              config.PemKeys,
				keyIssuer:               config.Issuer,
				keyDisableLocalCaJWT:    config.DisableLocalCaJWT,
				keyDisableISSValidation: config.DisableISSValidation,
				keyAllowedRoles:         config.AllowedRoles,
				keyAllowedClusterRoles:  config.AllowedClusterRoles,
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
