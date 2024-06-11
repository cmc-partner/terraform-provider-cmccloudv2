package cmccloudv2

import (
	"fmt"
	"strings"
	"time"

	"github.com/cmc-cloud/gocmcapiv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceRedisConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceRedisConfigurationCreate,
		Read:   resourceRedisConfigurationRead,
		Update: resourceRedisConfigurationUpdate,
		Delete: resourceRedisConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRedisConfigurationImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        redisconfigurationSchema(),
	}
}

func resourceRedisConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	datastores, _ := client.RedisInstance.ListDatastore(map[string]string{})
	database_version := d.Get("database_version").(string)
	database_type := d.Get("database_type").(string)

	datastoreModeId := ""
	cacheEngine := ""
	for _, datastore := range datastores {
		for _, version := range datastore.VersionInfos {
			if database_version == datastore.Name+" "+version.VersionName {
				cacheEngine = version.ID
				for _, mode := range version.ModeInfo {
					if strings.Contains(mode.Name, database_type) {
						datastoreModeId = mode.ID
					}
				}
			}
		}
	}

	if datastoreModeId == "" {
		return fmt.Errorf("Not found database_version")
	}
	if cacheEngine == "" {
		return fmt.Errorf("Not found database_type")
	}
	configuration, err := client.RedisConfiguration.Create(map[string]interface{}{
		"name":            d.Get("name").(string),
		"description":     d.Get("description").(string),
		"datastoreModeId": datastoreModeId,
		"cacheEngine":     cacheEngine,
		"overridesConfig": d.Get("parameters").(map[string]interface{}),
	})
	if err != nil {
		return fmt.Errorf("Error creating Redis Configuration: %s", err)
	}
	d.SetId(configuration.ID)
	return resourceRedisConfigurationRead(d, meta)
}

func resourceRedisConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	configuration, err := client.RedisConfiguration.Get(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Redis Configuration %s: %v", d.Id(), err)
	}

	_ = d.Set("id", configuration.ID)
	_ = d.Set("name", configuration.Name)
	_ = d.Set("description", configuration.Description)
	_ = d.Set("datastore_type", configuration.DatastoreName+" "+configuration.DatastoreVersion)
	_ = d.Set("datastore_version", configuration.DatastoreMode)
	_ = d.Set("parameters", convertRedisConfigurationParameters(configuration.Parameters))
	return nil
}

func convertRedisConfigurationParameters(obj []gocmcapiv2.RedisConfigurationParameter) map[string]interface{} {
	result := map[string]interface{}{}
	index := 0
	for _, param := range obj {
		result[param.Name] = param.Value
		index++
	}
	return result
}
func resourceRedisConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	id := d.Id()
	if d.HasChange("name") || d.HasChange("description") {
		_, err := client.RedisConfiguration.Update(id, map[string]interface{}{
			"name":        d.Get("name").(string),
			"description": d.Get("description").(string),
		})
		if err != nil {
			return fmt.Errorf("Error when update info of Redis Configuration [%s]: %v", id, err)
		}
	}
	if d.HasChange("parameters") {
		// parameters := convertRedisParametersJsonString(d.Get("parameters").(map[string]interface{}))
		_, err := client.RedisConfiguration.UpdateParameters(id, d.Get("parameters").(map[string]interface{}))
		if err != nil {
			return fmt.Errorf("Error when update parameters of Redis Configuration [%s]: %v", id, err)
		}
	}
	return resourceRedisConfigurationRead(d, meta)
}

// func convertRedisParametersJsonString(params map[string]interface{}) string {
// 	results := make(map[string]interface{})
// 	for _, param := range params.List() {
// 		_param := param.(map[string]interface{})
// 		key := _param["key"].(string)
// 		val := _param["value"]

//			if isStringType, ok := _param["string_type"].(bool); !ok || !isStringType {
//				// check if value can be converted into int
//				if valueInt, err := strconv.Atoi(val.(string)); err == nil {
//					val = valueInt
//					// check if value can be converted into bool
//				} else if valueBool, err := strconv.ParseBool(val.(string)); err == nil {
//					val = valueBool
//				}
//			}
//			results[key] = val
//		}
//		// return results
//		jsonData, err := json.Marshal(results)
//		if err != nil {
//			fmt.Errorf("Error converting map to JSON: %s", err)
//		}
//		return string(jsonData)
//	}
func resourceRedisConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	_, err := client.RedisConfiguration.Delete(d.Id())

	if err != nil {
		return fmt.Errorf("Error delete database configuration: %v", err)
	}
	_, err = waitUntilRedisConfigurationDeleted(d, meta)
	if err != nil {
		return fmt.Errorf("Error delete database configuration: %v", err)
	}
	return nil
}

func resourceRedisConfigurationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceRedisConfigurationRead(d, meta)
	return []*schema.ResourceData{d}, err
}

func waitUntilRedisConfigurationDeleted(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	return waitUntilResourceDeleted(d, meta, WaitConf{
		Delay:      10 * time.Second,
		MinTimeout: 30 * time.Second,
	}, func(id string) (any, error) {
		return getClient(meta).RedisConfiguration.Get(id)
	})
}
