package database

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/db"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/log"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
	"github.com/WeBankPartners/wecube-platform/platform-core/services/remote"
	"github.com/WeBankPartners/wecube-platform/platform-core/services/remote/monitor"
	"os"
	"sort"
	"strings"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// AnalyzeCMDBDataExport 分析cmdb数据并写入自动分析表
func AnalyzeCMDBDataExport(ctx context.Context, param *models.AnalyzeDataTransParam) (actions []*db.ExecAction, err error) {
	cmdbEngine, getDBErr := getCMDBPluginDBResource(ctx)
	if getDBErr != nil {
		err = getDBErr
		return
	}
	var ciTypeRows []*models.SysCiTypeTable
	err = cmdbEngine.SQL("select * from sys_ci_type where status='created'").Find(&ciTypeRows)
	if err != nil {
		err = fmt.Errorf("query ci type table fail,%s ", err.Error())
		return
	}
	var ciTypeAttrRows []*models.SysCiTypeAttrTable
	err = cmdbEngine.SQL("select * from sys_ci_type_attr where status='created'").Find(&ciTypeAttrRows)
	if err != nil {
		err = fmt.Errorf("query ci type attribute table fail,%s ", err.Error())
		return
	}
	ciTypeMap := make(map[string]*models.SysCiTypeTable)
	for _, row := range ciTypeRows {
		ciTypeMap[row.Id] = row
	}
	ciTypeAttrMap := make(map[string][]*models.SysCiTypeAttrTable)
	for _, row := range ciTypeAttrRows {
		if v, ok := ciTypeAttrMap[row.CiType]; ok {
			ciTypeAttrMap[row.CiType] = append(v, row)
		} else {
			ciTypeAttrMap[row.CiType] = []*models.SysCiTypeAttrTable{row}
		}
	}
	transConfig, getConfigErr := getDataTransVariableMap(ctx)
	if getConfigErr != nil {
		err = getConfigErr
		return
	}
	ciTypeDataMap := make(map[string]*models.CiTypeData)
	filters := []*models.CiTypeDataFilter{{CiType: transConfig.EnvCiType, Condition: "in", GuidList: []string{param.Env}}}
	log.Logger.Info("<--- start analyzeCMDBData --->")
	err = analyzeCMDBData(transConfig.BusinessCiType, param.Business, filters, ciTypeAttrMap, ciTypeDataMap, cmdbEngine, transConfig)
	if err != nil {
		log.Logger.Error("<--- fail analyzeCMDBData --->", log.Error(err))
		return
	}
	log.Logger.Info("<--- done analyzeCMDBData --->")
	// 写入cmdb ci数据
	for k, ciData := range ciTypeDataMap {
		ciData.CiType = ciTypeMap[k]
		ciData.Attributes = ciTypeAttrMap[k]
	}
	actions = getInsertAnalyzeCMDBActions(param.TransExportId, ciTypeDataMap)
	// 写入cmdb 报表和视图列表
	reportViewActions, reportViewErr := analyzeCMDBReportViewData(param.TransExportId, cmdbEngine)
	if reportViewErr != nil {
		err = reportViewErr
		return
	}
	actions = append(actions, reportViewActions...)
	// 分析监控数据
	var endpointList, serviceGroupList []string
	for _, ciData := range ciTypeDataMap {
		for _, attr := range ciData.Attributes {
			if attr.ExtRefEntity == "monitor:endpoint" {
				for _, rowData := range ciData.DataMap {
					endpointList = append(endpointList, rowData[attr.Name])
				}
			} else if attr.ExtRefEntity == "monitor:service_group" {
				for _, rowData := range ciData.DataMap {
					serviceGroupList = append(serviceGroupList, rowData[attr.Name])
				}
			}
		}
	}
	monitorActions, analyzeMonitorErr := analyzePluginMonitorExportData(param.TransExportId, trimAndSortStringList(endpointList), trimAndSortStringList(serviceGroupList))
	if analyzeMonitorErr != nil {
		err = fmt.Errorf("analyze monitor export data fail,%s ", analyzeMonitorErr.Error())
		return
	}
	actions = append(actions, monitorActions...)
	// 分析物料包数据
	artifactActions, analyzeArtifactErr := analyzeArtifactExportData(param.TransExportId, ciTypeDataMap, transConfig)
	if analyzeArtifactErr != nil {
		err = fmt.Errorf("analyze artifact export data fail,%s ", analyzeArtifactErr.Error())
		return
	}
	actions = append(actions, artifactActions...)
	return
}

