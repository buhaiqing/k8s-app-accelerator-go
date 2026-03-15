package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert/yaml"
)

// GitlabCfgGenerator GitLab 配置生成器
type GitlabCfgGenerator struct {
	projectConfig *config.ProjectConfig
	outputDir     string
	templateDir   string
	workerPool    *template.WorkerPool
	resources     map[string]interface{} // 资源数据
	mapping       map[string]string      // 映射数据
}

// NewGitlabCfgGenerator 创建新的 GitLab 配置生成器
func NewGitlabCfgGenerator(
	projectConfig *config.ProjectConfig,
	outputDir string,
	templateDir string,
	scriptPath string,
	resources map[string]interface{},
	mapping map[string]string,
) (*GitlabCfgGenerator, error) {
	pool, err := template.NewWorkerPool(5, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("创建 worker 池失败：%w", err)
	}

	return &GitlabCfgGenerator{
		projectConfig: projectConfig,
		outputDir:     outputDir,
		templateDir:   templateDir,
		workerPool:    pool,
		resources:     resources,
		mapping:       mapping,
	}, nil
}

// Close 关闭资源
func (g *GitlabCfgGenerator) Close() {
	if g.workerPool != nil {
		g.workerPool.Close()
	}
}

// loadRoleVars 加载 role 的 vars/main.yml
func (g *GitlabCfgGenerator) loadRoleVars(appName string) (map[string]interface{}, error) {
	varsFile := filepath.Join(g.templateDir, appName, "vars", "main.yml")

	data, err := os.ReadFile(varsFile)
	if err != nil {
		return nil, nil
	}

	var roleVars map[string]interface{}
	if err := yaml.Unmarshal(data, &roleVars); err != nil {
		return nil, fmt.Errorf("解析 role vars 失败：%w", err)
	}

	return roleVars, nil
}

// flattenResources 将 resources 中的字段展开为顶层变量
// 模拟 Ansible 的 resources.yml 逻辑：
// 1. 从 resources.yaml 加载资源列表（如 rabbitmq, redis, rds 等）
// 2. 根据 resource_group 配置选择对应的资源（默认 'default'）
// 3. 将选中资源的字段展开为顶层变量
func (g *GitlabCfgGenerator) flattenResources() map[string]interface{} {
	result := make(map[string]interface{})

	// 定义资源类型到变量前缀的映射
	resourceMappings := map[string]struct {
		prefix string
		fields map[string]string
	}{
		"rds": {
			prefix: "",
			fields: map[string]string{
				"datasource_url":      "datasource_url",
				"datasource_port":     "datasource_port",
				"datasource_db":       "datasource_db",
				"datasource_username": "datasource_username",
				"datasource_password": "datasource_password",
			},
		},
		"redis": {
			prefix: "",
			fields: map[string]string{
				"redisIp":       "redisIp",
				"redisPort":     "redisPort",
				"redisDb":       "redisDb",
				"redisPassword": "redisPassword",
			},
		},
		"mongo": {
			prefix: "",
			fields: map[string]string{
				"mongodb_url":      "mongodb_url",
				"mongo_db":         "mongo_db",
				"mongodb_username": "mongodb_username",
				"mongodb_password": "mongodb_password",
			},
		},
		"oss": {
			prefix: "",
			fields: map[string]string{
				"ossEndpoint":         "ossEndpoint",
				"bucketName":          "bucketName",
				"oss_accessKeyId":     "oss_accessKeyId",
				"oss_accessKeySecret": "oss_accessKeySecret",
			},
		},
		"rabbitmq": {
			prefix: "rabbitmq_",
			fields: map[string]string{
				"host":     "host",
				"user":     "user",
				"password": "password",
				"port":     "port",
			},
		},
	}

	// 遍历每种资源类型
	for resourceType, mapping := range resourceMappings {
		if resources, ok := g.resources[resourceType]; ok {
			// 获取资源列表
			var resourceList []interface{}
			switch v := resources.(type) {
			case []interface{}:
				resourceList = v
			}

			// 查找 name == "default" 的资源
			var selectedResource map[string]interface{}
			for _, item := range resourceList {
				if m, ok := item.(map[string]interface{}); ok {
					if name, ok := m["name"].(string); ok && name == "default" {
						selectedResource = m
						break
					}
				}
			}

			// 如果没找到 default，使用第一个
			if selectedResource == nil && len(resourceList) > 0 {
				if m, ok := resourceList[0].(map[string]interface{}); ok {
					selectedResource = m
				}
			}

			// 展开字段
			if selectedResource != nil {
				for fieldName, targetName := range mapping.fields {
					if val, ok := selectedResource[fieldName]; ok {
						key := mapping.prefix + targetName
						result[key] = val
					}
				}
			}
		}
	}

	// 复制其他未处理的资源
	for key, value := range g.resources {
		if _, ok := resourceMappings[key]; !ok {
			result[key] = value
		}
	}

	return result
}

