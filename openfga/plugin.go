package openfga

import (
	"context"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func Plugin(ctx context.Context) *plugin.Plugin {
	return &plugin.Plugin{
		Name:             "openfga_fdw",
		DefaultTransform: transform.FromGo().NullIfZero(),
		//DefaultIgnoreConfig: &plugin.IgnoreConfig{
		//	ShouldIgnoreErrorFunc: isNotFoundError,
		//},
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			//Schema:      ConfigSchema,
		},
		TableMap: map[string]*plugin.Table{
			"sys_acl_permission": tableAclPermission(ctx),
		},
	}
}
