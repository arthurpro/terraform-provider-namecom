package namecom

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/namedotcom/go/v4/namecom"
	"github.com/pkg/errors"
)

func resourceDNSSEC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSSECCreate,
		ReadContext:   resourceDNSSECRead,
		DeleteContext: resourceDNSSECDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDNSSECImporter,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Zone is the domain name that the DNSSEC belongs to.",
			},
			"key_tag": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "KeyTag contains the key tag value of the DNSKEY RR that validates this signature.",
			},
			"algorithm": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Algorithm is an integer identifying the algorithm used for signing.",
			},
			"digest_type": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "DigestType is an integer identifying the algorithm used to create the digest.",
			},
			"digest": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Digest is a digest of the DNSKEY RR that is registered with the registry.",
			},
		},
	}
}

// resourceDNSSECCreate creates a new DNSSEC in the Name.com API.
func resourceDNSSECCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, err := meta.(*namecom.NameCom).CreateDNSSEC(
		&namecom.DNSSEC{
			DomainName: data.Get("zone").(string),
			KeyTag:     data.Get("key_tag").(int32),
			Algorithm:  data.Get("algorithm").(int32),
			DigestType: data.Get("digest_type").(int32),
			Digest:     data.Get("digest").(string),
		},
	)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error CreateDNSSEC"))
	}

	domainNameString, ok := data.Get("zone").(string)
	if !ok {
		return diag.FromErr(errors.New("Error getting zone"))
	}

	data.SetId(domainNameString)

	return resourceDNSSECRead(ctx, data, meta)
}

// resourceDNSSECImporter import existing DNSSEC from the Name.com API.
func resourceDNSSECImporter(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, ok := meta.(*namecom.NameCom)
	if !ok {
		return nil, errors.New("Error getting client")
	}

	importDomainName, importDigest, err := resourceDNSSECImporterParseID(data.Id())
	if err != nil {
		return nil, err
	}

	request := namecom.GetDNSSECRequest{
		DomainName: importDomainName,
		Digest:     importDigest,
	}

	DNSSEC, err := client.GetDNSSEC(&request)
	if err != nil {
		return nil, errors.Wrap(err, "Error GetDNSSECRequest")
	}

	err = data.Set("zone", DNSSEC.DomainName)
	if err != nil {
		return nil, errors.Wrap(err, "Error setting zone")
	}

	err = data.Set("key_tag", int(DNSSEC.KeyTag))
	if err != nil {
		return nil, errors.Wrap(err, "Error setting key_tag")
	}

	err = data.Set("algorithm", int(DNSSEC.Algorithm))
	if err != nil {
		return nil, errors.Wrap(err, "Error setting algorithm")
	}

	err = data.Set("digest_type", int(DNSSEC.DigestType))
	if err != nil {
		return nil, errors.Wrap(err, "Error setting digest_type")
	}

	err = data.Set("digest", DNSSEC.Digest)
	if err != nil {
		return nil, errors.Wrap(err, "Error setting digest")
	}

	data.SetId(importDomainName)

	return []*schema.ResourceData{data}, nil
}

// resourceDNSSECImporterParseID parses the ID of the DNSSEC.
func resourceDNSSECImporterParseID(id string) (domainName, digest string, err error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("unexpected format of ID, expected Zone/Digest")
	}

	return parts[0], parts[1], nil
}

// resourceDNSSECRead reads a DNSSEC from the Name.com API.
func resourceDNSSECRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, ok := meta.(*namecom.NameCom)
	if !ok {
		return diag.FromErr(errors.New("Error getting client"))
	}

	domainNameString, ok := data.Get("zone").(string)
	if !ok {
		return diag.FromErr(errors.New("Error getting zone"))
	}

	digestString, ok := data.Get("digest").(string)
	if !ok {
		return diag.FromErr(errors.New("Error getting digest"))
	}

	request := namecom.GetDNSSECRequest{
		DomainName: domainNameString,
		Digest:     digestString,
	}

	DNSSEC, err := client.GetDNSSEC(&request)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error GetDNSSECRequest"))
	}

	err = data.Set("zone", DNSSEC.DomainName)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error setting zone"))
	}

	err = data.Set("key_tag", int(DNSSEC.KeyTag))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error setting key_tag"))
	}

	err = data.Set("algorithm", int(DNSSEC.Algorithm))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error setting algorithm"))
	}

	err = data.Set("digest_type", int(DNSSEC.DigestType))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error setting digest_type"))
	}

	err = data.Set("digest", DNSSEC.Digest)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error setting digest"))
	}

	return nil
}

// resourceDNSSECDelete deletes a DNSSEC from the Name.com API.
func resourceDNSSECDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, ok := meta.(*namecom.NameCom)
	if !ok {
		return diag.FromErr(errors.New("Error getting client"))
	}

	domainNameString, ok := data.Get("zone").(string)
	if !ok {
		return diag.FromErr(errors.New("Error getting zone"))
	}

	digestString, ok := data.Get("digest").(string)
	if !ok {
		return diag.FromErr(errors.New("Error getting digest"))
	}

	deleteRequest := namecom.DeleteDNSSECRequest{
		DomainName: domainNameString,
		Digest:     digestString,
	}

	_, err := client.DeleteDNSSEC(&deleteRequest)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error DeleteDNSSEC"))
	}

	data.SetId("")

	return nil
}