// GenerateAll 生成所有 GitLab 配置
func (g *GitlabCfgGenerator) GenerateAll() error {
	fmt.Printf("开始生成 GitLab 项目配置...\n")

	// 从配置中读取 data 列表
	dataList, err := g.loadGitlabCfgData()
	if err != nil {
		return fmt.Errorf("加载 GitLab 配置数据失败：%w", err)
	}

	fmt.Printf("找到 %d 个应用配置\n", len(dataList))

	// 为每个应用生成配置
	for _, data := range dataList {
		if err := g.generateForProject(data); err != nil {
			return fmt.Errorf("生成 %s 失败：%w", data.AppName, err)
		}
	}

	fmt.Printf("✓ 成功生成 %d 个应用配置\n", len(dataList))
	return nil
}

// GitlabCfgData GitLab 配置数据结构
type GitlabCfgData struct {
	AppName     string
	StackID     string
	CmdbStack   string
	Profile     string
	DNETProduct string
	Project     string
	Namespace   string
	ClusterID   string
	JiraID      string
	Apollo      interface{}
	ArgoCD      interface{}
	AppAuth     interface{}
}

// loadGitlabCfgData 加载 GitLab 配置数据
// 从 vars.yaml 中读取应用列表和配置
func (g *GitlabCfgGenerator) loadGitlabCfgData() ([]GitlabCfgData, error) {
	var dataList []GitlabCfgData

	// ============================================
	// Pre-Check: 验证关键字段
	// ============================================

	// 1. 验证 DNET_PRODUCT
	if g.projectConfig.DNETProduct == "" {
		return nil, fmt.Errorf("❌ 配置错误：DNET_PRODUCT 未定义或为空\n" +
			"   解决方案：请在配置文件中设置 'product' 字段，例如：\n" +
			"   product: cms")
	}

	// 2. 验证 Profiles
	if len(g.projectConfig.Profiles) == 0 {
		return nil, fmt.Errorf("❌ 配置错误：profiles 列表为空\n" +
			"   解决方案：请在配置文件中设置 'profiles' 字段，例如：\n" +
			"   profiles:\n" +
			"     - production\n" +
			"     - int")
	}

	// 3. 验证 Stack
	if len(g.projectConfig.Stack) == 0 {
		return nil, fmt.Errorf("❌ 配置错误：stack 字段为空\n" +
			"   解决方案：请在配置文件中设置 'stack' 字段，例如：\n" +
			"   stack:\n" +
			"     cms-service: baas\n" +
			"     order-service: production")
	}

	// 4. 验证每个应用的 stack 值
	for appName, stackID := range g.projectConfig.Stack {
		if stackID == "" {
			return nil, fmt.Errorf("❌ 配置错误：应用 '%s' 的 stack 值为空\n"+
				"   解决方案：请为应用 '%s' 指定有效的 stack ID", appName, appName)
		}
	}

	// 5. 验证 Project（可选，但建议设置）
	if g.projectConfig.Project == "" {
		fmt.Printf("⚠️  警告：project 未设置，将使用默认值 'default'\n")
	}

	// ============================================
	// 开始加载数据
	// ============================================

	// 构建 apollo 字典（用于模板中的点号访问）
	apolloMap := map[string]interface{}{
		"site":       g.projectConfig.Apollo.Site,
		"customerid": g.projectConfig.Apollo.CustomerID,
		"env":        g.projectConfig.Apollo.Env,
		"alias":      g.projectConfig.Apollo.Alias,
		"token":      g.projectConfig.Apollo.Token,
	}

	// 构建 argocd 字典
	argocdMap := map[string]interface{}{
		"site":    g.projectConfig.ArgoCD.Site,
		"cluster": g.projectConfig.ArgoCD.Cluster,
	}

	// 从 stack 字段读取所有应用
	for appName, stackID := range g.projectConfig.Stack {
		// 遍历所有 profiles
		for _, profile := range g.projectConfig.Profiles {
			// 验证 profile 不为空
			if profile == "" {
				return nil, fmt.Errorf("❌ 配置错误：profiles 列表中包含空值\n" +
					"   解决方案：请检查 profiles 列表，移除空字符串")
			}

			data := GitlabCfgData{
				AppName:     appName,
				StackID:     stackID,
				CmdbStack:   stackID, // 默认使用 stackID
				Profile:     profile,
				DNETProduct: g.projectConfig.DNETProduct,
				Project:     g.projectConfig.Project,
				Namespace:   g.projectConfig.Namespace,
				ClusterID:   g.projectConfig.ClusterID,
				JiraID:      g.projectConfig.JiraID,
				Apollo:      apolloMap,
				ArgoCD:      argocdMap,
				AppAuth:     g.projectConfig.AppAuth,
			}
			dataList = append(dataList, data)
		}
	}

	if len(dataList) == 0 {
		return nil, fmt.Errorf("❌ 配置错误：没有生成任何应用配置\n" +
			"   可能原因：\n" +
			"   1. stack 字段为空\n" +
			"   2. profiles 列表为空\n" +
			"   解决方案：请检查配置文件中的 'stack' 和 'profiles' 字段")
	}

	return dataList, nil
}

