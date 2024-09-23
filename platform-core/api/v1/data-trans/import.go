package data_trans

import (
	"context"
	"github.com/WeBankPartners/wecube-platform/platform-core/common/log"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
	"github.com/WeBankPartners/wecube-platform/platform-core/services/database"
)

func ExecTransImport(ctx context.Context, param models.ExecImportParam) (transImportId string, err error) {
	var transImport *models.TransImportTable
	// 解压文件
	if _, err = database.DecompressExportZip(ctx, param.ExportNexusUrl, param.TransImportId); err != nil {
		log.Logger.Error("DecompressExportZip err", log.Error(err))
		return
	}
	// 读解压后的文件录进数据库为了给用户展示要导入什么东西
	if transImport, err = database.GetTransImport(ctx, param.TransImportId); err != nil {
		log.Logger.Error("GetTransImport err", log.Error(err))
		return
	}
	// 没有查询到导入记录,表示第一步开始,初始化所有内容
	if transImport == nil {

	}
	// 开始导入
	// 1、导入角色
	// 2、导入cmdb插件服务、导入cmdb数据、同步cmdb数据模型、导入其它插件服务
	// 3、导入编排
	// 4、导入批量执行
	// 5、导入物料包
	// 6、导入监控基础类型、对象组、基础类型指标、对象组指标、对象组阈值配置、业务配置模版
	// 7、导入taskman模版和公共组件
	// 开始执行
	// 8、开始执行编排(创建资源、初始化资源、应用部署)
	// 继续导入
	// 9、导入监控业务配置、层级对象指标、层级对象阈值配置、自定义看板
	return
}

var importFuncList []func(context.Context, *models.TransImportJobParam) (string, error)

func init() {
	importFuncList = append(importFuncList, importRole)
	importFuncList = append(importFuncList, importPluginConfig)
	importFuncList = append(importFuncList, importWorkflow)
	importFuncList = append(importFuncList, importBatchExecution)
	importFuncList = append(importFuncList, importArtifactPackage)
	importFuncList = append(importFuncList, importMonitorBaseConfig)
	importFuncList = append(importFuncList, importTaskmanTemplate)
	importFuncList = append(importFuncList, execWorkflow)
	importFuncList = append(importFuncList, importMonitorServiceConfig)
}

func CallImportFunc(ctx context.Context, callParam *models.CallTransImportActionParam) (err error) {
	transImportJobParam, getConfigErr := database.GetTransImportWithDetail(ctx, callParam.TransImportId, false)
	if getConfigErr != nil {
		err = getConfigErr
		return
	}
	if callParam.Action == "start" {
		var currentStep int
		for _, detailRow := range transImportJobParam.Details {
			if detailRow.Status == "notStart" {
				currentStep = detailRow.Step
				break
			}
		}
		if currentStep == 8 {

		} else if currentStep == 9 {

		} else {
			for currentStep <= 7 {
				transImportJobParam.CurrentDetail = transImportJobParam.Details[currentStep-1]
				funcObj := importFuncList[currentStep-1]
				tmpOutput, tmpErr := funcObj(ctx, transImportJobParam)
				if tmpErr != nil {
					updateStatusErr := database.UpdateTransImportDetailStatus(ctx, callParam.TransImportId, transImportJobParam.CurrentDetail.Id, "fail", tmpOutput, tmpErr.Error())
					if updateStatusErr != nil {
						log.Logger.Error("CallImportFunc update detail status fail", log.String("transImport", callParam.TransImportId), log.String("detailId", transImportJobParam.CurrentDetail.Id), log.Error(updateStatusErr))
					}
					break
				} else {
					updateStatusErr := database.UpdateTransImportDetailStatus(ctx, callParam.TransImportId, transImportJobParam.CurrentDetail.Id, "success", tmpOutput, "")
					if updateStatusErr != nil {
						log.Logger.Error("CallImportFunc update detail status fail", log.String("transImport", callParam.TransImportId), log.String("detailId", transImportJobParam.CurrentDetail.Id), log.Error(updateStatusErr))
					}
				}
			}
		}
	}
	return
}

// 1、导入角色
func importRole(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 2、导入cmdb插件服务、导入cmdb数据、同步cmdb数据模型、导入其它插件服务
func importPluginConfig(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 3、导入编排
func importWorkflow(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 4、导入批量执行
func importBatchExecution(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 5、导入物料包
func importArtifactPackage(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 6、导入监控基础类型、对象组、基础类型指标、对象组指标、对象组阈值配置、业务配置模版
func importMonitorBaseConfig(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 7、导入taskman模版和公共组件
func importTaskmanTemplate(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 8、开始执行编排(创建资源、初始化资源、应用部署)
func execWorkflow(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}

// 9、导入监控业务配置、层级对象指标、层级对象阈值配置、自定义看板
func importMonitorServiceConfig(ctx context.Context, transImportParam *models.TransImportJobParam) (output string, err error) {

	return
}
