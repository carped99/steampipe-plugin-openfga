package main

import (
	"github.com/carped99/steampipe-plugin-openfga/openfga"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: openfga.Plugin,
	})
}
