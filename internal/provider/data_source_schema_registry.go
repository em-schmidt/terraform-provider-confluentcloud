package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSchemaRegistryRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceSchemaRegistryRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {

	type schemaRegistryCluster struct {
		Id                  string
		Name                string
		Kafka_cluster_id    string
		Endpoint            string
		Created             string
		Modified            string
		Status              string
		Physical_cluster_id string
		Account_id          string
		Organization_id     int
		Max_schemas         int
		Org_resource_id     string
	}

	type responseBody struct {
		Error    string
		Clusters []schemaRegistryCluster
	}

	c := m.(*Client)

	environment_id := d.Get("environment_id").(string)

	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/schema_registries?account_id=%s", "https://api.confluent.cloud", environment_id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("req: %v", req)
	log.Printf("client: %v", client)

	req.SetBasicAuth(c.apiKey, c.apiSecret)
	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return append(diags, diag.Errorf("HTTP request error. Response code: %d", r.StatusCode)...)
	}

	var resp responseBody

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.Error != "" {
		return append(diags, diag.Errorf("Unexpected API response. Body: %v", resp)...)
	}

	d.Set("name", resp.Clusters[0].Name)
	d.SetId(resp.Clusters[0].Id)

	return diags
}