func analyzeCMDBData(ciType string, ciDataGuidList []string, filters []*models.CiTypeDataFilter, ciTypeAttrMap map[string][]*models.SysCiTypeAttrTable, ciTypeDataMap map[string]*models.CiTypeData, cmdbEngine *xorm.Engine, transConfig *models.TransDataVariableConfig) (err error) {
	log.Logger.Info("analyzeCMDBData", log.String("ciType", ciType), log.StringList("guidList", ciDataGuidList))
	if len(ciDataGuidList) == 0 {
		return
	}
	ciTypeAttributes := ciTypeAttrMap[ciType]
	var queryFilterList []string
	queryFilterList = append(queryFilterList, fmt.Sprintf("guid in ('%s')", strings.Join(ciDataGuidList, "','")))
	for _, filterObj := range filters {
		filterSql, buildFilterSqlErr := getCMDBFilterSql(ciTypeAttributes, filterObj, cmdbEngine)
		if buildFilterSqlErr != nil {
			err = fmt.Errorf("try to build filter sql with ciType:%s fail,%s ", ciType, buildFilterSqlErr.Error())
			break
		}
		if filterSql != "" {
			queryFilterList = append(queryFilterList, filterSql)
		}
	}
	if err != nil {
		return
	}
	queryCiDataResult, queryErr := cmdbEngine.QueryString(fmt.Sprintf("select * from %s where "+strings.Join(queryFilterList, " and "), ciType))
	if queryErr != nil {
		err = fmt.Errorf("query ciType:%s data fail,%s ", ciType, queryErr.Error())
		return
	}
	if len(queryCiDataResult) == 0 {
		return
	}
	var newRows []map[string]string
	var newRowsGuidList []string
	if existData, ok := ciTypeDataMap[ciType]; ok {
		for _, row := range queryCiDataResult {
			if _, existFlag := existData.DataMap[row["guid"]]; !existFlag {
				newRows = append(newRows, row)
			}
		}
		if len(newRows) == 0 {
			// 此次数据已经全在分析过的数据中，不用再递归了
			return
		}
		for _, row := range newRows {
			tmpRowGuid := row["guid"]
			existData.DataMap[tmpRowGuid] = row
			newRowsGuidList = append(newRowsGuidList, tmpRowGuid)
		}
	} else {
		dataMap := make(map[string]map[string]string)
		for _, row := range queryCiDataResult {
			tmpRowGuid := row["guid"]
			dataMap[tmpRowGuid] = row
			newRowsGuidList = append(newRowsGuidList, tmpRowGuid)
			newRows = append(newRows, row)
		}
		ciTypeDataMap[ciType] = &models.CiTypeData{DataMap: dataMap}
	}
	// 往下分析数据行的依赖
	for _, attr := range ciTypeAttributes {
		if attr.InputType == "ref" {
			if checkArtifactCiTypeRefIllegal(ciType, attr.RefCiType, transConfig) {
				continue
			}
			refCiTypeGuidList := []string{}
			for _, row := range newRows {
				tmpRefCiDataGuid := row[attr.RefCiType]
				if tmpRefCiDataGuid != "" {
					refCiTypeGuidList = append(refCiTypeGuidList, tmpRefCiDataGuid)
				}
			}
			if len(refCiTypeGuidList) > 0 {
				if err = analyzeCMDBData(attr.RefCiType, refCiTypeGuidList, filters, ciTypeAttrMap, ciTypeDataMap, cmdbEngine, transConfig); err != nil {
					break
				}
			}
		} else if attr.InputType == "multiRef" {
			if checkArtifactCiTypeRefIllegal(ciType, attr.RefCiType, transConfig) {
				continue
			}
			toGuidList, getMultiToGuidErr := getCMDBMultiRefGuidList(ciType, attr.Name, "in", newRowsGuidList, []string{}, cmdbEngine)
			if getMultiToGuidErr != nil {
				err = fmt.Errorf("try to analyze ciType:%s dependent multiRef attr:%s error:%s ", ciType, attr.Name, getMultiToGuidErr.Error())
				break
			}
			if len(toGuidList) > 0 {
				if err = analyzeCMDBData(attr.RefCiType, toGuidList, filters, ciTypeAttrMap, ciTypeDataMap, cmdbEngine, transConfig); err != nil {
					break
				}
			}
		}
	}
	if err != nil {
		return
	}
	// 往下分析数据行的被依赖
	for depCiType, depCiTypeAttrList := range ciTypeAttrMap {
		for _, depCiAttr := range depCiTypeAttrList {
			if depCiAttr.RefCiType == ciType {
				if depCiAttr.InputType == "ref" {
					queryDepCiGuidRows, tmpQueryDepCiGuidErr := cmdbEngine.QueryString(fmt.Sprintf("select guid from %s where %s in ('%s')", depCiType, depCiAttr.Name, strings.Join(newRowsGuidList, "','")))
					if tmpQueryDepCiGuidErr != nil {
						err = fmt.Errorf("try to get ciType:%s with dependent ciType:%s ref attr:%s guidList fail,%s ", ciType, depCiType, depCiAttr.Name, tmpQueryDepCiGuidErr.Error())
						break
					}
					if len(queryDepCiGuidRows) > 0 {
						depCiGuidList := []string{}
						for _, row := range queryDepCiGuidRows {
							depCiGuidList = append(depCiGuidList, row["guid"])
						}
						if err = analyzeCMDBData(depCiType, depCiGuidList, filters, ciTypeAttrMap, ciTypeDataMap, cmdbEngine, transConfig); err != nil {
							break
						}
					}
				} else if depCiAttr.InputType == "multiRef" {
					depFromGuidList, tmpQueryDepCiGuidErr := getCMDBMultiRefGuidList(depCiType, depCiAttr.Name, "in", []string{}, newRowsGuidList, cmdbEngine)
					if tmpQueryDepCiGuidErr != nil {
						err = fmt.Errorf("try to get ciType:%s with dependent ciType:%s multiRef attr:%s guidList fail,%s ", ciType, depCiType, depCiAttr.Name, tmpQueryDepCiGuidErr.Error())
						break
					}
					if len(depFromGuidList) > 0 {
						if err = analyzeCMDBData(depCiType, depFromGuidList, filters, ciTypeAttrMap, ciTypeDataMap, cmdbEngine, transConfig); err != nil {
							break
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		return
	}
	return
}

func getCMDBPluginDBResource(ctx context.Context) (dbEngine *xorm.Engine, err error) {
	pluginMysqlInstance, getInstanceErr := GetPluginMysqlInstance(ctx, "wecmdb")
	if getInstanceErr != nil {
		err = fmt.Errorf("get cmdb mysql instance fail,%s ", getInstanceErr.Error())
		return
	}
	var resourceServerRows []*models.ResourceServer
	err = db.MysqlEngine.Context(ctx).SQL("select * from resource_server where id in (select resource_server_id from resource_item where id=?)", pluginMysqlInstance.ResourceItemId).Find(&resourceServerRows)
	if err != nil {
		return
	}
	if len(resourceServerRows) == 0 {
		err = fmt.Errorf("can not find resource server with item:%s ", pluginMysqlInstance.ResourceItemId)
		return
	}
	dbConfig := models.DatabaseConfig{
		User:     pluginMysqlInstance.Username,
		Password: pluginMysqlInstance.Password,
		Server:   resourceServerRows[0].Host,
		Port:     resourceServerRows[0].Port,
		DataBase: pluginMysqlInstance.SchemaName,
		MaxOpen:  5,
		MaxIdle:  5,
		Timeout:  60,
	}
	dbEngine, err = db.GetDatabaseEngine(&dbConfig)
	return
}

func getDataTransVariableMap(ctx context.Context) (result *models.TransDataVariableConfig, err error) {
	result = &models.TransDataVariableConfig{}
	var sysVarRows []*models.SystemVariables
	err = db.MysqlEngine.Context(ctx).SQL("select name,value,default_value from system_variables where status='active' and name like 'PLATFORM_EXPORT_%'").Find(&sysVarRows)
	if err != nil {
		err = fmt.Errorf("query system variable table fail,%s ", err.Error())
		return
	}
	for _, row := range sysVarRows {
		tmpValue := row.DefaultValue
		if row.Value != "" {
			tmpValue = row.Value
		}
		switch row.Name {
		case "PLATFORM_EXPORT_CI_BUSINESS":
			result.BusinessCiType = tmpValue
			if tmpValue == "" {
				err = fmt.Errorf("PLATFORM_EXPORT_CI_BUSINESS is empty")
			}
		case "PLATFORM_EXPORT_CI_ENV":
			result.EnvCiType = tmpValue
			if tmpValue == "" {
				err = fmt.Errorf("PLATFORM_EXPORT_CI_ENV is empty")
			}
		case "PLATFORM_EXPORT_NEXUS_URL":
			result.NexusUrl = tmpValue
		case "PLATFORM_EXPORT_NEXUS_USER":
			result.NexusUser = tmpValue
		case "PLATFORM_EXPORT_NEXUS_PWD":
			result.NexusPwd = tmpValue
		case "PLATFORM_EXPORT_NEXUS_REPO":
			result.NexusRepo = tmpValue
		case "PLATFORM_EXPORT_CI_ARTIFACT_INSTANCE":
			if tmpValue != "" {
				result.ArtifactInstanceCiTypeList = strings.Split(tmpValue, ",")
			}
		case "PLATFORM_EXPORT_CI_ARTIFACT_PACKAGE":
			result.ArtifactPackageCiType = tmpValue
		case "PLATFORM_EXPORT_CI_SYSTEM":
			result.ArtifactCiSystem = tmpValue
		case "PLATFORM_EXPORT_CI_TECH_PRODUCT":
			result.ArtifactCiTechProduct = tmpValue
		}
	}
	return
}

func getCMDBFilterSql(ciTypeAttributes []*models.SysCiTypeAttrTable, filter *models.CiTypeDataFilter, cmdbEngine *xorm.Engine) (filterSql string, err error) {
	matchAttr := &models.SysCiTypeAttrTable{}
	for _, attr := range ciTypeAttributes {
		if attr.RefCiType == filter.CiType {
			matchAttr = attr
			break
		}
	}
	if matchAttr.Id == "" {
		return
	}
	condition := "in"
	if filter.Condition == "notIn" {
		condition = "not in"
	}
	if matchAttr.InputType == "multiRef" {
		var fromGuidList []string
		if fromGuidList, err = getCMDBMultiRefGuidList(matchAttr.CiType, matchAttr.Name, condition, []string{}, filter.GuidList, cmdbEngine); err != nil {
			return
		}
		filterSql = fmt.Sprintf("guid in ('%s')", strings.Join(fromGuidList, "','"))
	} else if matchAttr.InputType == "ref" {
		filterSql = fmt.Sprintf("%s %s ('%s')", matchAttr.Name, condition, strings.Join(filter.GuidList, "','"))
	} else {
		err = fmt.Errorf("ciTypeAttr:%s refCiType:%s illegal with inputType:%s ", matchAttr.Id, filter.CiType, matchAttr.InputType)
	}
	return
}

func getCMDBMultiRefGuidList(ciType, attrName, condition string, fromGuidList, toGuidList []string, cmdbEngine *xorm.Engine) (resultGuidList []string, err error) {
	var filterColumn, filterSql, resultColumn string
	if len(fromGuidList) > 0 {
		filterColumn = "from_guid"
		resultColumn = "to_guid"
		filterSql = strings.Join(fromGuidList, "','")
	} else if len(toGuidList) > 0 {
		filterColumn = "to_guid"
		resultColumn = "from_guid"
		filterSql = strings.Join(toGuidList, "','")
	} else {
		return
	}
	queryResult, queryErr := cmdbEngine.QueryString(fmt.Sprintf("select from_guid,to_guid from `%s$%s` where %s %s ('%s')", ciType, attrName, filterColumn, condition, filterSql))
	if queryErr != nil {
		err = fmt.Errorf("query multiRef list fail,ciType:%s attrName:%s,error:%s ", ciType, attrName, queryErr.Error())
		return
	}
	for _, row := range queryResult {
		resultGuidList = append(resultGuidList, row[resultColumn])
	}
	return
}

func getInsertAnalyzeCMDBActions(transExportId string, ciTypeDataMap map[string]*models.CiTypeData) (actions []*db.ExecAction) {
	nowTime := time.Now()
	for ciType, ciTypeData := range ciTypeDataMap {
		rowDataBytes, _ := json.Marshal(ciTypeData.DataMap)
		actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data,data_len,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
			"ex_aly_" + guid.CreateGuid(), transExportId, "wecmdb", ciType, ciTypeData.CiType.DisplayName, string(rowDataBytes), len(ciTypeData.DataMap), nowTime,
		}})
	}
	return
}

func getInsertTransExport(transExport models.TransExportTable) (actions []*db.ExecAction) {
	nowTime := time.Now()
	actions = []*db.ExecAction{}
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export(id,business,business_name,environment,environment_name,status,output_url,created_user,created_time,updated_user,updated_time) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		transExport.Id, transExport.Business, transExport.BusinessName, transExport.Environment, transExport.EnvironmentName, transExport.Status, transExport.OutputUrl, transExport.CreatedUser, nowTime, transExport.UpdatedUser, nowTime,
	}})
	return
}

