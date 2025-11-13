package acl

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type aclConfig struct {
	ApiUrl               *string `hcl:"api_url"`
	ApiToken             *string `hcl:"api_token"`
	StoreId              *string `hcl:"store_id"`
	AuthorizationModelId *string `hcl:"authorization_model_id"`
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

func getConfig(connection *plugin.Connection) *aclConfig {
	if connection == nil || connection.Config == nil {
		return &aclConfig{}
	}
	config, ok := connection.Config.(*aclConfig)
	if !ok {
		return &aclConfig{}
	}
	return config
}
