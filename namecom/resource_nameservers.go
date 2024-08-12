package namecom

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/namedotcom/go/v4/namecom"
	"github.com/pkg/errors"
)

func resourceNameservers() *schema.Resource {
	return &schema.Resource{
		Create: resourceNameserversSet,
		Read:   resourceNameserversGet,
		Update: resourceNameserversSet,
		Delete: resourceNameserversDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNameserversImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone is the domain name to set the nameservers for.",
			},
			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DomainName is the domain name to set the nameservers for.",
			},
			"nameservers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Namesevers is a list of the nameservers to set. Nameservers should already be set up and hosting the zone properly as some registries will verify before allowing the change.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceNameserversSet(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)

	domainName := d.Get("domain_name").(string)

	request := namecom.SetNameserversRequest{
		DomainName: domainName,
	}
	for _, nameserver := range d.Get("nameservers").([]interface{}) {
		request.Nameservers = append(request.Nameservers, nameserver.(string))
	}

	if domain, err := client.SetNameservers(&request); err != nil {
		return errors.Wrap(err, "Error SetNameservers")
	} else {
		d.SetId(domain.DomainName)
		d.Set("nameservers", domain.Nameservers)
		d.Set("zone", domain.DomainName)
	}

	return nil
}

func resourceNameserversGet(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)

	domainName := d.Get("domain_name").(string)

	request := namecom.GetDomainRequest{
		DomainName: domainName,
	}

	if domain, err := client.GetDomain(&request); err != nil {
		return errors.Wrap(err, "Error GetDomain")
	} else {
		d.Set("nameservers", domain.Nameservers)
	}

	return nil
}

func resourceNameserversDelete(d *schema.ResourceData, m interface{}) error {
	err := resourceNameserversSet(d, m)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceNameserversImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*namecom.NameCom)

	request := namecom.GetDomainRequest{
		DomainName: d.Id(),
	}

	if domain, err := client.GetDomain(&request); err != nil {
		return nil, errors.Wrap(err, "Error GetDomain")
	} else {
		d.SetId(domain.DomainName)
		d.Set("domain_name", domain.DomainName)
		d.Set("zone", domain.DomainName)
		d.Set("nameservers", domain.Nameservers)
	}

	return []*schema.ResourceData{d}, nil
}