func getUpdateTransExport(transExport models.TransExportTable) (actions []*db.ExecAction) {

	return
}

func QueryBusinessList(c context.Context, userToken, language string, param models.QueryBusinessParam) (result []map[string]interface{}, err error) {
	var query models.QueryBusinessListParam
	var dataTransVariableConfig *models.TransDataVariableConfig
	if dataTransVariableConfig, err = getDataTransVariableMap(c); err != nil {
		return
	}
	if dataTransVariableConfig == nil {
		return
	}
	query = models.QueryBusinessListParam{
		PackageName: "wecmdb",
		UserToken:   userToken,
		Language:    language,
		EntityQueryParam: models.EntityQueryParam{
			AdditionalFilters: make([]*models.EntityQueryObj, 0),
		},
	}
	if param.QueryMode == "env" {
		query.Entity = dataTransVariableConfig.EnvCiType
	} else {
		query.Entity = dataTransVariableConfig.BusinessCiType
		if strings.TrimSpace(param.ID) != "" {
			query.EntityQueryParam.AdditionalFilters = append(query.EntityQueryParam.AdditionalFilters, &models.EntityQueryObj{
				AttrName:  "id",
				Op:        "like",
				Condition: param.ID,
			})
		}
		if strings.TrimSpace(param.DisplayName) != "" {
			query.EntityQueryParam.AdditionalFilters = append(query.EntityQueryParam.AdditionalFilters, &models.EntityQueryObj{
				AttrName:  "displayName",
				Op:        "like",
				Condition: param.DisplayName,
			})
		}
	}
	result, err = remote.QueryBusinessList(query)
	return
}

