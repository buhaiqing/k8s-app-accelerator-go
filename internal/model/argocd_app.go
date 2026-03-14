package model

// ArgoCDApplication ArgoCD Application 数据结构
type ArgoCDApplication struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata 元数据
type Metadata struct {
	Name       string   `yaml:"name"`
	Namespace  string   `yaml:"namespace"`
	Finalizers []string `yaml:"finalizers,omitempty"`
	Labels     Labels   `yaml:"labels"`
}

// Labels 标签
type Labels struct {
	Project string `yaml:"project"`
	Profile string `yaml:"profile"`
	Stack   string `yaml:"stack"`
	App     string `yaml:"app"`
}

// Spec 规格说明
type Spec struct {
	Destination Destination `yaml:"destination"`
	Source      Source      `yaml:"source"`
	Project     string      `yaml:"project"`
	SyncPolicy  SyncPolicy  `yaml:"syncPolicy"`
}

// Destination 目标集群
type Destination struct {
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace"`
	Server    string `yaml:"server"`
}

// Source 源码仓库
type Source struct {
	Path           string    `yaml:"path"`
	RepoURL        string    `yaml:"repoURL"`
	TargetRevision string    `yaml:"targetRevision"`
	Kustomize      Kustomize `yaml:"kustomize,omitempty"`
}

// Kustomize Kustomize 配置
type Kustomize struct {
	Version string `yaml:"version"`
}

// SyncPolicy 同步策略
type SyncPolicy struct {
	SyncOptions []string `yaml:"syncOptions,omitempty"`
}
