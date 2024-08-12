package namecom

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/namedotcom/go/v4/namecom"
	"github.com/pkg/errors"
)

func resourceNameservers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNameserversSet,
		ReadContext:   resourceNameserversGet,
		UpdateContext: resourceNameserversSet,
		DeleteContext: resourceNameserversDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNameserversImporter,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone is the domain name to set the nameservers for.",
			},
			"nameservers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Nameservers is a list of the nameservers to set. Nameservers should already be set up and hosting the zone properly as some registries will verify before allowing the change.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceNameserversSet(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	domainName := d.Get("zone").(string)

	request := namecom.SetNameserversRequest{
		DomainName: domainName,
	}
	for _, nameserver := range d.Get("nameservers").([]interface{}) {
		request.Nameservers = append(request.Nameservers, nameserver.(string))
	}

	if domain, err := client.SetNameservers(&request); err != nil {
		return diag.FromErr(errors.Wrap(err, "Error SetNameservers"))
	} else {
		d.SetId(domain.DomainName)

		err = d.Set("zone", domain.DomainName)
		if err != nil {
			return diag.FromErr(errors.New("Error setting zone"))
		}

		err = d.Set("nameservers", domain.Nameservers)
		if err != nil {
			return diag.FromErr(errors.New("Error setting nameservers"))
		}
	}

	return nil
}

func resourceNameserversGet(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	domainName := d.Get("zone").(string)

	request := namecom.GetDomainRequest{
		DomainName: domainName,
	}

	if domain, err := client.GetDomain(&request); err != nil {
		return diag.FromErr(errors.Wrap(err, "Error GetDomain"))
	} else {
		err = d.Set("nameservers", domain.Nameservers)
		if err != nil {
			return diag.FromErr(errors.New("Error setting nameservers"))
		}
	}

	return nil
}

func resourceNameserversDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	domainName := d.Id()

	request := namecom.SetNameserversRequest{
		DomainName:  domainName,
		Nameservers: defaultNameservers(),
	}

	if _, err := client.SetNameservers(&request); err != nil {
		return diag.FromErr(errors.Wrap(err, "Error SetNameservers"))
	}

	d.SetId("")

	return nil
}

func resourceNameserversImporter(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*namecom.NameCom)

	request := namecom.GetDomainRequest{
		DomainName: d.Id(),
	}

	if domain, err := client.GetDomain(&request); err != nil {
		return nil, errors.Wrap(err, "Error GetDomain")
	} else {
		d.SetId(domain.DomainName)

		err = d.Set("zone", domain.DomainName)
		if err != nil {
			return nil, errors.Wrap(err, "Error setting zone")
		}

		err = d.Set("nameservers", domain.Nameservers)
		if err != nil {
			return nil, errors.Wrap(err, "Error setting nameservers")
		}
	}

	return []*schema.ResourceData{d}, nil
}

func defaultNameservers() []string {
	return []string{
		"ns1.name.com",
		"ns2.name.com",
		"ns3.name.com",
		"ns4.name.com",
	}
}
