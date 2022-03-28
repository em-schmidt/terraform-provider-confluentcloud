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

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiKeyCreate,
		ReadContext:   resourceApiKeyRead,
		DeleteContext: resourceApiKeyDelete,
		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"owner_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_email": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"resource_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type clientLogicalCluster struct {
	Name string
}

type logicalCluster struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type apiKey struct {
	AccountId              string                 `json:"account_id"`
	ClientLogicalClusters  []clientLogicalCluster `json:"client_logical_clusters"`
	Created                string                 `json:"created"`
	Deactivate             bool                   `json:"deactivated"`
	Description            string                 `json:"description"`
	HashFunciton           string                 `json:"hash_function"`
	HashedSecret           string                 `json:"hashed_secret"`
	Id                     int                    `json:"id"`
	Internal               string                 `json:"internal"`
	Key                    string                 `json:"key"`
	LogicalClusters        []logicalCluster       `json:"logical_clusters"`
	Modified               string                 `json:"modified"`
	OrganizationId         int                    `json:"organization_id"`
	OrganizationResourceId string                 `json:"organization_resource_id"`
	SaslMechanism          string                 `json:"sasl_mechanism"`
	Secret                 string                 `json:"secret"`
	ServiceAccount         bool                   `json:"service_account"`
	StoreSecret            bool                   `json:"store_secret"`
	UserId                 int                    `json:"user_id"`
	UserResourceId         string                 `json:"user_resource_id"`
}

type response struct {
	ApiKey apiKey `json:"api_key"`
	Error  string
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	type apiKeyRequest struct {
		AccountId       string           `json:"accountId"`
		Description     string           `json:"description"`
		LogicalClusters []logicalCluster `json:"logicalClusters"`
		UserId          int              `json:"userId"`
		UserResourceId  string           `json:"userResourceId"`
	}

	type request struct {
		ApiKey apiKeyRequest `json:"apiKey"`
	}

	c := m.(*Client)
	environmentId := d.Get("environment_id").(string)
	resourceId := d.Get("resource_id").(string)
	resourceType := d.Get("resource_type").(string)
	description := d.Get("description").(string)

	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/api_keys?account_id=%s", "https://api.confluent.cloud", environmentId)

	post_data := request{
		apiKeyRequest{
			AccountId:       environmentId,
			Description:     description,
			LogicalClusters: []logicalCluster{{Id: resourceId, Type: resourceType}},
			UserId:          0,
			UserResourceId:  "sa-gq8no3",
		},
	}

	json_post_data, err := json.Marshal(post_data)
	if err != nil {
		return diag.FromErr(err)
	}
	requestBody := bytes.NewBuffer(json_post_data)

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
		log.Printf("body: %v", body)
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

	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("resp-body-raw: %v", body)

	log.Printf("resp-body: %v", resp)

	d.Set("secret", resp.ApiKey.Secret)
	d.Set("key", resp.ApiKey.Key)
	d.SetId(resp.ApiKey.Key)

	return diags
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceApiKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
