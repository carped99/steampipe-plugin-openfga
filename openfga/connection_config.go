package openfga

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type Config struct {
	Endpoint           string  `hcl:"endpoint" env:"OPENFGA_ENDPOINT"`
	UseTLS             *bool   `hcl:"use_tls" env:"OPENFGA_USE-TLS"`                           // 기본은 false (내부망)
	CACertPath         *string `hcl:"ca_cert_path" env:"OPENFGA_CA-CERT-PATH"`                 // TLS 시 CA 경로
	InsecureSkipVerify *bool   `hcl:"insecure_skip_verify" env:"OPENFGA_INSECURE-SKIP-VERIFY"` // 필요시만

	ApiToken             *string `hcl:"api_token" env:"OPENFGA_API-TOKEN"`
	StoreId              *string `hcl:"store_id" env:"OPENFGA_STORE-ID"`
	AuthorizationModelId *string `hcl:"authorization_model_id" env:"OPENFGA_AUTHORIZATION-MODEL-ID"`
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
