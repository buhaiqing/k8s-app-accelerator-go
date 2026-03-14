package config

// ProjectConfig 对应 vars.yaml
type ProjectConfig struct {
	RootDir           string            `yaml:"rootdir"`
	Project           string            `yaml:"project"`
	Profiles          []string          `yaml:"profiles"`
	SSLSecretName     string            `yaml:"ssl_secret_name"`
	Stack             map[string]string `yaml:"stack"`
	ToolsetGitBaseURL string            `yaml:"toolset_git_base_url"`
	ToolsetGitGroup   string            `yaml:"toolset_git_group"`
	ToolsetGitProject string            `yaml:"toolset_git_project"`
	ClusterID         string            `yaml:"cluster_id"`
	JiraID            string            `yaml:"jira_id"`
	HarborProject     string            `yaml:"harbor_project"`

	// Ansible 兼容字段
	DNETProduct   string            `yaml:"DNET_PRODUCT"`
	Namespace     string            `yaml:"namespace"`
	ResourceGroup string            `yaml:"resource_group"`
	AppAuth       map[string]interface{} `yaml:"app_auth"`

	// 嵌套配置
	Apollo  ApolloConfig  `yaml:"apollo"`
	ArgoCD  ArgoCDConfig  `yaml:"argocd"`
	Jenkins JenkinsConfig `yaml:"jenkins"`
}

// ApolloConfig Apollo 配置
type ApolloConfig struct {
	Site       string `yaml:"site"`
	CustomerID string `yaml:"customerid"`
	Env        string `yaml:"env"`
	Alias      string `yaml:"alias"`
	Token      string `yaml:"token"`
}

// ArgoCDConfig ArgoCD 配置
type ArgoCDConfig struct {
	Site    string `yaml:"site"`
	Cluster string `yaml:"cluster"`
}

// JenkinsConfig Jenkins 配置
type JenkinsConfig struct {
	Site string `yaml:"site"`
}