func CreateExport(c context.Context, param models.CreateExportParam, operator string) (transExportId string, err error) {
	var actions, addTransExportActions []*db.ExecAction
	transExportId = guid.CreateGuid()
	transExport := models.TransExportTable{
		Id:          transExportId,
		Environment: param.Env,
		Business:    strings.Join(param.PIds, ","),
		Status:      string(models.TransExportStatusStart),
		CreatedUser: operator,
		UpdatedUser: operator,
	}
	if addTransExportActions = getInsertTransExport(transExport); len(addTransExportActions) > 0 {
		actions = append(actions, addTransExportActions...)
	}
	return
}

func checkArtifactCiTypeRefIllegal(sourceCiType, refCiType string, transConfig *models.TransDataVariableConfig) (illegal bool) {
	illegal = false
	if refCiType == transConfig.ArtifactPackageCiType {
		matchFlag := false
		for _, v := range transConfig.ArtifactInstanceCiTypeList {
			if sourceCiType == v {
				matchFlag = true
				break
			}
		}
		if !matchFlag {
			illegal = true
		}
	}
	return
}

func trimAndSortStringList(input []string) (output []string) {
	for _, v := range input {
		if v == "" {
			continue
		}
		output = append(output, v)
	}
	sort.Strings(output)
	return
}

