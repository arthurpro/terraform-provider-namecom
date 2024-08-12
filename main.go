package main

import (
	"github.com/arthurpro/terraform-provider-namecom/namecom"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: namecom.Provider,
	})
}
