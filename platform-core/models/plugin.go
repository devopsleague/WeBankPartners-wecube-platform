package models

import "time"

var (
	PluginPackagesStatusMap  = map[int8]string{0: "UNREGISTERED", 1: "REGISTERED", 2: "DECOMMISSIONED"}
	PluginPackagesEditionMap = map[int8]string{0: "community", 1: "enterprise"}
)

type PluginPackages struct {
	Id                string    `json:"id" xorm:"id"`           // 唯一标识
	Name              string    `json:"name" xorm:"name"`       // 显示名
	Version           string    `json:"version" xorm:"version"` // 版本
	Status            int8      `json:"-" xorm:"status"`        // 状态->0(unregistered已上传未注册态)|1(registered注册态)|2(decommissioned注销态)
	StatusString      string    `json:"status" json:"-"`
	UploadTimestamp   time.Time `json:"uploadTimestamp" xorm:"upload_timestamp"`      // 上传时间
	UiPackageIncluded bool      `json:"uiPackageIncluded" xorm:"ui_package_included"` // 是否有ui->0(无)|1(有)
	Edition           int8      `json:"-" xorm:"edition"`                             // 发行版本->0(community社区版)|1(enterprise企业版)
	EditionString     string    `json:"edition" json:"-"`
}

type PluginInstances struct {
	Id                            string `json:"id" xorm:"id"`                                                           // 唯一标识
	Host                          string `json:"host" xorm:"host"`                                                       // 主机ip
	ContainerName                 string `json:"containerName" xorm:"container_name"`                                    // 容器名
	Port                          int    `json:"port" xorm:"port"`                                                       // 服务端口
	ContainerStatus               string `json:"containerStatus" xorm:"container_status"`                                // 容器状态
	PackageId                     string `json:"packageId" xorm:"package_id"`                                            // 插件
	DockerInstanceResourceId      string `json:"dockerInstanceResourceId" xorm:"docker_instance_resource_id"`            // 容器实例id
	InstanceName                  string `json:"instanceName" xorm:"instance_name"`                                      // 容器实例名
	PluginMysqlInstanceResourceId string `json:"pluginMysqlInstanceResourceId" xorm:"plugin_mysql_instance_resource_id"` // 数据库实例id
	S3bucketResourceId            string `json:"s3bucketResourceId" xorm:"s3bucket_resource_id"`                         // s3资源id
}

type PluginPackageRuntimeResourcesDocker struct {
	Id              string `json:"id" xorm:"id"`                             // 唯一标识
	PluginPackageId string `json:"pluginPackageId" xorm:"plugin_package_id"` // 插件
	ImageName       string `json:"imageName" xorm:"image_name"`              // 镜像名
	ContainerName   string `json:"containerName" xorm:"container_name"`      // 容器名
	PortBindings    string `json:"portBindings" xorm:"port_bindings"`        // 端口信息
	VolumeBindings  string `json:"volumeBindings" xorm:"volume_bindings"`    // 目录映射
	EnvVariables    string `json:"envVariables" xorm:"env_variables"`        // 容器环境变量
}

type PluginPackageRuntimeResourcesMysql struct {
	Id              string `json:"id" xorm:"id"`                             // 唯一标识
	PluginPackageId string `json:"pluginPackageId" xorm:"plugin_package_id"` // 插件
	SchemaName      string `json:"schemaName" xorm:"schema_name"`            // 数据库名
	InitFileName    string `json:"initFileName" xorm:"init_file_name"`       // 初始化脚本
	UpgradeFileName string `json:"upgradeFileName" xorm:"upgrade_file_name"` // 升级脚本
}

type PluginPackageRuntimeResourcesS3 struct {
	Id                   string `json:"id" xorm:"id"`                                      // 唯一标识
	PluginPackageId      string `json:"pluginPackageId" xorm:"plugin_package_id"`          // 插件
	BucketName           string `json:"bucketName" xorm:"bucket_name"`                     // 桶名
	AdditionalProperties string `json:"additionalProperties" xorm:"additional_properties"` // 自动上传文件
}

type PluginMysqlInstances struct {
	Id              string    `json:"id" xorm:"id"`                             // 唯一标识
	Password        string    `json:"password" xorm:"password"`                 // 密码
	PluginPackageId string    `json:"pluginPackageId" xorm:"plugin_package_id"` // 插件
	ResourceItemId  string    `json:"resourceItemId" xorm:"resource_item_id"`   // 资源实例id
	SchemaName      string    `json:"schemaName" xorm:"schema_name"`            // 数据库名
	Status          bool      `json:"status" xorm:"status"`                     // 状态->0(inactive)|1(active)
	Username        string    `json:"username" xorm:"username"`                 // 用户名
	PreVersion      string    `json:"preVersion" xorm:"pre_version"`            // 插件版本
	CreatedTime     time.Time `json:"createdTime" xorm:"created_time"`          // 创建时间
	UpdatedTime     time.Time `json:"updatedTime" xorm:"updated_time"`          // 更新时间
}

type PluginPackageAuthorities struct {
	Id              string `json:"id" xorm:"id"`                             // 唯一标识
	PluginPackageId string `json:"pluginPackageId" xorm:"plugin_package_id"` // 插件
	RoleName        string `json:"roleName" xorm:"role_name"`                // 角色
	MenuCode        string `json:"menuCode" xorm:"menu_code"`                // 菜单编码
}

type PluginPackageDependencies struct {
	Id                       string `json:"id" xorm:"id"`                                               // 唯一标识
	PluginPackageId          string `json:"pluginPackageId" xorm:"plugin_package_id"`                   // 插件
	DependencyPackageName    string `json:"dependencyPackageName" xorm:"dependency_package_name"`       // 依赖包名
	DependencyPackageVersion string `json:"dependencyPackageVersion" xorm:"dependency_package_version"` // 依赖包版本
}

type PluginPackageResourceFiles struct {
	Id              string `json:"id" xorm:"id"`                             // 唯一标识
	PluginPackageId string `json:"pluginPackageId" xorm:"plugin_package_id"` // 插件
	PackageName     string `json:"packageName" xorm:"package_name"`          // 插件包名
	PackageVersion  string `json:"packageVersion" xorm:"package_version"`    // 插件版本
	Source          string `json:"source" xorm:"source"`                     // 压缩文件
	RelatedPath     string `json:"relatedPath" xorm:"related_path"`          // 静态文件路径
}

type PluginPackageMenus struct {
	Id               string `json:"id" xorm:"id"`                               // 唯一标识
	PluginPackageId  string `json:"pluginPackageId" xorm:"plugin_package_id"`   // 插件
	Code             string `json:"code" xorm:"code"`                           // 编码
	Category         string `json:"category" xorm:"category"`                   // 目录
	DisplayName      string `json:"displayName" xorm:"display_name"`            // 英文显示名
	LocalDisplayName string `json:"localDisplayName" xorm:"local_display_name"` // 本地语言显示名
	MenuOrder        int    `json:"menuOrder" xorm:"menu_order"`                // 菜单排序
	Path             string `json:"path" xorm:"path"`                           // 前端请求路径
	Active           bool   `json:"active" xorm:"active"`                       // 是否启用->0(不启用)|1(启用)
}