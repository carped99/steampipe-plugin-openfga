package main

import (
	"github.com/carped99/steampipe-plugin-openfga/acl"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: acl.Plugin,
	})
}
