package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKsqlDbCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKsqlDbClusterCreate,
		ReadContext:   resourceKsqlDbClusterRead,
		UpdateContext: resourceKsqlDbClusterUpdate,
		DeleteContext: resourceKsqlDbClusterDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"topic_prefix": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"kafka_api_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"kafka_api_secret": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type clusterConfigOverrides struct {
	Name string
}

type ksqlDbCluster struct {
	AccountId              string                 `json:"account_id"`
	ClusterConfigOverrides clusterConfigOverrides `json:"cluster_config_overrides"`
	Created                string                 `json:"created"`
	Deployment             string                 `json:"deployment"`
	Endpoint               string                 `json:"endpoint"`
	Id                     string                 `json:"id"`
	Image                  string                 `json:"image"`
	IsPaused               bool                   `json:"is_paused"`
	KafkaApiKey            apiKey                 `json:"kafka_api_key"`
	KafkaClusterId         string                 `json:"kafka_cluster_id"`
	KafkaUserId            int                    `json:"kafka_user_id"`
	KafkaUserName          string                 `json:"kafka_user_name"`
	Modified               string                 `json:"modified"`
	Name                   string                 `json:"name"`
	OrgResourceId          string                 `json:"org_resource_id"`
	OrganizationId         int                    `json:"organization_id"`
	OutputToicPrefix       string                 `json:"output_topic_prefix"`
	PhysicalClusterId      string                 `json:"physical_cluster_id"`
	Servers                int                    `json:"servers"`
	ServiceAccountId       int                    `jsonb:"service_account_id"`
	Status                 string                 `json:"status"`
	Storage                int                    `json:"storage"`
	TotalNumCsu            int                    `json:"total_num_csu"`
}

func resourceKsqlDbClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	type simpleApiKey struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
	}

	type requestConfig struct {
		AccountId      string       `json:"accountId"`
		KafkaApiKey    simpleApiKey `json:"kafkaApiKey"`
		KafkaClusterId string       `json:"kafkaClusterId"`
		Name           string       `json:"name"`
		TotalNumCsu    int          `json:"totalNumCsu"`
	}

	type request struct {
		Config requestConfig `json:"config"`
	}

	type response struct {
		Cluster          ksqlDbCluster
		Credentials      string
		Error            string
		ValidationErrors string
	}

	c := m.(*Client)

	environment_id := d.Get("environment_id").(string)

	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/ksqls?account_id=%s", "https://api.confluent.cloud", environment_id)

	post_data := request{
		requestConfig{
			AccountId:      environment_id,
			KafkaApiKey:    simpleApiKey{Key: d.Get("kafka_api_key").(string), Secret: d.Get("kafka_api_secret").(string)},
			KafkaClusterId: d.Get("kafka_id").(string),
			Name:           d.Get("name").(string),
			TotalNumCsu:    4,
		},
	}

	json_post_data, err := json.Marshal(post_data)
	requestBody := bytes.NewBuffer(json_post_data)

	log.Printf("key: %s secret %s", d.Get("kafka_api_key").(string), d.Get("kafka_api_secret").(string))
	log.Printf("req-body: %v", requestBody)

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
		log.Printf("body: %v", string(body))
		return append(diags, diag.Errorf("HTTP request error. Response code: %d", r.StatusCode)...)
	}
	var resp response

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

func resourceKsqlDbClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceKsqlDbClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceKsqlDbClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
