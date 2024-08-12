package namecom

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/namedotcom/go/v4/namecom"
	"github.com/pkg/errors"
)

func resourceRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceRecordCreate,
		Read:   resourceRecordRead,
		Update: resourceRecordUpdate,
		Delete: resourceRecordDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRecordImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone is the zone that the record belongs to.",
			},
			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DomainName is the zone that the record belongs to.",
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
				ValidateFunc: validation.StringInSlice([]string{"A", "AAAA", "ANAME", "CNAME", "MX", "NS", "SRV", "TXT"}, false),
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Answer is either the IP address for A or AAAA records; the target for ANAME, CNAME, MX, or NS records; the text for TXT records. For SRV records, answer has the following format: \"{weight} {port} {target}\" e.g. \"1 5061 sip.example.org\".",
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
			"data": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRecordCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)
	record := getRecordResourceData(d)

	if record, err := client.CreateRecord(&record); err != nil {
		return errors.Wrap(err, "Error GetRecord")
	} else {
		setRecordResourceData(d, record)
	}

	return nil
}

func resourceRecordRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)

	recordID, err := strconv.ParseInt(d.Id(), 10, 32)
	if err != nil {
		return errors.Wrap(err, "Error converting Record ID")
	}

	request := namecom.GetRecordRequest{
		DomainName: d.Get("domain_name").(string),
		ID:         int32(recordID),
	}

	record, err := client.GetRecord(&request)
	if err != nil {
		return errors.Wrap(err, "Error GetRecord")
	}

	setRecordResourceData(d, record)

	return nil
}

func resourceRecordUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)

	recordID, err := strconv.ParseInt(d.Id(), 10, 32)
	if err != nil {
		return errors.Wrap(err, "Error Parsing Record ID")
	}

	record := getRecordResourceData(d)
	record.ID = int32(recordID)

	if record, err := client.UpdateRecord(&record); err != nil {
		return errors.Wrap(err, "Error UpdateRecord")
	} else {
		setRecordResourceData(d, record)
	}

	return nil
}

func resourceRecordDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*namecom.NameCom)

	recordID, err := strconv.ParseInt(d.Id(), 10, 32)
	if err != nil {
		return errors.Wrap(err, "Error converting Record ID")
	}

	deleteRequest := namecom.DeleteRecordRequest{
		DomainName: d.Get("domain_name").(string),
		ID:         int32(recordID),
	}

	_, err = client.DeleteRecord(&deleteRequest)
	if err != nil {
		return errors.Wrap(err, "Error DeleteRecord")
	}

	d.SetId("")
	return nil
}

func resourceRecordImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*namecom.NameCom)

	idAttr := strings.SplitN(d.Id(), "/", 2)
	var domain string
	var record string
	if len(idAttr) == 2 {
		domain = idAttr[0]
		record = idAttr[1]
	} else {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"DomainName/RecordID\" for import", d.Id())
	}
	recordID, err := strconv.ParseInt(record, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "Error converting Record ID")
	}

	request := namecom.GetRecordRequest{
		DomainName: domain,
		ID:         int32(recordID),
	}

	r, err := client.GetRecord(&request)
	if err != nil {
		return nil, errors.Wrap(err, "Error GetRecord")
	}

	setRecordResourceData(d, r)

	return []*schema.ResourceData{d}, nil
}

func setRecordResourceData(d *schema.ResourceData, r *namecom.Record) {
	host := r.Host
	if host == "" {
		host = "@"
	}
	d.SetId(strconv.Itoa(int(r.ID)))
	d.Set("zone", r.DomainName)
	d.Set("domain_name", r.DomainName)
	d.Set("host", host)
	d.Set("type", r.Type)
	d.Set("fqdn", r.Fqdn)
	d.Set("value", r.Answer)
	d.Set("ttl", r.TTL)
	d.Set("priority", r.Priority)

	//data, _ := json.Marshal(r)
	//d.Set("data", string(data))
}

func getRecordResourceData(d *schema.ResourceData) namecom.Record {
	host := d.Get("host").(string)
	if host == "@" {
		host = ""
	}
	record := namecom.Record{
		DomainName: d.Get("zone").(string),
		Host:       host,
		Type:       d.Get("type").(string),
		Answer:     d.Get("value").(string),
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

func suppressHost(k, old, new string, d *schema.ResourceData) bool {
	if old == "@" && new == "" {
		return true
	}
	return false
}

func suppressPriority(k, old, new string, d *schema.ResourceData) bool {
	recordType := d.Get("type").(string)
	if recordType != "MX" && recordType != "SRV" {
		return true
	}
	return false
}
