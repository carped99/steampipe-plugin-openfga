package openfga

import (
	"context"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             "steampipe-plugin-acl",
		DefaultTransform: transform.FromGo().NullIfZero(),
		//DefaultIgnoreConfig: &plugin.IgnoreConfig{
		//	ShouldIgnoreErrorFunc: isNotFoundError,
		//},
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: func() any {
				return &Config{}
			},
		},
		TableMap: map[string]*plugin.Table{
			"sys_acl_permission": tableAclPermission(),
		},
	}
	return p
}