func analyzePluginMonitorExportData(transExportId string, endpointList, serviceGroupList []string) (actions []*db.ExecAction, err error) {
	analyzeResult, analyzeErr := monitor.GetMonitorExportAnalyzeData(endpointList, serviceGroupList)
	if analyzeErr != nil {
		err = analyzeErr
		return
	}
	nowTime := time.Now()
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "monitor_type", "monitor_type", len(analyzeResult.MonitorType), parseStringListToJsonString(analyzeResult.MonitorType), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "endpoint_group", "endpoint_group", len(analyzeResult.EndpointGroup), parseStringListToJsonString(analyzeResult.EndpointGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "custom_metric_service_group", "custom_metric_service_group", len(analyzeResult.CustomMetricServiceGroup), parseStringListToJsonString(analyzeResult.CustomMetricServiceGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "custom_metric_endpoint_group", "custom_metric_endpoint_group", len(analyzeResult.CustomMetricEndpointGroup), parseStringListToJsonString(analyzeResult.CustomMetricEndpointGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "custom_metric_monitor_type", "custom_metric_monitor_type", len(analyzeResult.CustomMetricMonitorType), parseStringListToJsonString(analyzeResult.CustomMetricMonitorType), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "log_monitor_service_group", "log_monitor_service_group", len(analyzeResult.LogMonitorServiceGroup), parseStringListToJsonString(analyzeResult.LogMonitorServiceGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "log_monitor_template", "log_monitor_template", len(analyzeResult.LogMonitorTemplate), parseStringListToJsonString(analyzeResult.LogMonitorTemplate), nowTime,
	}})

	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "strategy_service_group", "strategy_service_group", len(analyzeResult.StrategyServiceGroup), parseStringListToJsonString(analyzeResult.StrategyServiceGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "strategy_endpoint_group", "strategy_endpoint_group", len(analyzeResult.StrategyEndpointGroup), parseStringListToJsonString(analyzeResult.StrategyEndpointGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "logKeyword_service_group", "logKeyword_service_group", len(analyzeResult.LogKeywordServiceGroup), parseStringListToJsonString(analyzeResult.LogKeywordServiceGroup), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "monitor", "dashboard", "dashboard", len(analyzeResult.DashboardIdList), parseStringListToJsonString(analyzeResult.DashboardIdList), nowTime,
	}})
	return
}

func parseStringListToJsonString(input []string) string {
	b, _ := json.Marshal(input)
	return string(b)
}

func analyzeArtifactExportData(transExportId string, ciTypeDataMap map[string]*models.CiTypeData, transConfig *models.TransDataVariableConfig) (actions []*db.ExecAction, err error) {
	nowTime := time.Now()
	if ciTypeData, ok := ciTypeDataMap[transConfig.ArtifactPackageCiType]; ok {
		b, _ := json.Marshal(ciTypeData.DataMap)
		actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
			"ex_aly_" + guid.CreateGuid(), transExportId, "artifact", "deploy_package", "deploy_package", len(ciTypeData.DataMap), string(b), nowTime,
		}})
	}
	return
}

func analyzeCMDBReportViewData(transExportId string, cmdbEngine *xorm.Engine) (actions []*db.ExecAction, err error) {
	nowTime := time.Now()
	var reportRows []*models.SysReportTable
	err = cmdbEngine.SQL("select * from sys_report").Find(&reportRows)
	if err != nil {
		err = fmt.Errorf("query cmdb sys report table fail,%s", err.Error())
		return
	}
	var viewRows []*models.SysViewTable
	err = cmdbEngine.SQL("select * from sys_view").Find(&viewRows)
	if err != nil {
		err = fmt.Errorf("query cmdb sys view table fail,%s", err.Error())
		return
	}
	reportRowBytes, _ := json.Marshal(reportRows)
	viewRowBytes, _ := json.Marshal(viewRows)
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "wecmdb_report", "report", "report", len(reportRows), string(reportRowBytes), nowTime,
	}})
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "wecmdb_view", "view", "view", len(viewRows), string(viewRowBytes), nowTime,
	}})
	return
}

