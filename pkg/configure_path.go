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
const keyHost = "host"
const keyDefaultTTL = "default_ttl"

const configPath = "config"

// PluginConfig contains all the configuration for the plugin
type PluginConfig struct {
	MaxTTL            int    `json:"max_ttl"`
	DefaulTTL         int    `json:"default_ttl"`
	AdminRole         string `json:"admin_role"`
	EditorRole        string `json:"editor_role"`
	ViewerRole        string `json:"viewer_role"`
	ServiceAccountJWT string `json:"jwt"`
	CACert            string `json:"ca_cert"`
	Host              string `json:"host"`
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
			keyDefaultTTL: {
				Type:        framework.TypeInt,
				Description: "Deafult time to live for when a user does not provide a TTL",
				Default:     600, // 10 minutes
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
			keyHost: {
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

	config := PluginConfig{
		MaxTTL:            d.Get(keyMaxTTL).(int),
		DefaulTTL:         d.Get(keyDefaultTTL).(int),
		AdminRole:         d.Get(keyAdminRole).(string),
		EditorRole:        d.Get(keyEditorRole).(string),
		ViewerRole:        d.Get(keyViewerRole).(string),
		ServiceAccountJWT: d.Get(keyJWT).(string),
		CACert:            d.Get(keyCACert).(string),
		Host:              d.Get(keyHost).(string),
	}

	err := config.Validate()

	if err != nil {
		return logical.ErrorResponse("Configuration not valid: %s", err), err
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
				keyDefaultTTL: config.DefaulTTL,
				keyAdminRole:  config.AdminRole,
				keyEditorRole: config.EditorRole,
				keyViewerRole: config.ViewerRole,
				keyJWT:        config.ServiceAccountJWT,
				keyCACert:     config.CACert,
				keyHost:       config.Host,
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

// Validate validates the plugin config by checking all required values are correct
func (c *PluginConfig) Validate() error {

	_, err := url.Parse(c.Host)
	if err != nil {
		return fmt.Errorf("Host '%s' not a valid host: %s", c.Host, err)
	}

	if c.AdminRole == "" {
		return fmt.Errorf("%s can not be empty", keyAdminRole)
	}

	if c.EditorRole == "" {
		return fmt.Errorf("%s can not be empty", keyEditorRole)
	}

	if c.ViewerRole == "" {
		return fmt.Errorf("%s can not be empty", keyViewerRole)
	}

	if c.ServiceAccountJWT == "" {
		return fmt.Errorf("%s can not be empty", keyJWT)
	}

	if c.CACert == "" {
		return fmt.Errorf("%s can not be empty", keyCACert)
	}

	return nil
}