// generateForProject 为单个项目生成配置
func (g *GitlabCfgGenerator) generateForProject(data GitlabCfgData) error {
	fmt.Printf("  - 生成 %s (%s)\n", data.AppName, data.StackID)

	// 创建输出目录：{outputDir}/{DNET_PRODUCT}/{app}/overlays/{profile}/{cmdb_stack}
	// 对应 Ansible: {outputDir}/{DNET_PRODUCT}/{app}/overlays/{profile}/{cmdb_stack}
	outputSubDir := filepath.Join(g.outputDir, data.DNETProduct, data.AppName, "overlays", data.Profile, data.CmdbStack)
	if err := os.MkdirAll(outputSubDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}

	// 定义要生成的配置文件列表（基础文件，总是生成）
	baseConfigFiles := []string{
		"cm-suffix-transformer.yaml",
		"config.yaml",
		"deployment.yaml",
		"dev.env",
		"kustomization.yaml",
		"ops.env",
		"ops.j2",
		"service.yaml",
	}

	// 加载 role vars 以检查条件变量
	roleVars, err := g.loadRoleVars(data.AppName)
	if err != nil {
		return fmt.Errorf("加载 role vars 失败：%w", err)
	}

	// 检查是否启用 HPA
	enableHpa := false
	if v, ok := roleVars["enable_hpa"]; ok {
		if b, ok := v.(bool); ok {
			enableHpa = b
		}
	}

	// 检查是否启用 RDB 和应用类型
	enableRdb := false
	appType := ""
	if v, ok := roleVars["enable_rdb"]; ok {
		if b, ok := v.(bool); ok {
			enableRdb = b
		}
	}
	if v, ok := roleVars["_type"]; ok {
		if s, ok := v.(string); ok {
			appType = s
		}
	}

	// 为每个基础配置文件生成内容
	for _, configFile := range baseConfigFiles {
		if err := g.generateConfigFile(data, outputSubDir, configFile); err != nil {
			return fmt.Errorf("生成 %s 失败：%w", configFile, err)
		}
	}

	// 条件生成 hpa.yaml
	if enableHpa {
		if err := g.generateConfigFile(data, outputSubDir, "hpa.yaml"); err != nil {
			return fmt.Errorf("生成 hpa.yaml 失败：%w", err)
		}
	}

	// 条件生成 job.yaml（当 enable_rdb == true 且 _type == "backend"）
	if enableRdb && appType == "backend" {
		if err := g.generateConfigFile(data, outputSubDir, "job.yaml"); err != nil {
			return fmt.Errorf("生成 job.yaml 失败：%w", err)
		}
	}

	// 生成 ArgoCD Application 配置
	if err := g.generateArgoCDConfig(data); err != nil {
		return fmt.Errorf("生成 ArgoCD 配置失败：%w", err)
	}

	// 生成 Jenkins Job 配置
	if err := g.generateJenkinsConfig(data); err != nil {
		return fmt.Errorf("生成 Jenkins 配置失败：%w", err)
	}

	fmt.Printf("    ✅ 已生成：%s\n", outputSubDir)
	return nil
}

