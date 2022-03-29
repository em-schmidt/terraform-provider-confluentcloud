package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSchemaRegistryCreate,
		ReadContext:   resourceSchemaRegistryRead,
		UpdateContext: resourceSchemaRegistryUpdate,
		DeleteContext: resourceSchemaRegistryDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"kafka_cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_provider": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"location": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schemaRegistryImport,
		},
	}
}

/*
				"account_id": "env-18vqv",
        "created": "2022-03-28T00:35:19.860568Z",
        "endpoint": "https://psrc-1wydj.us-east-2.aws.confluent.cloud",
        "id": "lsrc-o338zj",
        "kafka_cluster_id": "lkc-415jz",
        "max_schemas": 1000,
        "modified": "2022-03-28T00:35:19.860568Z",
        "name": "account schema-registry",
        "org_resource_id": "8698e965-f21a-4fba-b440-170f630352d6",
        "organization_id": 51445,
        "physical_cluster_id": "psrc-1wydj",
        "status": "UP"
*/

func resourceSchemaRegistryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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
		Cluster          schemaRegistryCluster
		Credentials      string
		Error            string
		ValidationErrors string
	}

	type requestConfig struct {
		AccountId       string `json:"accountId"`
		Location        string `json:"location"`
		Name            string `json:"name"`
		ServiceProvider string `json:"serviceProvider"`
	}

	type request struct {
		Config requestConfig `json:"config"`
	}

	var diags diag.Diagnostics

	c := m.(*Client)

	environment_id := d.Get("environment_id").(string)
	location := d.Get("location").(string)
	service_provider := d.Get("service_provider").(string)

	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/schema_registries?account_id=%s", "https://api.confluent.cloud", environment_id)

	post_data := request{
		requestConfig{
			AccountId:       environment_id,
			Location:        location,
			Name:            "account schema-registry",
			ServiceProvider: service_provider,
		},
	}
	json_post_data, err := json.Marshal(post_data)
	requestBody := bytes.NewBuffer(json_post_data)

	req, err := http.NewRequestWithContext(ctx, "POST", url, requestBody)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.SetBasicAuth(c.apiKey, c.apiSecret)

	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("body: %v", body)
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

	d.Set("name", resp.Cluster.Name)
	d.SetId(resp.Cluster.Id)

	return diags
}

func resourceSchemaRegistryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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
		Error   string
		Cluster schemaRegistryCluster
	}

	c := m.(*Client)

	var diags diag.Diagnostics

	srId := d.Id()
	environmentId := d.Get("environment_id")

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/schema_registries/%s?account_id=%s", "https://api.confluent.cloud", srId, environmentId)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

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

	d.Set("name", resp.Cluster.Name)

	return diags
}

func resourceSchemaRegistryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}

func resourceSchemaRegistryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}

func schemaRegistryImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	envIDAndClusterID := d.Id()
	parts := strings.Split(envIDAndClusterID, "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format for kafka import: expected '<env ID>/<lkc ID>'")
	}

	environmentId := parts[0]
	clusterId := parts[1]
	d.SetId(clusterId)
	d.Set("environment_id", environmentId)
	log.Printf("[INFO] Schema Registry import for %s", clusterId)

	return []*schema.ResourceData{d}, nil
}
