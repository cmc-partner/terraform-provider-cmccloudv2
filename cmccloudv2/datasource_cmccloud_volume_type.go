package cmccloudv2

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cmc-cloud/gocmcapiv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVolumeTypeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"volume_type_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Id of the volumetype",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Id of the server that volumetype attached (case-insensitive), match exactly",
		},
		"multi_attach": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Filter by description that match exactly this text (case-insensitive), description is the text display on the portal",
			Optional:    true,
			ForceNew:    true,
		},
	}
}

func datasourceVolumeType() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceVolumeTypeECRead,
		Schema: datasourceVolumeTypeSchema(),
	}
}

func datasourceVolumeTypeDatabase() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceVolumeTypeDatabaseRead,
		Schema: datasourceVolumeTypeSchema(),
	}
}

func dataSourceVolumeTypeECRead(d *schema.ResourceData, meta interface{}) error {
	return dataSourceVolumeTypeRead(d, meta, false)
}

func dataSourceVolumeTypeDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	return dataSourceVolumeTypeRead(d, meta, true)
}

func dataSourceVolumeTypeRead(d *schema.ResourceData, meta interface{}, forDatabase bool) error {
	client := meta.(*CombinedConfig).goCMCClient()

	var allVolumeTypes []gocmcapiv2.VolumeType
	if volumeTypeId := d.Get("volume_type_id").(string); volumeTypeId != "" {
		volumetype, err := client.VolumeType.Get(volumeTypeId)
		if err != nil {
			if errors.Is(err, gocmcapiv2.ErrNotFound) {
				d.SetId("")
				return fmt.Errorf("unable to retrieve volume type [%s]: %s", volumeTypeId, err)
			}
		}
		allVolumeTypes = append(allVolumeTypes, volumetype)
	} else {
		params := map[string]string{}
		volumetypes, err := client.VolumeType.List(params)
		if err != nil {
			return fmt.Errorf("error when get volume types %v", err)
		}
		allVolumeTypes = append(allVolumeTypes, volumetypes...)
	}
	if len(allVolumeTypes) > 0 {
		var filteredVolumeTypes []gocmcapiv2.VolumeType
		for _, volumetype := range allVolumeTypes {
			if v := d.Get("description").(string); v != "" {
				if strings.ToLower(volumetype.Description) != strings.ToLower(v) {
					continue
				}
			}
			isMultiAttach := d.Get("multi_attach").(bool)
			// volume type cho database => ko hien thi cac loai khac
			if forDatabase && !strings.Contains(volumetype.Name, "database") {
				continue
			}
			// volume type cho ec => ko hien thi volume type cho database
			if !forDatabase && strings.Contains(volumetype.Name, "database") {
				continue
			}
			// neu la multi attach => filter nhung loai co name chua attach
			if isMultiAttach && !strings.Contains(volumetype.Name, "attach") {
				continue
			}
			// neu la ko multi attach => bo di nhung loai co name chua attach
			if !isMultiAttach && strings.Contains(volumetype.Name, "attach") {
				continue
			}
			filteredVolumeTypes = append(filteredVolumeTypes, volumetype)
		}
		allVolumeTypes = filteredVolumeTypes
	}
	if len(allVolumeTypes) < 1 {
		return fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}

	if len(allVolumeTypes) > 1 {
		gocmcapiv2.Logo("[DEBUG] Multiple results found: %#v", allVolumeTypes)
		return fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}

	return dataSourceComputeVolumeTypeAttributes(d, allVolumeTypes[0])
}

func dataSourceComputeVolumeTypeAttributes(d *schema.ResourceData, volumetype gocmcapiv2.VolumeType) error {
	log.Printf("[DEBUG] Retrieved volumetype %s: %#v", volumetype.ID, volumetype)
	d.SetId(volumetype.ID)
	return errors.Join(
		d.Set("name", volumetype.Name),
		d.Set("description", volumetype.Description),
	)
}
