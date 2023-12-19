package database

import (
	"context"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/db"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/exterror"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
	"strconv"
	"time"
)

func GetPackages(ctx context.Context, allFlag bool) (result []*models.PluginPackages, err error) {
	result = []*models.PluginPackages{}
	if allFlag {
		err = db.MysqlEngine.SQL("select * from plugin_packages").Find(&result)
	} else {
		err = db.MysqlEngine.SQL("select * from plugin_packages where status in (0,1)").Find(&result)
	}
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	return
}

func GetPluginDependencies(ctx context.Context, pluginPackageId string) (result []*models.PluginPackageDepObj, err error) {
	var dependRows []*models.PluginPackageDependencies
	err = db.MysqlEngine.SQL("select dependency_package_name,dependency_package_version from plugin_package_dependencies where plugin_package_id=?", pluginPackageId).Find(&dependRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	result = []*models.PluginPackageDepObj{}
	for _, v := range dependRows {
		result = append(result, &models.PluginPackageDepObj{PackageName: v.DependencyPackageName, Version: v.DependencyPackageVersion})
	}
	return
}

func GetPluginMenus(ctx context.Context, pluginPackageId string) (result []*models.PluginPackageMenus, err error) {
	result = []*models.PluginPackageMenus{}
	err = db.MysqlEngine.SQL("select id,plugin_package_id,code,display_name,local_display_name,menu_order from plugin_package_menus where plugin_package_id=?", pluginPackageId).Find(&result)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	for _, v := range result {
		v.Source = v.PluginPackageId
	}
	return
}

func GetPluginSystemParameters(ctx context.Context, pluginPackageId string) (result []*models.SystemVariables, err error) {
	pluginPackageObj, tmpErr := getPluginPackageObj(ctx, pluginPackageId)
	if tmpErr != nil {
		err = tmpErr
		return
	}
	packageName := fmt.Sprintf("%s__%s", pluginPackageObj.Name, pluginPackageObj.Version)
	result = []*models.SystemVariables{}
	err = db.MysqlEngine.SQL("select * from system_variables where source=?", packageName).Find(&result)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
	}
	return
}

func getPluginPackageObj(ctx context.Context, pluginPackageId string) (pluginPackageObj *models.PluginPackages, err error) {
	var pluginPackageRows []*models.PluginPackages
	err = db.MysqlEngine.SQL("select id,name,`version`,status,ui_package_included from plugin_packages where id=?", pluginPackageId).Find(&pluginPackageRows)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
	} else {
		if len(pluginPackageRows) == 0 {
			err = fmt.Errorf("can not find pluginPackages with id:%s ", pluginPackageId)
		} else {
			pluginPackageObj = pluginPackageRows[0]
		}
	}
	return
}

func GetPluginAuthorities(ctx context.Context, pluginPackageId string) (result []*models.PluginPackageAuthorities, err error) {
	result = []*models.PluginPackageAuthorities{}
	err = db.MysqlEngine.SQL("select * from plugin_package_authorities where plugin_package_id=?", pluginPackageId).Find(&pluginPackageId)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
	}
	return
}

func GetPluginRuntimeResources(ctx context.Context, pluginPackageId string) (result *models.PluginRuntimeResourceData, err error) {
	result = &models.PluginRuntimeResourceData{Docker: []*models.PluginPackageRuntimeResourcesDocker{}, Mysql: []*models.PluginPackageRuntimeResourcesMysql{}, S3: []*models.PluginPackageRuntimeResourcesS3{}}
	err = db.MysqlEngine.SQL("select * from plugin_package_runtime_resources_docker where plugin_package_id=?", pluginPackageId).Find(&result.Docker)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	err = db.MysqlEngine.SQL("select * from plugin_package_runtime_resources_mysql where plugin_package_id=?", pluginPackageId).Find(&result.Mysql)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	err = db.MysqlEngine.SQL("select * from plugin_package_runtime_resources_s3 where plugin_package_id=?", pluginPackageId).Find(&result.S3)
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
		return
	}
	return
}

