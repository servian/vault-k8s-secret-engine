package servian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const keyMaxTTL = "max_ttl"
const keyAllowedRoles = "allowed_roles"
const keyAllowedClusterRoles = "allowed_cluster_roles"
const keyJWT = "jwt"
const keyCACert = "ca_cert"
const keyBaseUrl = "base_url"
const configPath = "config"

type PluginConfig struct {
	MaxTTL              int      `json:"max_ttl"`
	AllowedRoles        []string `json:"allowed_roles"`
	AllowedClusterRoles []string `json:"allowed_cluster_roles"`
	ServiceAccountJWT   string   `json:"jwt"`
	CACert              string   `json:"ca_cert"`
	BaseUrl             string   `json:"base_url"`
	VersionedAPIPath    string   `json:"versioned_api_path"`
}

func configurePlugin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			keyMaxTTL: {
				Type:        framework.TypeInt,
				Description: "Time to live for the credentials returned.",
			},
			keyAllowedRoles: {
				Type:        framework.TypeCommaStringSlice,
				Description: "Kubernetes roles that can be assigned to service accounts created by this plugin.",
			},
			keyAllowedClusterRoles: {
				Type:        framework.TypeCommaStringSlice,
				Description: "Kubernetes cluster roles that can be assigned to service accounts created by this plugin.",
			},
			keyJWT: {
				Type:        framework.TypeString,
				Description: "JTW for the service account used to create and remove credentials in the Kubernetes Cluster",
			},
			keyCACert: {
				Type:        framework.TypeString,
				Description: "CA cert from the Kubernetes Cluster, to validate the connection",
			},
			keyBaseUrl: {
				Type:        framework.TypeString,
				Description: "URL for kubernetes cluster for vault to use to comunicate to k8s. https://{url}:{port}",
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
	allowedRoles := d.Get(keyAllowedRoles).([]string)
	allowedClusterRoles := d.Get(keyAllowedClusterRoles).([]string)
	jwt := d.Get(keyJWT).(string)
	cacert := d.Get(keyCACert).(string)
	baseurl := d.Get(keyBaseUrl).(string)

	_, err := url.Parse(baseurl)
	if err != nil {
		return logical.ErrorResponse("baseurl '%s' not a valid host: %s", baseurl, err), err
	}

	b.Logger().Info(fmt.Sprintf("MaxTTL specified is: %d", ttl))
	config := PluginConfig{
		MaxTTL:              ttl,
		AllowedRoles:        allowedRoles,
		AllowedClusterRoles: allowedClusterRoles,
		ServiceAccountJWT:   jwt,
		CACert:              cacert,
		BaseUrl:             baseurl,
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
	if config, err := loadPluginConfig(ctx, req.Storage); err != nil {
		return nil, err
	} else if config == nil {
		return nil, nil
	} else {

		resp := &logical.Response{
			Data: map[string]interface{}{
				keyMaxTTL:              config.MaxTTL,
				keyAllowedRoles:        config.AllowedRoles,
				keyAllowedClusterRoles: config.AllowedClusterRoles,
				keyJWT:                 config.ServiceAccountJWT,
				keyCACert:              config.CACert,
				keyBaseUrl:             config.BaseUrl,
			},
		}
		return resp, nil
	}
}

// loadPluginConfig is a helper function to simplify the loading of plugin configuration from the logical store
func loadPluginConfig(ctx context.Context, s logical.Storage) (*PluginConfig, error) {
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
