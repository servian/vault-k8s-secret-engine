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
const keyAdminRole = "admin_role"
const keyEditorRole = "editor_role"
const keyViewerRole = "viewer_role"
const keyJWT = "jwt"
const keyCACert = "ca_cert"
const keyBaseUrl = "base_url"
const configPath = "config"

type PluginConfig struct {
	MaxTTL            int    `json:"max_ttl"`
	AdminRole         string `json:"admin_role"`
	EditorRole        string `json:"editor_role"`
	ViewerRole        string `json:"viewer_role"`
	ServiceAccountJWT string `json:"jwt"`
	CACert            string `json:"ca_cert"`
	BaseUrl           string `json:"base_url"`
}

func configurePlugin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			keyMaxTTL: {
				Type:        framework.TypeInt,
				Description: "Time to live for the credentials returned.",
				Default:     1800, // 30 minutes
			},
			keyAdminRole: {
				Type:        framework.TypeString,
				Description: "Name of Kubernetes Admin ClusterRole that can be assigned to service accounts created by this plugin.",
				Required:    true,
			},
			keyEditorRole: {
				Type:        framework.TypeString,
				Description: "Name of Kubernetes Editor ClusterRole that can be assigned to service accounts created by this plugin.",
				Required:    true,
			},
			keyViewerRole: {
				Type:        framework.TypeString,
				Description: "Name of Kubernetes Viewer ClusterRole that can be assigned to service accounts created by this plugin.",
				Required:    true,
			},
			keyJWT: {
				Type:        framework.TypeString,
				Description: "JTW for the service account used to create and remove credentials in the Kubernetes Cluster",
				Required:    true,
			},
			keyCACert: {
				Type:        framework.TypeString,
				Description: "CA cert from the Kubernetes Cluster, to validate the connection",
				Required:    true,
			},
			keyBaseUrl: {
				Type:        framework.TypeString,
				Description: "URL for kubernetes cluster for vault to use to comunicate to k8s. https://{url}:{port}",
				Required:    true,
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
	adminRole := d.Get(keyAdminRole).(string)
	editorRole := d.Get(keyEditorRole).(string)
	viewerRole := d.Get(keyViewerRole).(string)
	jwt := d.Get(keyJWT).(string)
	cacert := d.Get(keyCACert).(string)
	baseurl := d.Get(keyBaseUrl).(string)

	_, err := url.Parse(baseurl)
	if err != nil {
		return logical.ErrorResponse("baseurl '%s' not a valid host: %s", baseurl, err), err
	}

	b.Logger().Info(fmt.Sprintf("MaxTTL specified is: %d", ttl))
	config := PluginConfig{
		MaxTTL:            ttl,
		AdminRole:         adminRole,
		EditorRole:        editorRole,
		ViewerRole:        viewerRole,
		ServiceAccountJWT: jwt,
		CACert:            cacert,
		BaseUrl:           baseurl,
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
				keyMaxTTL:     config.MaxTTL,
				keyAdminRole:  config.AdminRole,
				keyEditorRole: config.EditorRole,
				keyViewerRole: config.ViewerRole,
				keyJWT:        config.ServiceAccountJWT,
				keyCACert:     config.CACert,
				keyBaseUrl:    config.BaseUrl,
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
