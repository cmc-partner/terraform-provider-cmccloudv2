package cmccloudv2

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cmc-cloud/gocmcapiv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceAutoScalingConfigurationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"autoscaling_configuration_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Id of the autoscaling configuration",
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"created_at": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
	}
}

func datasourceAutoScalingConfiguration() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceAutoScalingConfigurationRead,
		Schema: datasourceAutoScalingConfigurationSchema(),
	}
}

func dataSourceAutoScalingConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()

	var allAutoScalingConfigurations []gocmcapiv2.AutoScalingConfiguration
	if autoscalingConfigurationId := d.Get("autoscaling_configuration_id").(string); autoscalingConfigurationId != "" {
		autoscalingConfiguration, err := client.AutoScalingConfiguration.Get(autoscalingConfigurationId)
		if err != nil {
			if errors.Is(err, gocmcapiv2.ErrNotFound) {
				d.SetId("")
				return fmt.Errorf("unable to retrieve autoscaling configuration [%s]: %s", autoscalingConfigurationId, err)
			}
		}
		allAutoScalingConfigurations = append(allAutoScalingConfigurations, autoscalingConfiguration)
	} else {
		params := map[string]string{
			"name": d.Get("name").(string),
		}
		autoscalingconfigurations, err := client.AutoScalingConfiguration.List(params)
		if err != nil {
			return fmt.Errorf("error when get autoscaling configuration %v", err)
		}
		allAutoScalingConfigurations = append(allAutoScalingConfigurations, autoscalingconfigurations...)
	}
	if len(allAutoScalingConfigurations) > 0 {
		var filteredAutoScalingConfigurations []gocmcapiv2.AutoScalingConfiguration
		for _, autoscalingConfiguration := range allAutoScalingConfigurations {
			if v := d.Get("name").(string); v != "" {
				if strings.ToLower(autoscalingConfiguration.Name) != strings.ToLower(v) {
					continue
				}
			}
			filteredAutoScalingConfigurations = append(filteredAutoScalingConfigurations, autoscalingConfiguration)
		}
		allAutoScalingConfigurations = filteredAutoScalingConfigurations
	}
	if len(allAutoScalingConfigurations) < 1 {
		return fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}

	if len(allAutoScalingConfigurations) > 1 {
		gocmcapiv2.Logo("[DEBUG] Multiple results found: %#v", allAutoScalingConfigurations)
		return fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}

	return dataSourceComputeAutoScalingConfigurationAttributes(d, allAutoScalingConfigurations[0])
}

func dataSourceComputeAutoScalingConfigurationAttributes(d *schema.ResourceData, autoscalingConfiguration gocmcapiv2.AutoScalingConfiguration) error {
	log.Printf("[DEBUG] Retrieved autoscaling configuration %s: %#v", autoscalingConfiguration.ID, autoscalingConfiguration)
	d.SetId(autoscalingConfiguration.ID)
	return errors.Join(
		d.Set("name", autoscalingConfiguration.Name),
		d.Set("created_at", autoscalingConfiguration.CreatedAt),
	)
}