// generateConfigFile 生成单个配置文件
func (g *GitlabCfgGenerator) generateConfigFile(data GitlabCfgData, outputDir string, configFile string) error {
	// 模板路径：直接使用 Ansible roles 中的模板
	// Ansible 使用 _path: "production" 来指定模板路径
	// 模板路径：{baseDir}/roles/{app}/templates/overlays/production/production/{config}.j2
	// 输出路径：{outputDir}/{DNET_PRODUCT}/{app}/overlays/{profile}/{cmdb_stack}/{config}
	templatePath := filepath.Join(g.templateDir, data.AppName, "templates", "overlays", "production", "production", configFile+".j2")

	// 转换为绝对路径
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("获取模板绝对路径失败：%w", err)
	}

	fmt.Printf("    模板路径：%s\n", absTemplatePath)

	// 检查模板是否存在
	if _, err := os.Stat(absTemplatePath); os.IsNotExist(err) {
		// 如果模板不存在，跳过该文件
		fmt.Printf("      ⚠️  模板不存在，跳过：%s\n", absTemplatePath)
		return nil
	}

	// 确定 harbor_project 值
	harborProject := ""
	if hp, ok := g.resources["harbor_project"].(string); ok && hp != "" {
		harborProject = hp
	} else {
		harborProject = data.DNETProduct
	}

	// 加载 role 的 vars/main.yml
	roleVars, err := g.loadRoleVars(data.AppName)
	if err != nil {
		return fmt.Errorf("加载 role vars 失败：%w", err)
	}

	// 展开 resources 中的字段为顶层变量
	flattenedResources := g.flattenResources()

	// 构建渲染上下文（先设置基础变量）
	// 构建 jenkins_defs（对应 Ansible overlays.yml 中的 set_fact）
	profileConvert := func(profile string) string {
		switch profile {
		case "int":
			return "Int"
		case "uat":
			return "Uat"
		case "branch":
			return "BRA"
		case "production":
			return "PRD"
		default:
			return profile
		}
	}
	dnetProductUpper := strings.ToUpper(data.DNETProduct)
	profileConverted := profileConvert(data.Profile)
	jenkinsDefs := map[string]string{
		"appinstall":     fmt.Sprintf("%s_%s_K8s/%s_deploy_no_rdb_%s", dnetProductUpper, profileConverted, dnetProductUpper, profileConverted),
		"rdb_appinstall": fmt.Sprintf("%s_%s_K8s/%s_deploy_%s", dnetProductUpper, profileConverted, dnetProductUpper, profileConverted),
		"rdb":            fmt.Sprintf("%s_%s_K8s/%s_deploy_rdb_%s", dnetProductUpper, profileConverted, dnetProductUpper, profileConverted),
	}

	// 生成 random_value（对应 Ansible: 10000 | random | to_uuid | truncate(7,True,'')）
	// 使用 UUID 的前 7 位作为随机值
	randomUUID := uuid.New().String()
	randomValue := strings.ReplaceAll(randomUUID, "-", "")[:7]

	// 从 role vars 获取 setup_image 和 setup_db
	// 注意：setup_image 可能是 Jinja2 模板，需要先渲染
	setupImage := ""
	setupDb := ""
	if v, ok := roleVars["setup_image"]; ok {
		if s, ok := v.(string); ok {
			// 渲染 setup_image 中的 Jinja2 模板变量
			setupImage = strings.ReplaceAll(s, "{{harbor_project}}", harborProject)
			setupImage = strings.ReplaceAll(setupImage, "{{app}}", data.AppName)
		}
	}
	if v, ok := roleVars["setup_db"]; ok {
		if s, ok := v.(string); ok {
			setupDb = s
		}
	}

	ctx := map[string]interface{}{
		"app":            data.AppName,
		"stackId":        data.StackID,
		"cmdb_stack":     data.CmdbStack,
		"profile":        data.Profile,
		"DNET_PRODUCT":   data.DNETProduct,
		"project":        data.Project,
		"namespace":      data.Namespace,
		"cluster_id":     data.ClusterID,
		"jira_id":        data.JiraID,
		"apollo":         data.Apollo,
		"argocd":         data.ArgoCD,
		"app_auth":       data.AppAuth,
		"rootdir":        g.outputDir,
		"harbor_project": harborProject,
		"resources":      g.resources,
		"mappings":       g.mapping,
		"jenkins": map[string]interface{}{
			"site": g.projectConfig.Jenkins.Site,
		},
		"jenkins_defs": jenkinsDefs,
		"random_value": randomValue,
		"setup_image":  setupImage,
		"setup_db":     setupDb,
		"minReplicas":  2,
		"maxReplicas":  2,
	}

	// 合并展开的 resources 变量
	for k, v := range flattenedResources {
		ctx[k] = v
	}

	// 合并 role vars（role vars 中的值优先级更高）
	// 但对于 setup_image 和 setup_db，如果已预渲染则保留预渲染的值
	for k, v := range roleVars {
		ctx[k] = v
	}
	// 恢复预渲染的 setup_image 和 setup_db 值
	if setupImage != "" {
		ctx["setup_image"] = setupImage
	}
	if setupDb != "" {
		ctx["setup_db"] = setupDb
	}

	req := template.RenderRequest{
		TemplatePath: absTemplatePath,
		Context:      ctx,
	}

	content, err := g.workerPool.Render(req)
	if err != nil {
		return fmt.Errorf("渲染模板失败：%w", err)
	}

	// 写入文件（确保末尾有换行符）
	outputPath := filepath.Join(outputDir, configFile)
	if err := g.writeFileWithNewline(outputPath, content); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	return nil
}

