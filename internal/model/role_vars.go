package model

// RoleVars 角色变量（从 roles/{app}/vars/main.yml 加载）
// 与 Ansible roles/*/vars/main.yml 结构完全一致
type RoleVars struct {
	// 基础信息
	App           string `yaml:"app" json:"app"`
	DNETProduct   string `yaml:"DNET_PRODUCT" json:"dnet_product"`
	HarborProject string `yaml:"harbor_project" json:"harbor_project"`
	Image         string `yaml:"image" json:"image"`
	Type          string `yaml:"_type" json:"_type"`
	Profile       string `yaml:"profile,omitempty" json:"profile"` // 添加 profile 字段

	// 功能开关
	EnableHPA bool `yaml:"enable_hpa" json:"enable_hpa"`
	EnableRDB bool `yaml:"enable_rdb" json:"enable_rdb"`

	// DB 迁移配置
	SetupImage string `yaml:"setup_image" json:"setup_image"`
	SetupDB    string `yaml:"setup_db" json:"setup_db"`

	// 资源配置（按环境）
	Resources RoleResources `yaml:"resources" json:"resources"`

	// 敏感键校验
	SensitiveKeys []string `yaml:"sensitive_keys" json:"sensitive_keys"`

	// 以下字段保留用于向后兼容
	Replicas       int                    `yaml:"replicas" json:"replicas"`
	CPURequests    string                 `yaml:"cpu_requests" json:"cpu_requests"`
	CPULimits      string                 `yaml:"cpu_limits" json:"cpu_limits"`
	MemoryRequests string                 `yaml:"memory_requests" json:"memory_requests"`
	MemoryLimits   string                 `yaml:"memory_limits" json:"memory_limits"`
	ExtraEnv       map[string]interface{} `yaml:"extra_env" json:"extra_env"`
}

// RoleResources 资源配置（与 Ansible vars/main.yml 一致）
type RoleResources struct {
	Default    ResourceConfig `yaml:"default" json:"default"`
	Production ResourceConfig `yaml:"production" json:"production"`
}

// ResourceConfig 资源配置详情
type ResourceConfig struct {
	LimitsCPU      string `yaml:"limits_cpu" json:"limits_cpu"`
	LimitsMemory   string `yaml:"limits_memory" json:"limits_memory"`
	RequestsCPU    string `yaml:"requests_cpu" json:"requests_cpu"`
	RequestsMemory string `yaml:"requests_memory" json:"requests_memory"`
}

// Bootstrap bootstrap.yml 的根结构
type Bootstrap struct {
	Plays []Play `yaml:"plays"`
}

// Play Ansible Play 定义
type Play struct {
	Name       string   `yaml:"name"`
	Hosts      string   `yaml:"hosts"`
	VarsFiles  []string `yaml:"vars_files"`
	Tasks      []Task   `yaml:"tasks"`
	ImportRole string   `yaml:"import_role"`
}

// Task Ansible Task 定义
type Task struct {
	Name       string `yaml:"name"`
	Debug      string `yaml:"debug"`
	Fail       string `yaml:"fail"`
	Assert     Assert `yaml:"assert"`
	ImportRole string `yaml:"import_role"`
	Include    string `yaml:"include"`
}

// Assert Ansible Assert 任务
type Assert struct {
	That string `yaml:"that"`
	Msg  string `yaml:"msg"`
}
