package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
)

func TestLoadProjectConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "vars.yaml")

	configContent := `
project: test-project
profiles:
  - int
  - production
apollo:
  site: https://test.example.com
  token: test-token
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// 测试加载
	loader := config.NewFileLoader()
	cfg, err := loader.LoadProjectConfig(configPath)

	assert.NoError(t, err)
	assert.Equal(t, "test-project", cfg.Project)
	assert.Len(t, cfg.Profiles, 2)
	assert.Equal(t, "https://test.example.com", cfg.Apollo.Site)
}

func TestLoadResourceGroup(t *testing.T) {
	tmpDir := t.TempDir()
	resourcesPath := filepath.Join(tmpDir, "resources.yaml")

	resourcesContent := `
rds:
  - name: default
    datasource_url: rm-test.mysql.rds.aliyuncs.com
    datasource_db: testdb
redis:
  - name: default
    redisIp: r-test.redis.rds.aliyuncs.com
    redisPort: "6379"
`
	err := os.WriteFile(resourcesPath, []byte(resourcesContent), 0644)
	assert.NoError(t, err)

	loader := config.NewFileLoader()
	resources, err := loader.LoadResourceGroup(resourcesPath)

	assert.NoError(t, err)
	assert.Len(t, resources.RDS, 1)
	assert.Len(t, resources.Redis, 1)
	assert.Equal(t, "rm-test.mysql.rds.aliyuncs.com", resources.RDS[0].DatasourceURL)
}

func TestLoadMapping(t *testing.T) {
	tmpDir := t.TempDir()
	mappingPath := filepath.Join(tmpDir, "mapping.yaml")

	mappingContent := `
mappings:
  cms-service: cms
  fms-service: fms
`
	err := os.WriteFile(mappingPath, []byte(mappingContent), 0644)
	assert.NoError(t, err)

	loader := config.NewFileLoader()
	mapping, err := loader.LoadMapping(mappingPath)

	assert.NoError(t, err)
	assert.Equal(t, "cms", mapping.Mappings["cms-service"])
	assert.Equal(t, "fms", mapping.Mappings["fms-service"])
}
