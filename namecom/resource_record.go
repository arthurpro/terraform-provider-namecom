package namecom

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/namedotcom/go/v4/namecom"
	"github.com/pkg/errors"
)

func resourceRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordCreate,
		ReadContext:   resourceRecordRead,
		UpdateContext: resourceRecordUpdate,
		DeleteContext: resourceRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordImporter,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Zone is the domain name that the record belongs to.",
			},
			"host": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "@",
				DiffSuppressFunc: suppressHost,
				Description:      "Host is the hostname relative to the zone.",
			},
			"fqdn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "FQDN is the Fully Qualified Domain Name.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type is one of the following: A, AAAA, ANAME, CNAME, MX, NS, SRV, or TXT.",
				ValidateFunc: validation.StringInSlice([]string{"A", "AAAA", "ANAME", "CNAME", "MX", "NS", "SRV", "TXT"}, true),
			},
			"answer": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Answer is either the IP address for A or AAAA records; the target for ANAME, CNAME, MX, or NS records; the text for TXT records. For SRV records, answer has the following format: "{weight} {port} {target}" e.g. "1 5061 sip.example.org".`,
			},
			"ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "TTL is the time this record can be cached for in seconds. Minimum TTL is 300, or 5 minutes.",
			},
			"priority": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: suppressPriority,
				Description:      "Priority is only required for MX and SRV records, it is ignored for all others.",
			},
		},
	}
}

func resourceRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)
	record := getRecordResourceData(d)

	if record, err := client.CreateRecord(&record); err != nil {
		return diag.FromErr(errors.Wrap(err, "Error GetRecord"))
	} else {
		err = setRecordResourceData(d, record)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	recordID, err := recordId(d.Id())
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error parsing RecordID, should be int32"))
	}

	request := namecom.GetRecordRequest{
		DomainName: d.Get("zone").(string),
		ID:         recordID,
	}

	record, err := client.GetRecord(&request)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error GetRecord"))
	}

	err = setRecordResourceData(d, record)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	recordID, err := recordId(d.Id())
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error parsing RecordID, should be int32"))
	}

	record := getRecordResourceData(d)
	record.ID = recordID

	if record, err := client.UpdateRecord(&record); err != nil {
		return diag.FromErr(errors.Wrap(err, "Error UpdateRecord"))
	} else {
		err = setRecordResourceData(d, record)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*namecom.NameCom)

	recordID, err := recordId(d.Id())
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error parsing RecordID, should be int32"))
	}

	deleteRequest := namecom.DeleteRecordRequest{
		DomainName: d.Get("zone").(string),
		ID:         recordID,
	}

	_, err = client.DeleteRecord(&deleteRequest)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Error DeleteRecord"))
	}

	d.SetId("")
	return nil
}

func resourceRecordImporter(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*namecom.NameCom)

	idAttr := strings.SplitN(d.Id(), "/", 2)
	var domain string
	var record string
	if len(idAttr) == 2 {
		domain = idAttr[0]
		record = idAttr[1]
	} else {
		return nil, fmt.Errorf(`invalid id %q specified, should be in format "DomainName/RecordID" for import`, d.Id())
	}
	recordID, err := recordId(record)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing RecordID, should be int")
	}

	request := namecom.GetRecordRequest{
		DomainName: domain,
		ID:         recordID,
	}

	r, err := client.GetRecord(&request)
	if err != nil {
		return nil, errors.Wrap(err, "Error GetRecord")
	}

	err = setRecordResourceData(d, r)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func setRecordResourceData(d *schema.ResourceData, r *namecom.Record) error {
	host := r.Host
	if host == "" {
		host = "@"
	}
	d.SetId(fmt.Sprintf("%d", r.ID))

	err := d.Set("zone", r.DomainName)
	if err != nil {
		return errors.New("Error setting zone")
	}

	err = d.Set("host", host)
	if err != nil {
		return errors.New("Error setting host")
	}

	err = d.Set("type", r.Type)
	if err != nil {
		return errors.New("Error setting type")
	}

	err = d.Set("fqdn", r.Fqdn)
	if err != nil {
		return errors.New("Error setting fqdn")
	}

	err = d.Set("answer", r.Answer)
	if err != nil {
		return errors.New("Error setting answer")
	}

	err = d.Set("ttl", r.TTL)
	if err != nil {
		return errors.New("Error setting ttl")
	}

	err = d.Set("priority", r.Priority)
	if err != nil {
		return errors.New("Error setting priority")
	}

	return nil
}

func getRecordResourceData(d *schema.ResourceData) namecom.Record {
	host := d.Get("host").(string)
	if host == "@" {
		host = ""
	}
	record := namecom.Record{
		DomainName: d.Get("zone").(string),
		Host:       host,
		Type:       strings.ToUpper(d.Get("type").(string)),
		Answer:     d.Get("answer").(string),
	}

	ttl := d.Get("ttl").(int)
	if ttl > 0 {
		record.TTL = uint32(ttl)
	}

	priority := d.Get("priority").(int)
	if priority > 0 {
		record.Priority = uint32(priority)
	}

	return record
}

func recordId(id string) (int32, error) {
	recordID, err := strconv.ParseInt(id, 10, 32)
	return int32(recordID), err
}

func suppressHost(k, old, new string, d *schema.ResourceData) bool {
	if old == "@" && new == "" {
		return true
	}
	return false
}

func suppressPriority(k, old, new string, d *schema.ResourceData) bool {
	recordType := strings.ToUpper(d.Get("type").(string))
	if recordType != "MX" && recordType != "SRV" {
		return true
	}
	return false
}
