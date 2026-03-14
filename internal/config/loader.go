package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader 配置加载器接口
type Loader interface {
	LoadProjectConfig(path string) (*ProjectConfig, error)
	LoadResourceGroup(path string) (*ResourceGroup, error)
	LoadMapping(path string) (*Mapping, error)
}

// FileLoader 基于文件的配置加载器
type FileLoader struct{}

// NewFileLoader 创建文件配置加载器
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// LoadProjectConfig 加载项目配置
func (l *FileLoader) LoadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	return &config, nil
}

// LoadResourceGroup 加载资源组配置
func (l *FileLoader) LoadResourceGroup(path string) (*ResourceGroup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取资源文件失败：%w", err)
	}

	var resources ResourceGroup
	if err := yaml.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("解析资源文件失败：%w", err)
	}

	return &resources, nil
}

// LoadMapping 加载映射配置
func (l *FileLoader) LoadMapping(path string) (*Mapping, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取映射文件失败：%w", err)
	}

	var mapping Mapping
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("解析映射文件失败：%w", err)
	}

	return &mapping, nil
}

// ResolvePath 解析相对路径为绝对路径
func ResolvePath(basePath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(basePath, relativePath)
}