func UploadPackage(registerConfig *models.RegisterXML, withUi, enterprise bool) (err error) {
	var actions []*db.ExecAction
	pluginPackageId := "plugin_" + guid.CreateGuid()
	nowTime := time.Now()
	edition := models.PluginEditionCommunity
	if enterprise {
		edition = models.PluginEditionEnterprise
	}
	actions = append(actions, &db.ExecAction{Sql: "insert into plugin_packages ( id,name,`version`,status,upload_timestamp,ui_package_included,`edition` ) values (?,?,?,?,?,?,?)", Param: []interface{}{
		pluginPackageId, registerConfig.Name, registerConfig.Version, models.PluginStatusUnRegistered, nowTime, withUi, edition}})
	for _, pluginConfig := range registerConfig.Plugins.Plugin {
		pluginConfigId := "p_config_" + guid.CreateGuid()
		actions = append(actions, &db.ExecAction{Sql: "insert into plugin_configs (id,plugin_package_id,name,target_package,target_entity,target_entity_filter_rule,register_name) values (?,?,?,?,?,?,?)", Param: []interface{}{
			pluginConfigId, pluginPackageId, pluginConfig.Name, pluginConfig.TargetPackage, pluginConfig.TargetEntity, pluginConfig.TargetEntityFilterRule, pluginConfig.RegisterName,
		}})
		for _, pluginConfigInterface := range pluginConfig.Interface {
			pluginConfigInterfaceId := "p_conf_inf_" + guid.CreateGuid()
			serviceName := fmt.Sprintf("%s/%s/%s", registerConfig.Name, pluginConfig.Name, pluginConfigInterface.Action)
			httpMethod := pluginConfigInterface.HttpMethod
			if httpMethod == "" {
				httpMethod = "POST"
			}
			isAsyncProcessing := false
			if pluginConfigInterface.IsAsyncProcessing == "Y" {
				isAsyncProcessing = true
			}
			actions = append(actions, &db.ExecAction{Sql: "insert into plugin_config_interfaces (id,plugin_config_id,action,service_name,service_display_name,path,http_method,is_async_processing,type,filter_rule,description) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				pluginConfigInterfaceId, pluginConfigId, pluginConfigInterface.Action, serviceName, serviceName, pluginConfigInterface.Path, httpMethod, isAsyncProcessing, pluginConfigInterface.Type, pluginConfigInterface.FilterRule, pluginConfigInterface.Description,
			}})
			for _, interfaceParam := range pluginConfigInterface.InputParameters.Parameter {
				interfaceParamId := "p_conf_inf_param_" + guid.CreateGuid()
				required, sensitive, multiple := false, false, false
				if interfaceParam.Required == "Y" {
					required = true
				}
				if interfaceParam.SensitiveData == "Y" {
					sensitive = true
				}
				if interfaceParam.Multiple == "Y" {
					multiple = true
				}
				actions = append(actions, &db.ExecAction{Sql: "insert into plugin_config_interface_parameters (id,plugin_config_interface_id,type,name,data_type,mapping_type,mapping_entity_expression,mapping_system_variable_name,required,sensitive_data,description,mapping_val,ref_object_name,multiple ) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
					interfaceParamId, pluginConfigInterfaceId, models.PluginParamTypeInput, interfaceParam.Text, interfaceParam.Datatype, interfaceParam.MappingType, interfaceParam.MappingEntityExpression, interfaceParam.MappingSystemVariableName, required, sensitive, interfaceParam.Description, interfaceParam.MappingVal, interfaceParam.RefObjectName, multiple,
				}})
			}
			for _, interfaceParam := range pluginConfigInterface.OutputParameters.Parameter {
				interfaceParamId := "p_conf_inf_param_" + guid.CreateGuid()
				required, sensitive, multiple := false, false, false
				if interfaceParam.Required == "Y" {
					required = true
				}
				if interfaceParam.SensitiveData == "Y" {
					sensitive = true
				}
				if interfaceParam.Multiple == "Y" {
					multiple = true
				}
				actions = append(actions, &db.ExecAction{Sql: "insert into plugin_config_interface_parameters (id,plugin_config_interface_id,type,name,data_type,mapping_type,mapping_entity_expression,mapping_system_variable_name,required,sensitive_data,description,mapping_val,ref_object_name,multiple ) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
					interfaceParamId, pluginConfigInterfaceId, models.PluginParamTypeOutput, interfaceParam.Text, interfaceParam.Datatype, interfaceParam.MappingType, interfaceParam.MappingEntityExpression, interfaceParam.MappingSystemVariableName, required, sensitive, interfaceParam.Description, interfaceParam.MappingVal, interfaceParam.RefObjectName, multiple,
				}})
			}
		}
	}
	variableSource := fmt.Sprintf("%s__%s", registerConfig.Name, registerConfig.Version)
	for _, systemVariable := range registerConfig.SystemParameters.SystemParameter {
		svId := "sys_var_" + guid.CreateGuid()
		actions = append(actions, &db.ExecAction{Sql: "INSERT INTO system_variables (id,package_name,name,value,default_value,`scope`,source,status) VALUES (?,?,?,?,?,?,?,?)", Param: []interface{}{
			svId, registerConfig.Name, systemVariable.Name, "", systemVariable.DefaultValue, systemVariable.ScopeType, variableSource, models.SystemVariableInactive,
		}})
	}
	for _, dependence := range registerConfig.PackageDependencies.PackageDependency {
		depId := "p_pkg_dep_" + guid.CreateGuid()
		actions = append(actions, &db.ExecAction{Sql: "insert into plugin_package_dependencies (id,plugin_package_id,dependency_package_name,dependency_package_version) values (?,?,?,?)", Param: []interface{}{
			depId, pluginPackageId, dependence.Name, dependence.Version,
		}})
	}
	actions = append(actions, &db.ExecAction{Sql: "insert into plugin_package_runtime_resources_docker (id,plugin_package_id,image_name,container_name,port_bindings,volume_bindings,env_variables) values (?,?,?,?,?,?,?)", Param: []interface{}{
		"p_res_docker_" + guid.CreateGuid(), pluginPackageId, registerConfig.ResourceDependencies.Docker.ImageName, registerConfig.ResourceDependencies.Docker.ContainerName, registerConfig.ResourceDependencies.Docker.PortBindings, registerConfig.ResourceDependencies.Docker.VolumeBindings, registerConfig.ResourceDependencies.Docker.EnvVariables,
	}})
	if len(registerConfig.DataModel.Entity) > 0 {
		maxVersion, getVersionErr := getMaxDataModelVersion(registerConfig.Name)
		if getVersionErr != nil {
			err = getVersionErr
			return
		}
		maxVersion = maxVersion + 1
		for _, dataModel := range registerConfig.DataModel.Entity {
			dmId := "p_model_" + guid.CreateGuid()
			actions = append(actions, &db.ExecAction{Sql: "INSERT INTO plugin_package_data_model (id,`version`,package_name,is_dynamic,update_path,update_method,update_source,update_time) VALUES (?,?,?,?,?,?,?,?)", Param: []interface{}{
				dmId, maxVersion, registerConfig.Name, 0, "/data-model", "GET", "PLUGIN_PACKAGE", nowTime,
			}})
			entityId := "p_mod_entity_" + guid.CreateGuid()
			actions = append(actions, &db.ExecAction{Sql: "INSERT INTO plugin_package_entities (id,data_model_id,data_model_version,package_name,name,display_name,description) VALUES (?,?,?,?,?,?,?)", Param: []interface{}{
				entityId, dmId, maxVersion, registerConfig.Name, dataModel.Name, dataModel.DisplayName, dataModel.Description,
			}})
			for attrIndex, attr := range dataModel.Attribute {
				attrId := "p_mod_attr_" + guid.CreateGuid()
				actions = append(actions, &db.ExecAction{Sql: "INSERT INTO plugin_package_attributes (id,entity_id,name,description,data_type,mandatory,multiple,created_time,order_no) values  (?,?,?,?,?,?,?,?,?)", Param: []interface{}{
					attrId, entityId, attr.Name, attr.Description, attr.Datatype, 0, 0, nowTime, attrIndex,
				}})
			}
		}
	}
	err = db.Transaction(actions, db.NewDBCtx(fmt.Sprintf("upload_%s_%s", registerConfig.Name, registerConfig.Version)))
	return
}

