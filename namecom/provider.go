package namecom

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/namedotcom/go/v4/namecom"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECOM_USER", nil),
				Description: "Name.com API Username",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECOM_TOKEN", nil),
				Description: "Name.com API Token Value",
			},
			"test": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Use Name.com Test API",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"namecom_record":      resourceRecord(),
			"namecom_nameservers": resourceNameservers(),
			"namecom_dnssec":      resourceDNSSEC(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var n *namecom.NameCom
	if d.Get("test").(bool) {
		n = namecom.Test(d.Get("username").(string), d.Get("token").(string))
	} else {
		n = namecom.New(d.Get("username").(string), d.Get("token").(string))
	}
	return n, nil
}
