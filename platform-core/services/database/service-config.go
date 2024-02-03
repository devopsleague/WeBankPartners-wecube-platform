package database

import (
	"context"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/db"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/exterror"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
)

func GetPluginConfigs(ctx context.Context, pluginPackageId string, roles []string) (result []*models.PluginConfigQueryObj, err error) {
	var pluginConfigRows []*models.PluginConfigs
	err = db.MysqlEngine.Context(ctx).SQL("select * from plugin_configs where plugin_package_id = ? order by name", pluginPackageId).Find(&pluginConfigRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	result = []*models.PluginConfigQueryObj{}
	if len(pluginConfigRows) == 0 {
		return
	}
	var pluginConfigIds []string
	for _, v := range pluginConfigRows {
		pluginConfigIds = append(pluginConfigIds, v.Id)
	}
	pluginConfigFilter, pluginConfigParams := db.CreateListParams(pluginConfigIds, "")
	var pluginConfigRolesRows []*models.PluginConfigRoles
	err = db.MysqlEngine.SQL("select plugin_cfg_id,perm_type,role_id,role_name from plugin_config_roles where plugin_cfg_id in ("+pluginConfigFilter+")", pluginConfigParams...).Find(&pluginConfigRolesRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	permissionMap := make(map[string]*models.PermissionRoleData)
	for _, row := range pluginConfigRolesRows {
		if v, b := permissionMap[row.PluginCfgId]; b {
			if row.PermType == "MGMT" {
				v.MGMT = append(v.MGMT, row.RoleName)
			} else {
				v.USE = append(v.USE, row.RoleName)
			}
		} else {
			newPermObj := models.PermissionRoleData{MGMT: []string{}, USE: []string{}}
			if row.PermType == "MGMT" {
				newPermObj.MGMT = append(newPermObj.MGMT, row.RoleName)
			} else {
				newPermObj.USE = append(newPermObj.USE, row.RoleName)
			}
			permissionMap[row.PluginCfgId] = &newPermObj
		}
	}
	for k, v := range permissionMap {
		matchFlag := false
		for _, legalRole := range append(v.MGMT, v.USE...) {
			for _, userRole := range roles {
				if userRole == legalRole {
				}
				matchFlag = true
				break
			}
			if matchFlag {
				break
			}
		}
		if !matchFlag {
			delete(permissionMap, k)
		}
	}
	pluginConfigNameMapIndex := make(map[string]int)
	for _, row := range pluginConfigRows {
		if permObj, ok := permissionMap[row.Id]; ok {
			tmpObj := models.PluginConfigDto{PluginConfigs: *row, PermissionToRole: permObj}
			if nameIndex, existFlag := pluginConfigNameMapIndex[row.Name]; existFlag {
				result[nameIndex].PluginConfigDtoList = append(result[nameIndex].PluginConfigDtoList, &tmpObj)
			} else {
				pluginConfigNameMapIndex[row.Name] = len(result)
				result = append(result, &models.PluginConfigQueryObj{PluginConfigName: row.Name, PluginConfigDtoList: []*models.PluginConfigDto{&tmpObj}})
			}
		}
	}
	return
}

func GetConfigInterfaces(ctx context.Context, pluginPackageId string) (result []*models.PluginInterfaceQueryObj, err error) {
	var interfaceRows []*models.PluginConfigInterfaces
	err = db.MysqlEngine.Context(ctx).SQL("select * from plugin_config_interfaces where plugin_config_id=?", pluginPackageId).Find(&interfaceRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	result = []*models.PluginInterfaceQueryObj{}
	if len(interfaceRows) == 0 {
		return
	}
	var interfaceIds []string
	for _, v := range interfaceRows {
		interfaceIds = append(interfaceIds, v.Id)
		result = append(result, &models.PluginInterfaceQueryObj{PluginConfigInterfaces: *v, InputParameters: []*models.PluginConfigInterfaceParameters{}, OutputParameters: []*models.PluginConfigInterfaceParameters{}})
	}
	interfaceFilter, interfaceParams := db.CreateListParams(interfaceIds, "")
	var interfaceParamRows []*models.PluginConfigInterfaceParameters
	err = db.MysqlEngine.Context(ctx).SQL("select * from plugin_config_interface_parameters where plugin_config_interface_id in ("+interfaceFilter+") order by name", interfaceParams...).Find(&interfaceParamRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	inputMap := make(map[string][]*models.PluginConfigInterfaceParameters)
	outputMap := make(map[string][]*models.PluginConfigInterfaceParameters)
	for _, row := range interfaceParamRows {
		if row.Type == "INPUT" {
			if v, b := inputMap[row.PluginConfigInterfaceId]; b {
				inputMap[row.PluginConfigInterfaceId] = append(v, row)
			} else {
				inputMap[row.PluginConfigInterfaceId] = []*models.PluginConfigInterfaceParameters{row}
			}
		} else {
			if v, b := outputMap[row.PluginConfigInterfaceId]; b {
				outputMap[row.PluginConfigInterfaceId] = append(v, row)
			} else {
				outputMap[row.PluginConfigInterfaceId] = []*models.PluginConfigInterfaceParameters{row}
			}
		}
	}
	for _, v := range result {
		if inputParams, b := inputMap[v.Id]; b {
			v.InputParameters = inputParams
		}
		if outputParams, b := outputMap[v.Id]; b {
			v.OutputParameters = outputParams
		}
	}
	return
}

func GetConfigInterfacesById(ctx context.Context, id string) (result *models.PluginConfigInterfaces, err error) {
	var list []*models.PluginConfigInterfaces
	err = db.MysqlEngine.Context(ctx).SQL("select id, plugin_config_id, action, service_name,service_display_name, path, http_method,"+
		"is_async_processing,type, filter_rule,description from  plugin_config_interfaces where id= ?", id).Find(&list)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	if len(list) > 0 {
		result = list[0]
	}
	return
}

func GetPluginConfigById(ctx context.Context, id string) (result *models.PluginConfigs, err error) {
	var list []*models.PluginConfigs
	err = db.MysqlEngine.Context(ctx).SQL("select id,plugin_package_id, name, target_package,target_entity,"+
		"target_entity_filter_rule,register_name,status from plugin_configs where id=?", id).Find(&list)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	if len(list) > 0 {
		result = list[0]
	}
	return
}

func GetPluginPackageById(ctx context.Context, id string) (result *models.PluginPackages, err error) {
	var list []*models.PluginPackages
	err = db.MysqlEngine.Context(ctx).SQL("select id,name,version,status,upload_timestamp,ui_package_included,edition from plugin_packages where id=?", id).Find(&list)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	if len(list) > 0 {
		result = list[0]
	}
	return
}

func GetPluginConfigInterfaceParameters(ctx context.Context, pluginConfigInterfaceId, parameterType string) (list []*models.PluginConfigInterfaceParameters, err error) {
	err = db.MysqlEngine.Context(ctx).SQL("select id,plugin_config_interface_id, type, name,data_type,mapping_type,mapping_entity_expression,"+
		"mapping_system_variable_name,required,sensitive_data,description,mapping_val,ref_object_name,multiple from plugin_config_interface_parameters"+
		" where plugin_config_interface_id = ? and type =?", pluginConfigInterfaceId, parameterType).Find(&list)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	return
}