// writeFileWithNewline 写入文件并确保末尾有换行符
func (g *GitlabCfgGenerator) writeFileWithNewline(path string, content string) error {
	// 确保内容末尾有换行符
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// generateArgoCDConfig 生成 ArgoCD Application 配置
func (g *GitlabCfgGenerator) generateArgoCDConfig(data GitlabCfgData) error {
	fmt.Printf("    生成 ArgoCD Application 配置...\n")

	// 创建输出目录：{outputDir}/argo-app/{project}/{profile}/{k8s_product}
	// 对应 Ansible: {outputDir}/argo-app/{project}/{profile}/k8s_{DNET_PRODUCT}
	k8sProduct := "k8s_" + data.DNETProduct
	outputSubDir := filepath.Join(g.outputDir, "argo-app", data.Project, data.Profile, k8sProduct)
	if err := os.MkdirAll(outputSubDir, 0755); err != nil {
		return fmt.Errorf("创建 ArgoCD 输出目录失败：%w", err)
	}

	// 模板路径：{templateDir}/argo-app/templates/app.yaml.j2
	// templateDir = {baseDir}/roles
	// 所以完整路径是：{baseDir}/roles/argo-app/templates/app.yaml.j2
	templatePath := filepath.Join(g.templateDir, "argo-app", "templates", "app.yaml.j2")

	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("获取模板绝对路径失败：%w", err)
	}

	if _, err := os.Stat(absTemplatePath); os.IsNotExist(err) {
		fmt.Printf("      ⚠️  ArgoCD 模板不存在，跳过：%s\n", absTemplatePath)
		return nil
	}

	// 确定 harbor_project 值
	harborProject := ""
	if hp, ok := g.resources["harbor_project"].(string); ok && hp != "" {
		harborProject = hp
	} else {
		harborProject = data.DNETProduct
	}

	// 加载 role 的 vars/main.yml
	roleVars, err := g.loadRoleVars(data.AppName)
	if err != nil {
		return fmt.Errorf("加载 role vars 失败：%w", err)
	}

	// 展开 resources 中的字段为顶层变量
	flattenedResources := g.flattenResources()

	ctx := map[string]interface{}{
		"app":            data.AppName,
		"stackId":        data.StackID,
		"cmdb_stack":     data.CmdbStack,
		"profile":        data.Profile,
		"DNET_PRODUCT":   data.DNETProduct,
		"project":        data.Project,
		"namespace":      data.Namespace,
		"cluster_id":     data.ClusterID,
		"jira_id":        data.JiraID,
		"apollo":         data.Apollo,
		"argocd":         data.ArgoCD,
		"app_auth":       data.AppAuth,
		"rootdir":        g.outputDir,
		"resources":      g.resources,
		"mappings":       g.mapping,
		"harbor_project": harborProject,
		// ArgoCD 特有变量
		"app_name": data.AppName,
		"stack_id": data.StackID,
		// Ansible 兼容变量（用于模板中的变量引用）
		"toolset_git_base_url": g.projectConfig.ToolsetGitBaseURL,
		"toolset_git_group":    g.projectConfig.ToolsetGitGroup,
		"toolset_git_project":  g.projectConfig.ToolsetGitProject,
		// 注意：不主动设置 k8s_apiserver，让模板使用 default filter
	}

	// 合并展开的 resources 变量
	for k, v := range flattenedResources {
		ctx[k] = v
	}

	// 合并 role vars（role vars 中的值优先级更高）
	for k, v := range roleVars {
		ctx[k] = v
	}

	req := template.RenderRequest{
		TemplatePath: absTemplatePath,
		Context:      ctx,
	}

	content, err := g.workerPool.Render(req)
	if err != nil {
		return fmt.Errorf("渲染 ArgoCD 模板失败：%w", err)
	}

	// 输出文件名：{DNET_PRODUCT}-{app_name}.yaml
	// 对应 Ansible: cms-cms-service.yaml（使用 DNET_PRODUCT 而非 project）
	outputFileName := fmt.Sprintf("%s-%s.yaml", data.DNETProduct, data.AppName)
	outputPath := filepath.Join(outputSubDir, outputFileName)
	if err := g.writeFileWithNewline(outputPath, content); err != nil {
		return fmt.Errorf("写入 ArgoCD 文件失败：%w", err)
	}

	return nil
}

