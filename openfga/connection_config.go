package openfga

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type Config struct {
	Endpoint             *string `hcl:"endpoint,required"`
	ApiToken             *string `hcl:"api_token,optional"`
	StoreId              *string `hcl:"store_id,required"`
	AuthorizationModelId *string `hcl:"authorization_model_id,optional"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"api_url": {
		Type: schema.TypeString,
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

func getConfig(connection *plugin.Connection) (*Config, error) {
	if connection == nil || connection.Config == nil {
		return &Config{}, nil
	}
	config, ok := connection.Config.(*Config)
	if !ok {
		return &Config{}, nil
	}
	return config, nil
}
