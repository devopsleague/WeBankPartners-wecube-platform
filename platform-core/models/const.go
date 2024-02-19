package models

import "fmt"

const (
	GlobalProjectName = "platform"
	DateTimeFormat    = "2006-01-02 15:04:05"

	// header key
	AuthorizationHeader    = "Authorization"
	TransactionIdHeader    = "transactionId"
	RequestIdHeader        = "requestId"
	DefaultHttpErrorCode   = "ERROR"
	DefaultHttpSuccessCode = "OK"
	DefaultHttpConfirmCode = "CONFIRM"
	ContinueToken          = "continueToken"
	// context key
	ContextRequestBody  = "requestBody"
	ContextResponseBody = "responseBody"
	ContextOperator     = "operator"
	ContextRoles        = "roles"
	ContextAuth         = "auth"
	ContextAuthorities  = "authorities"
	ContextErrorCode    = "errorCode"
	ContextErrorKey     = "errorKey"
	ContextErrorMessage = "errorMessage"
	ContextUserId       = "userId"

	JwtSignKey = "authJwtSecretKey"
	AESPrefix  = "{AES}"

	// table name
	TableNameBatchExec                = "batch_execution"
	TableNameBatchExecJobs            = "batch_exec_jobs"
	TableNameBatchExecTemplate        = "batch_execution_template"
	TableNameBatchExecTemplateRole    = "batch_execution_template_role"
	TableNameBatchExecTemplateCollect = "batch_execution_template_collect"
	TableNamePluginConfigRoles        = "plugin_config_roles"
	TableNamePluginConfigs            = "plugin_configs"

	// batch execution
	BatchExecTemplateStatusAvailable    = "available"
	BatchExecTemplateStatusUnauthorized = "unauthorized"
	BatchExecTmplPublishStatusDraft     = "draft"
	BatchExecTmplPublishStatusPublished = "published"
	BatchExecErrorCodeSucceed           = "0"
	BatchExecErrorCodeFailed            = "1"
	BatchExecErrorCodePending           = "2"
	BatchExecErrorCodeDangerousBlock    = "3"

	DefaultKeepBatchExecDays = 365

	// permission type
	PermissionTypeMGMT = "MGMT"
	PermissionTypeUSE  = "USE"
)

var (
	UrlPrefix = fmt.Sprintf("/%s", GlobalProjectName)
)