// generateJenkinsConfig 生成 Jenkins Job 配置
func (g *GitlabCfgGenerator) generateJenkinsConfig(data GitlabCfgData) error {
	fmt.Printf("    生成 Jenkins Job 配置...\n")

	// 创建输出目录：{outputDir}/jenkins-job/{harbor_project}
	// 对应 Ansible: {outputDir}/jenkins-job/{harbor_project}
	// 需要从 resources 中获取 harbor_project 值
	harborProject := ""
	if hp, ok := g.resources["harbor_project"].(string); ok {
		harborProject = hp
	} else {
		// 如果没有设置，使用 DNET_PRODUCT 作为默认值
		harborProject = data.DNETProduct
	}

	outputSubDir := filepath.Join(g.outputDir, "jenkins-job", harborProject)
	if err := os.MkdirAll(outputSubDir, 0755); err != nil {
		return fmt.Errorf("创建 Jenkins 输出目录失败：%w", err)
	}

	// 模板路径：{templateDir}/jenkins-job/templates/job.j2
	// templateDir = {baseDir}/roles
	// 所以完整路径是：{baseDir}/roles/jenkins-job/templates/job.j2
	templatePath := filepath.Join(g.templateDir, "jenkins-job", "templates", "job.j2")

	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("获取模板绝对路径失败：%w", err)
	}

	if _, err := os.Stat(absTemplatePath); os.IsNotExist(err) {
		fmt.Printf("      ⚠️  Jenkins 模板不存在，跳过：%s\n", absTemplatePath)
		return nil
	}

	// 加载 role 的 vars/main.yml
	roleVars, err := g.loadRoleVars(data.AppName)
	if err != nil {
		return fmt.Errorf("加载 role vars 失败：%w", err)
	}

	// 展开 resources 中的字段为顶层变量
	flattenedResources := g.flattenResources()

	// Jenkins 特有变量（对应 Ansible jenkins-job/tasks/main.yml）
	profileConvert := func(profile string) string {
		switch profile {
		case "int":
			return "Int"
		case "uat":
			return "Uat"
		case "branch":
			return "BRA"
		case "production":
			return "PRD"
		default:
			return profile
		}
	}
	surfix := profileConvert(data.Profile)
	prefix := strings.ToUpper(data.DNETProduct)
	jjbProjName := fmt.Sprintf("%s_%s_k8s", data.DNETProduct, surfix)
	folder := fmt.Sprintf("%s_%s_K8s", prefix, surfix)

	ctx := map[string]interface{}{
		"app":            data.AppName,
		"stackId":        data.StackID,
		"cmdb_stack":     data.CmdbStack,
		"profile":        data.Profile,
		"DNET_PRODUCT":   data.DNETProduct,
		"project":        data.Project,
		"namespace":      data.Namespace,
		"cluster_id":     data.ClusterID,
		"jira_id":        data.JiraID,
		"apollo":         data.Apollo,
		"argocd":         data.ArgoCD,
		"app_auth":       data.AppAuth,
		"rootdir":        g.outputDir,
		"resources":      g.resources,
		"mappings":       g.mapping,
		"harbor_project": harborProject,
		// Jenkins 特有变量
		"app_name":             data.AppName,
		"stack_id":             data.StackID,
		"surfix":               surfix,
		"prefix":               prefix,
		"jjb_proj_name":        jjbProjName,
		"folder":               folder,
		"customer":             g.projectConfig.Project,
		"toolset_git_base_url": g.projectConfig.ToolsetGitBaseURL,
		"toolset_git_group":    g.projectConfig.ToolsetGitGroup,
		"toolset_git_project":  g.projectConfig.ToolsetGitProject,
		"credentials_id":       fmt.Sprintf("%s-git-credential", data.DNETProduct),
	}

	// 合并展开的 resources 变量
	for k, v := range flattenedResources {
		ctx[k] = v
	}

	// 合并 role vars（role vars 中的值优先级更高）
	for k, v := range roleVars {
		ctx[k] = v
	}

	req := template.RenderRequest{
		TemplatePath: absTemplatePath,
		Context:      ctx,
	}

	content, err := g.workerPool.Render(req)
	if err != nil {
		return fmt.Errorf("渲染 Jenkins 模板失败：%w", err)
	}

	outputPath := filepath.Join(outputSubDir, "project.yml")
	if err := g.writeFileWithNewline(outputPath, content); err != nil {
		return fmt.Errorf("写入 Jenkins 文件失败：%w", err)
	}

	return nil
}
