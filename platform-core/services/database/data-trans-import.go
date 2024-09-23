package database

import (
	"context"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/db"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/log"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
	"time"
)

func GetDataTransImportNexusConfig(ctx context.Context) (result *models.TransDataImportNexusConfig, err error) {
	result = &models.TransDataImportNexusConfig{}
	var sysVarRows []*models.SystemVariables
	err = db.MysqlEngine.Context(ctx).SQL("select name,value,default_value from system_variables where status='active' and name like 'PLATFORM_IMPORT_%'").Find(&sysVarRows)
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
		case "PLATFORM_IMPORT_NEXUS_URL":
			result.NexusUrl = tmpValue
		case "PLATFORM_IMPORT_NEXUS_USER":
			result.NexusUser = tmpValue
		case "PLATFORM_IMPORT_NEXUS_PWD":
			result.NexusPwd = tmpValue
		case "PLATFORM_IMPORT_NEXUS_REPO":
			result.NexusRepo = tmpValue
		}
	}
	return
}

func GetTransImportWithDetail(ctx context.Context, transImportId string, withDetailData bool) (result *models.TransImportJobParam, err error) {
	var transImportRows []*models.TransImportTable
	err = db.MysqlEngine.Context(ctx).SQL("select * from trans_import where id=?", transImportId).Find(&transImportRows)
	if err != nil {
		err = fmt.Errorf("query trans import table fail,%s ", err.Error())
		return
	}
	if len(transImportRows) == 0 {
		err = fmt.Errorf("can not find trans import with id:%s ", transImportId)
		return
	}
	result = &models.TransImportJobParam{TransImport: transImportRows[0], Details: []*models.TransImportDetailTable{}}
	if withDetailData {
		err = db.MysqlEngine.Context(ctx).SQL("select * from trans_import_detail where trans_import=? order by step").Find(&result.Details)
	} else {
		err = db.MysqlEngine.Context(ctx).SQL("select id,trans_import,name,step,status,error_msg,start_time,end_time from trans_import_detail where trans_import=? order by step").Find(&result.Details)
	}
	if err != nil {
		err = fmt.Errorf("query trans import detail table fail,%s ", err.Error())
	}
	return
}

func UpdateTransImportDetailStatus(ctx context.Context, transImportId, transImportDetailId, status, output, errorMsg string) (err error) {
	var actions []*db.ExecAction
	nowTime := time.Now()
	if status == "success" {
		actions = append(actions, &db.ExecAction{Sql: "update trans_import_detail set status=?,`output`=?,end_time=? where id=?", Param: []interface{}{
			status, output, nowTime, transImportDetailId,
		}})
	} else if status == "fail" {
		actions = append(actions, &db.ExecAction{Sql: "update trans_import_detail set status=?,error_msg=?,end_time=? where id=?", Param: []interface{}{
			status, errorMsg, nowTime, transImportDetailId,
		}})
		actions = append(actions, &db.ExecAction{Sql: "update trans_import set status=?,updated_time=? where id=?", Param: []interface{}{
			status, nowTime, transImportId,
		}})
	} else if status == "doing" {
		actions = append(actions, &db.ExecAction{Sql: "update trans_import_detail set status=?,start_time=? where id=?", Param: []interface{}{
			status, nowTime, transImportDetailId,
		}})
	}
	err = db.Transaction(actions, ctx)
	if err != nil {
		err = fmt.Errorf("update trans import detail status fail,%s ", err.Error())
		log.Logger.Error("UpdateTransImportDetailStatus fail", log.String("transImportId", transImportId), log.String("detailId", transImportDetailId), log.Error(err))
	}
	return
}

func RecordTransImportAction(ctx context.Context, callParam *models.CallTransImportActionParam) (err error) {
	if callParam.ActionId == "" {
		callParam.ActionId = "t_imp_action_" + guid.CreateGuid()
		_, err = db.MysqlEngine.Context(ctx).Exec("insert into trans_import_action(id,trans_import,trans_import_detail,`action`,created_user,updated_time) values (?,?,?,?,?,?)",
			callParam.ActionId, callParam.TransImportId, callParam.TransImportDetailId, callParam.Action, callParam.Operator, time.Now())
	} else {
		_, err = db.MysqlEngine.Context(ctx).Exec("update trans_import_action set error_msg=?,updated_time=? where id=?", callParam.ErrorMsg, time.Now(), callParam.ActionId)
	}
	return
}

func GetTransImportProcExecList() (result []*models.TransImportProcExecTable, err error) {
	err = db.MysqlEngine.SQL("select * from trans_import_proc_exec where trans_import_detail in (select id from trans_import_detail where status='doing' and trans_import in (select id from trans_import where status='doing')) order by trans_import_detail,exec_order").Find(&result)
	if err != nil {
		err = fmt.Errorf("query trans import proc exec table fail,%s ", err.Error())
	}
	return
}