func DataTransExportCMDBData(ctx context.Context, transExportId, path string) (err error) {
	cmdbEngine, getDBErr := getCMDBPluginDBResource(ctx)
	if getDBErr != nil {
		err = getDBErr
		return
	}
	var sqlBytes []byte
	sqlBuffer := bytes.NewBuffer(sqlBytes)
	var ciTypeList, reportList, viewList []string
	// 读analyze表cmdb数据
	var transExportAnalyzeRows []*models.TransExportAnalyzeDataTable
	err = db.MysqlEngine.Context(ctx).SQL("select `source`,data_type,`data`,data_len from trans_export_analyze_data where trans_export=? and `source` in ('wecmdb','wecmdb_report','wecmdb_view') order by data_type", transExportId).Find(&transExportAnalyzeRows)
	if err != nil {
		err = fmt.Errorf("query trans export analyze table data fail,%s ", err.Error())
		return
	}
	ciDataGuidMap := make(map[string][]string)
	for _, row := range transExportAnalyzeRows {
		if row.Source == "wecmdb_report" {
			var tmpReportList []*models.SysReportTable
			if tmpErr := json.Unmarshal([]byte(row.Data), &tmpReportList); tmpErr != nil {
				err = fmt.Errorf("json unmarshal report data fail,%s ", tmpErr.Error())
				break
			}
			for _, v := range tmpReportList {
				reportList = append(reportList, v.Id)
			}
		} else if row.Source == "wecmdb_view" {
			tmpViewList := []*models.SysViewTable{}
			if tmpErr := json.Unmarshal([]byte(row.Data), &tmpViewList); tmpErr != nil {
				err = fmt.Errorf("json unmarshal view data fail,%s ", tmpErr.Error())
				break
			}
			for _, v := range tmpViewList {
				reportList = append(reportList, v.Id)
			}
		} else {
			tmpDataMap := make(map[string]interface{})
			if tmpErr := json.Unmarshal([]byte(row.Data), &tmpDataMap); tmpErr != nil {
				err = fmt.Errorf("json unmarshal ciType:%s data fail,%s ", row.DataType, tmpErr.Error())
				break
			}
			ciTypeList = append(ciTypeList, row.DataType)
			tmpCiDataGuidList := []string{}
			for k, _ := range tmpDataMap {
				tmpCiDataGuidList = append(tmpCiDataGuidList, k)
			}
			ciDataGuidMap[row.DataType] = tmpCiDataGuidList
		}
	}
	if err != nil {
		return
	}
	tables, getDBMetaErr := cmdbEngine.DBMetas()
	if getDBMetaErr != nil {
		err = fmt.Errorf("get db meta error:%s ", getDBMetaErr.Error())
		return
	}
	ciTypeFilterSql := strings.Join(ciTypeList, "','")
	reportFilterSql := strings.Join(reportList, "','")
	viewFilterSql := strings.Join(viewList, "','")
	sqlBuffer.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_basekey_cat", "", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_basekey_code", "", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_files", "select * from sys_files where guid in (select image_file from sys_ci_type where id in ('"+ciTypeFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_state_machine", "select * from sys_state_machine where id in (select state_machine from sys_ci_template where id in (select ci_template from sys_ci_type where id in ('"+ciTypeFilterSql+"')))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_state", "select * from sys_state where state_machine in (select state_machine from sys_ci_template where id in (select ci_template from sys_ci_type where id in ('"+ciTypeFilterSql+"')))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_state_transition", "select * from sys_state_transition where state_machine in (select state_machine from sys_ci_template where id in (select ci_template from sys_ci_type where id in ('"+ciTypeFilterSql+"')))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_ci_template", "select * from sys_ci_template where id in (select ci_template from sys_ci_type where id in ('"+ciTypeFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_ci_template_attr", "select * from sys_ci_template_attr where ci_template in (select ci_template from sys_ci_type where id in ('"+ciTypeFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_ci_type", "select * from sys_ci_type where id in ('"+ciTypeFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_ci_type_attr", "select * from sys_ci_type_attr where ci_type in ('"+ciTypeFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role", "select * from sys_role", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_ci_type", "select * from sys_role_ci_type where ci_type in ('"+ciTypeFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_ci_type_condition", "select * from sys_role_ci_type_condition where role_ci_type in (select guid from sys_role_ci_type where ci_type in ('"+ciTypeFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_ci_type_condition_filter", "select * from sys_role_ci_type_condition_filter where role_ci_type_condition in (select guid from sys_role_ci_type_condition where role_ci_type in (select guid from sys_role_ci_type where ci_type in ('"+ciTypeFilterSql+"')))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_ci_type_list", "select * from sys_role_ci_type_list where role_ci_type in (select guid from sys_role_ci_type where ci_type in ('"+ciTypeFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_report", "select * from sys_report where id in ('"+reportFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_report_object", "select * from sys_report_object where report in ('"+reportFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_report_object_attr", "select * from sys_report_object_attr where report_object in (select id from sys_report_object where report in ('"+reportFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_report", "select * from sys_role_report where report in ('"+reportFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_view", "select * from sys_view where id in ('"+viewFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_graph", "select * from sys_graph where `view` in ('"+viewFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_graph_element", "select * from sys_graph_element where graph in (select id from sys_graph where `view` in ('"+viewFilterSql+"'))", sqlBuffer); err != nil {
		return
	}
	if err = dumpCMDBTableData(cmdbEngine, tables, "sys_role_view", "select * from sys_role_view where `view` in ('"+viewFilterSql+"')", sqlBuffer); err != nil {
		return
	}
	for _, ciType := range ciTypeList {
		tmpQuerySql := "select * from " + ciType
		if tmpGuidList, ok := ciDataGuidMap[ciType]; ok {
			tmpQuerySql += " where guid in ('" + strings.Join(tmpGuidList, "','") + "')"
		}
		if err = dumpCMDBTableData(cmdbEngine, tables, ciType, tmpQuerySql, sqlBuffer); err != nil {
			return
		}
		if err = dumpCMDBTableData(cmdbEngine, tables, "history_"+ciType, tmpQuerySql, sqlBuffer); err != nil {
			return
		}
	}
	sqlBuffer.WriteString("\nSET FOREIGN_KEY_CHECKS=1;\n")
	tmpFilePath := fmt.Sprintf("%s/wecmdb_data.sql", path)
	err = os.WriteFile(tmpFilePath, sqlBuffer.Bytes(), 0666)
	if err != nil {
		err = fmt.Errorf("try to write cmdb database dump file fail,%s ", err.Error())
	}
	return
}

func dumpCMDBTableData(cmdbEngine *xorm.Engine, tables []*schemas.Table, tableName, querySql string, bf *bytes.Buffer) (err error) {
	var columnNameList []string
	var tableObj *schemas.Table
	for _, t := range tables {
		if t.Name == tableName {
			tableObj = t
			for _, c := range t.Columns() {
				columnNameList = append(columnNameList, c.Name)
			}
		}
	}
	if tableObj == nil {
		err = fmt.Errorf("tableName:%s illegal", tableName)
		return
	}
	if !strings.HasPrefix(tableName, "sys_") {
		// 如果不是系统表，要把表结构导出来
		queryTableRows, queryTableErr := cmdbEngine.QueryString("show create table " + tableName)
		if queryTableErr != nil {
			err = fmt.Errorf("query cmdb table %s struct fail,error:%s ", tableName, queryTableErr.Error())
			return
		}
		if len(queryTableRows) > 0 {
			bf.WriteString(queryTableRows[0]["Create Table"] + ";\n")
		} else {
			err = fmt.Errorf("can not find table %s struct", tableName)
			return
		}
	}
	if strings.HasPrefix(tableName, "history_") {
		return
	}
	if querySql == "" {
		querySql = fmt.Sprintf("select * from " + tableName)
	}
	queryRows, queryErr := cmdbEngine.Query(querySql)
	if queryErr != nil {
		err = fmt.Errorf("query cmdb table %s data fail,%s ", tableName, queryErr.Error())
		return
	}
	var rowValueList []string
	for _, row := range queryRows {
		tmpValueList := []string{}
		for _, c := range tableObj.Columns() {
			if c.SQLType.IsBlob() {
				tmpValueList = append(tmpValueList, "0x"+hex.EncodeToString(row[c.Name]))
				continue
			}
			tmpStringValue := string(row[c.Name])
			tmpStringValue = strings.ReplaceAll(tmpStringValue, "'", "\\'")
			tmpValueList = append(tmpValueList, "'"+tmpStringValue+"'")
		}
		rowValueList = append(rowValueList, strings.Join(tmpValueList, ","))
	}
	if len(rowValueList) == 0 {
		return
	}
	for _, v := range rowValueList {
		bf.WriteString("INSERT INTO " + tableName + " (`" + strings.Join(columnNameList, "`,`") + "`) VALUES (" + v + ");\n")
	}
	return
}

func DataTransImportCMDBData(ctx context.Context, inputFile string) (err error) {
	cmdbEngine, getDBErr := getCMDBPluginDBResource(ctx)
	if getDBErr != nil {
		err = getDBErr
		return
	}
	session := cmdbEngine.NewSession()
	session.Begin()
	fileBytes, openFileErr := os.ReadFile(inputFile)
	if openFileErr != nil {
		err = openFileErr
		return
	}
	for _, lineSql := range strings.Split(string(fileBytes), ";\n") {
		if lineSql == "" {
			continue
		}
		_, tmpErr := session.Exec(lineSql)
		if tmpErr != nil {
			if strings.Contains(tmpErr.Error(), "Duplicate entry") {
				continue
			}
			err = tmpErr
			break
		}
	}
	if err != nil {
		fmt.Printf("error:%s \n", err.Error())
		if rollbackErr := session.Rollback(); rollbackErr != nil {
			fmt.Printf("rollback err:%s \n", rollbackErr.Error())
		} else {
			fmt.Println("rollback done")
		}
	} else {
		if commitErr := session.Commit(); commitErr != nil {
			fmt.Printf("commit err:%s \n", commitErr.Error())
		} else {
			fmt.Println("commit done")
		}
	}
	session.Close()
	return
}

func DataTransExportMonitorData(transExportId string) (tmpFilePathList []string, err error) {

	return
}

func DataTransExportArtifactData(transExportId string) (tmpFilePathList []string, err error) {

	return
}

// AnalyzePluginConfigDataExport 分析插件服务和系统数据数据
func AnalyzePluginConfigDataExport(ctx context.Context, transExportId string) (actions []*db.ExecAction, err error) {
	var pluginPackageRows []*models.PluginPackages
	err = db.MysqlEngine.Context(ctx).SQL("select id,name,`version` from plugin_packages where status='REGISTERED' and id in (select package_id from plugin_instances where container_status='RUNNING')").Find(&pluginPackageRows)
	if err != nil {
		err = fmt.Errorf("query export plugin package table fail,%s ", err.Error())
		return
	}
	if len(pluginPackageRows) == 0 {
		return
	}
	var analyzeData []*models.DataTransPluginExportData
	var pluginPackageIdList, variableSourceList []string
	for _, row := range pluginPackageRows {
		tmpSource := fmt.Sprintf("%s__%s", row.Name, row.Version)
		tmpAnalyzeData := models.DataTransPluginExportData{PluginPackageId: row.Id, Source: tmpSource, PluginPackageName: row.Name}
		analyzeData = append(analyzeData, &tmpAnalyzeData)
		pluginPackageIdList = append(pluginPackageIdList, row.Id)
		variableSourceList = append(variableSourceList, tmpSource)
	}
	filterSql, filterParam := db.CreateListParams(pluginPackageIdList, "")
	var interfaceQueryRows []*models.DataTransPluginExportData
	err = db.MysqlEngine.Context(ctx).SQL("select t3.plugin_package_id,count(1) as plugin_interface_num from (select t1.id,t2.plugin_package_id from plugin_config_interfaces t1 left join plugin_configs t2 on t1.plugin_config_id=t2.id where t2.plugin_package_id in ("+filterSql+")) t3 group by t3.plugin_package_id", filterParam...).Find(&interfaceQueryRows)
	if err != nil {
		err = fmt.Errorf("query plugin config interface table fail,%s ", err.Error())
		return
	}
	for _, v := range analyzeData {
		for _, row := range interfaceQueryRows {
			if v.PluginPackageId == row.PluginPackageId {
				v.PluginInterfaceNum = row.PluginInterfaceNum
				break
			}
		}
	}
	var variableQueryRows []*models.DataTransPluginExportData
	sourceFilterSql, sourceFilterParam := db.CreateListParams(variableSourceList, "")
	err = db.MysqlEngine.Context(ctx).SQL("select source,count(1) as system_variable_num from system_variables where source in ("+sourceFilterSql+") group by source", sourceFilterParam...).Find(&variableQueryRows)
	if err != nil {
		err = fmt.Errorf("query plugin system variable table fail,%s ", err.Error())
		return
	}
	for _, v := range analyzeData {
		for _, row := range variableQueryRows {
			if v.Source == row.Source {
				v.SystemVariableNum = row.SystemVariableNum
				break
			}
		}
	}
	analyzeBytes, _ := json.Marshal(analyzeData)
	actions = append(actions, &db.ExecAction{Sql: "insert into trans_export_analyze_data(id,trans_export,source,data_type,data_type_name,data_len,data,start_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
		"ex_aly_" + guid.CreateGuid(), transExportId, "plugin_package", "plugin_package", "plugin_package", len(analyzeData), string(analyzeBytes), time.Now(),
	}})
	return
}

func DataTransExportPluginConfig(ctx context.Context, transExportId, path string) (err error) {
	// 读analyze表plugin_package数据
	var transExportAnalyzeRows []*models.TransExportAnalyzeDataTable
	err = db.MysqlEngine.Context(ctx).SQL("select `source`,data_type,`data`,data_len from trans_export_analyze_data where trans_export=? and `source`='plugin_package'", transExportId).Find(&transExportAnalyzeRows)
	if err != nil {
		err = fmt.Errorf("query trans export analyze table data fail,%s ", err.Error())
		return
	}
	if len(transExportAnalyzeRows) == 0 {
		log.Logger.Warn("no analyze plugin package data found in database", log.String("transExportId", transExportId))
		return
	}
	var analyzeData []*models.DataTransPluginExportData
	if err = json.Unmarshal([]byte(transExportAnalyzeRows[0].Data), &analyzeData); err != nil {
		err = fmt.Errorf("json unmarshal export plugin package data fail,%s ", err.Error())
		return
	}
	for _, row := range analyzeData {
		retData, tmpErr := ExportPluginConfigs(ctx, row.PluginPackageId, []*models.PluginConfigsBatchEnable{}, []string{"SUPER_ADMIN"})
		if tmpErr != nil {
			err = fmt.Errorf("build plugin export config fail,pluginPackageId:%s,error:%s ", row.PluginPackageId, tmpErr.Error())
			return
		}
		fileName := fmt.Sprintf("%s/plugin-%s-%s-%s.xml", path, retData.Name, retData.Version, time.Now().Format("20060102150405"))
		retDataBytes, _ := xml.MarshalIndent(retData, "", "    ")
		if err = os.WriteFile(fileName, retDataBytes, 0666); err != nil {
			err = fmt.Errorf("wirte export plugin config xml file fail,%s ", err.Error())
			return
		}
	}
	return
}