func getMaxDataModelVersion(packageName string) (maxV int, err error) {
	queryResult, queryErr := db.MysqlEngine.QueryString("SELECT max(`version`) as ver FROM plugin_package_data_model WHERE package_name =? GROUP BY package_name")
	if queryErr != nil {
		err = fmt.Errorf("query data model max version fail,%s ", queryErr.Error())
		return
	}
	maxV, _ = strconv.Atoi(queryResult[0]["ver"])
	return
}

func GetSimplePluginPackage(param *models.PluginPackages, emptyCheck bool) (err error) {
	var pluginPackagesRows []*models.PluginPackages
	if param.Id != "" {
		err = db.MysqlEngine.SQL("select * from plugin_packages where id=?", param.Id).Find(&pluginPackagesRows)
	} else {
		if param.Edition == "" {
			err = db.MysqlEngine.SQL("select * from plugin_packages where name=? and `version`=?", param.Name, param.Version).Find(&pluginPackagesRows)
		} else {
			err = db.MysqlEngine.SQL("select * from plugin_packages where name=? and `version`=? and `edition`=?", param.Name, param.Version, param.Edition).Find(&pluginPackagesRows)
		}
	}
	if err != nil {
		err = exterror.Catch(exterror.New().DatabaseQueryError, err)
	} else {
		if len(pluginPackagesRows) == 0 {
			if emptyCheck {
				err = fmt.Errorf("can not find plugin packages with id:%s name:%s version:%s", param.Id, param.Name, param.Version)
			}
		} else {
			param = pluginPackagesRows[0]
		}
	}
	return
}

func CheckPluginPackageDependence() {

}
