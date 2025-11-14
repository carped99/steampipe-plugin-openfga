package openfga

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type Config struct {
	Endpoint           string  `hcl:"endpoint"`
	UseTLS             *bool   `hcl:"use_tls"`              // 기본은 false (내부망)
	CACertPath         *string `hcl:"ca_cert_path"`         // TLS 시 CA 경로
	InsecureSkipVerify *bool   `hcl:"insecure_skip_verify"` // 필요시만

	ApiToken             *string `hcl:"api_token"`
	StoreId              *string `hcl:"store_id"`
	AuthorizationModelId *string `hcl:"authorization_model_id"`
}

func ConfigInstance() any {
	return &Config{}
}

var ConfigSchema = map[string]*schema.Attribute{
	"endpoint": {
		Type: schema.TypeString,
	},
	"use_tls": {
		Type: schema.TypeBool,
	},
	"ca_cert_path": {
		Type: schema.TypeString,
	},
	"insecure_skip_verify": {
		Type: schema.TypeBool,
	},
	"api_token": {
		Type: schema.TypeString,
	},
	"store_id": {
		Type: schema.TypeString,
	},
	"authorization_model_id": {
		Type: schema.TypeString,
	},
}

func getConfig(connection *plugin.Connection) Config {
	if connection == nil || connection.Config == nil {
		return Config{}
	}
	config, ok := connection.Config.(Config)
	if !ok {
		return Config{}
	}
	return config
}
