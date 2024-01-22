package models

import "time"

type BatchExecution struct {
	Id                       string     `json:"id" xorm:"id"`                                                // 唯一标识
	BatchExecutionTemplateId string     `json:"batchExecutionTemplateId" xorm:"batch_execution_template_id"` // 模板id
	CreatedBy                string     `json:"createdBy" xorm:"created_by"`                                 // 创建者
	UpdatedBy                string     `json:"updatedBy" xorm:"updated_by"`                                 // 更新者
	CreatedTime              *time.Time `json:"createdTime" xorm:"created_time"`                             // 创建时间
	UpdatedTime              *time.Time `json:"updatedTime" xorm:"updated_time"`                             // 更新时间
}

func (BatchExecution) TableName() string {
	return "batch_execution"
}

type BatchExecutionJobs struct {
	Id                      string     `json:"id" xorm:"id"`                                              // 唯一标识
	BatchExecutionId        string     `json:"batchExecutionId" xorm:"batch_execution_id"`                // 批量执行任务id
	PackageName             string     `json:"packageName" xorm:"package_name"`                           // 包名
	EntityName              string     `json:"entityName" xorm:"entity_name"`                             // 实体名
	BusinessKey             string     `json:"businessKey" xorm:"business_key"`                           // 业务key
	RootEntityId            string     `json:"rootEntityId" xorm:"root_entity_id"`                        // 根实体id
	ExecuteTime             *time.Time `json:"executeTime" xorm:"execute_time"`                           // 执行时间
	CompleteTime            *time.Time `json:"completeTime" xorm:"complete_time"`                         // 完成时间
	ErrorCode               string     `json:"errorCode" xorm:"error_code"`                               // 错误码
	ErrorMessage            string     `json:"errorMessage" xorm:"error_message"`                         // 错误信息
	InputJson               string     `json:"inputJson" xorm:"input_json"`                               // 输入json
	ReturnJson              string     `json:"returnJson" xorm:"return_json"`                             // 输出json
	PluginConfigInterfaceId string     `json:"pluginConfigInterfaceId" xorm:"plugin_config_interface_id"` // 插件配置接口id
}

func (BatchExecutionJobs) TableName() string {
	return "batch_execution_jobs"
}

type BatchExecutionTemplate struct {
	Id            string     `json:"id" xorm:"id"`                        // 唯一标识
	Name          string     `json:"name" xorm:"name"`                    // 名称
	Status        string     `json:"status" xorm:"status"`                // 使用状态
	OperateObject string     `json:"operateObject" xorm:"operate_object"` // 操作对象
	PluginService string     `json:"pluginService" xorm:"plugin_service"` // 插件服务
	ConfigData    string     `json:"configData" xorm:"config_data"`       // 配置数据
	CreatedBy     string     `json:"createdBy" xorm:"created_by"`         // 创建者
	CreatedTime   *time.Time `json:"createdTime" xorm:"created_time"`     // 创建时间
	UpdatedBy     string     `json:"updatedBy" xorm:"updated_by"`         // 更新者
	UpdatedTime   *time.Time `json:"updatedTime" xorm:"updated_time"`     // 更新时间
}

func (BatchExecutionTemplate) TableName() string {
	return "batch_execution_template"
}

type BatchExecutionTemplateRole struct {
	Id                       string `json:"id" xorm:"id"`                                                // 唯一标识
	BatchExecutionTemplateId string `json:"batchExecutionTemplateId" xorm:"batch_execution_template_id"` // 批量执行模板id
	Permission               string `json:"permission" xorm:"permission"`                                // 权限类型->MGMT(管理) | USE(使用)
	RoleId                   string `json:"roleId" xorm:"role_id"`                                       // 角色id
	RoleName                 string `json:"roleName" xorm:"role_name"`                                   // 角色名
}

func (BatchExecutionTemplateRole) TableName() string {
	return "batch_execution_template_role"
}

type BatchExecutionTemplateCollect struct {
	Id                       string     `json:"id" xorm:"id"`                                                // 唯一标识
	BatchExecutionTemplateId string     `json:"batchExecutionTemplateId" xorm:"batch_execution_template_id"` // 批量执行模板id
	UserId                   string     `json:"userId" xorm:"user_id"`                                       // 用户id
	CreatedTime              *time.Time `json:"createdTime" xorm:"created_time"`                             // 创建时间
}

func (BatchExecutionTemplateCollect) TableName() string {
	return "batch_execution_template_collect"
}

type BatchExecTemplateInfo struct {
	BatchExecutionTemplate
	PermissionToRole *PermissionToRole `json:"permissionToRole" xorm:"-"`
}

type BatchExecTemplatePageData struct {
	PageInfo PageInfo                 `json:"pageInfo"`
	Contents []*BatchExecTemplateInfo `json:"contents"`
}