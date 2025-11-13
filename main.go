package main

import (
	"github.com/gaia3d/steampipe-plugin-acl/acl"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: acl.Plugin,
	})
}
