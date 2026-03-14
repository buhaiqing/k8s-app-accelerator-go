# K8s App Accelerator Go - Makefile
# 支持跨平台编译、测试、打包和部署

# ==============================================================================
# 变量定义
# ==============================================================================

# 项目信息
PROJECT_NAME := k8s-app-accelerator-go
BINARY_NAME := k8s-gen
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关
GO := go
GO_VERSION := 1.25.0
GO_LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目录
CMD_DIR := cmd
INTERNAL_DIR := internal
SCRIPTS_DIR := scripts
BUILD_DIR := build
DIST_DIR := dist
OUTPUT_DIR := output

# Python 相关
PYTHON := python3
PYTHON_SCRIPT := $(SCRIPTS_DIR)/render_worker.py
PYTHON_REQUIREMENTS := $(SCRIPTS_DIR)/requirements.txt

# 配置文件示例
CONFIG_EXAMPLES := configs/vars.example.yaml configs/resources.example.yaml configs/mapping.example.yaml

# 平台支持
PLATFORMS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64

# 颜色输出（仅在不支持颜色的终端禁用）
COLOR_RESET := \033[0m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_RED := \033[31m

# 默认目标
.DEFAULT_GOAL := help

# ==============================================================================
# 主要目标
# ==============================================================================

.PHONY: all
all: deps test build ## 完整构建：依赖 + 测试 + 构建

.PHONY: help
help: ## 显示帮助信息
	@echo "$(COLOR_BLUE)K8s App Accelerator Go - Makefile 帮助$(COLOR_RESET)"
	@echo ""
	@echo "用法：make [目标]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

# ==============================================================================
# 依赖管理
# ==============================================================================

.PHONY: deps
deps: ## 安装 Go 依赖
	@echo "$(COLOR_BLUE)[Go] 安装依赖...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Go 依赖安装完成$(COLOR_RESET)"

.PHONY: deps-python
deps-python: ## 安装 Python 依赖
	@echo "$(COLOR_BLUE)[Python] 安装依赖...$(COLOR_RESET)"
	@if [ -f $(PYTHON_REQUIREMENTS) ]; then \
		$(PYTHON) -m pip install -r $(PYTHON_REQUIREMENTS); \
		echo "$(COLOR_GREEN)✓ Python 依赖安装完成$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ requirements.txt 不存在，跳过$(COLOR_RESET)"; \
	fi

.PHONY: deps-all
deps-all: deps deps-python ## 安装所有依赖（Go + Python）

# ==============================================================================
# 代码质量
# ==============================================================================

.PHONY: fmt
fmt: ## 格式化 Go 代码
	@echo "$(COLOR_BLUE)[Go] 格式化代码...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ 代码格式化完成$(COLOR_RESET)"

.PHONY: vet
vet: ## 运行 Go vet 检查
	@echo "$(COLOR_BLUE)[Go] 运行 vet 检查...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ Vet 检查完成$(COLOR_RESET)"

.PHONY: lint
lint: fmt vet ## 代码检查（格式化 + vet）

# ==============================================================================
# 测试
# ==============================================================================

.PHONY: test
test: ## 运行 Go 测试
	@echo "$(COLOR_BLUE)[Test] 运行 Go 测试...$(COLOR_RESET)"
	@$(GO) test -v -race -cover ./...
	@echo "$(COLOR_GREEN)✓ 测试完成$(COLOR_RESET)"

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	@echo "$(COLOR_BLUE)[Test] 运行测试并生成覆盖率报告...$(COLOR_RESET)"
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ 覆盖率报告已生成：coverage.html$(COLOR_RESET)"

.PHONY: test-clean
test-clean: ## 清理测试缓存
	@echo "$(COLOR_BLUE)[Test] 清理测试缓存...$(COLOR_RESET)"
	@$(GO) clean -testcache
	@echo "$(COLOR_GREEN)✓ 测试缓存已清理$(COLOR_RESET)"

.PHONY: test-python
test-python: ## 运行 Python 测试
	@echo "$(COLOR_BLUE)[Python] 运行 Filters 测试...$(COLOR_RESET)"
	@cd $(SCRIPTS_DIR) && $(PYTHON) filters.py
	@echo "$(COLOR_GREEN)✓ Python 测试完成$(COLOR_RESET)"

.PHONY: test-all
test-all: test test-python ## 运行所有测试（Go + Python）

# ==============================================================================
# 构建
# ==============================================================================

.PHONY: build
build: ## 构建当前平台的二进制文件
	@echo "$(COLOR_BLUE)[Build] 构建 $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "$(COLOR_GREEN)✓ 构建完成：$(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-debug
build-debug: ## 构建调试版本（包含调试符号）
	@echo "$(COLOR_BLUE)[Build] 构建调试版本...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -gcflags "all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug main.go
	@echo "$(COLOR_GREEN)✓ 调试版本构建完成：$(BUILD_DIR)/$(BINARY_NAME)-debug$(COLOR_RESET)"

# ==============================================================================
# 跨平台编译
# ==============================================================================

.PHONY: build-all
build-all: $(PLATFORMS) ## 构建所有支持的平台

# macOS AMD64
.PHONY: darwin-amd64
darwin-amd64: ## 构建 macOS AMD64 版本
	@echo "$(COLOR_BLUE)[Build] 构建 darwin-amd64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GO) build $(GO_LDFLAGS) -o $(DIST_DIR)/darwin-amd64/$(BINARY_NAME) main.go
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/darwin-amd64/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/darwin-amd64/configs/ 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ darwin-amd64 构建完成$(COLOR_RESET)"

# macOS ARM64 (Apple Silicon)
.PHONY: darwin-arm64
darwin-arm64: ## 构建 macOS ARM64 版本
	@echo "$(COLOR_BLUE)[Build] 构建 darwin-arm64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GO) build $(GO_LDFLAGS) -o $(DIST_DIR)/darwin-arm64/$(BINARY_NAME) main.go
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/darwin-arm64/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/darwin-arm64/configs/ 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ darwin-arm64 构建完成$(COLOR_RESET)"

# Linux AMD64
.PHONY: linux-amd64
linux-amd64: ## 构建 Linux AMD64 版本
	@echo "$(COLOR_BLUE)[Build] 构建 linux-amd64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 $(GO) build $(GO_LDFLAGS) -o $(DIST_DIR)/linux-amd64/$(BINARY_NAME) main.go
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/linux-amd64/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/linux-amd64/configs/ 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ linux-amd64 构建完成$(COLOR_RESET)"

# Linux ARM64
.PHONY: linux-arm64
linux-arm64: ## 构建 Linux ARM64 版本
	@echo "$(COLOR_BLUE)[Build] 构建 linux-arm64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/linux-arm64
	GOOS=linux GOARCH=arm64 $(GO) build $(GO_LDFLAGS) -o $(DIST_DIR)/linux-arm64/$(BINARY_NAME) main.go
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/linux-arm64/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/linux-arm64/configs/ 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ linux-arm64 构建完成$(COLOR_RESET)"

# Windows AMD64
.PHONY: windows-amd64
windows-amd64: ## 构建 Windows AMD64 版本
	@echo "$(COLOR_BLUE)[Build] 构建 windows-amd64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 $(GO) build $(GO_LDFLAGS) -o $(DIST_DIR)/windows-amd64/$(BINARY_NAME).exe main.go
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/windows-amd64/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/windows-amd64/configs/ 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ windows-amd64 构建完成$(COLOR_RESET)"

# ==============================================================================
# 打包发布
# ==============================================================================

.PHONY: package
package: build-all ## 打包所有平台
	@echo "$(COLOR_BLUE)[Package] 打包发布文件...$(COLOR_RESET)"
	@for platform in $(PLATFORMS); do \
		cd $(DIST_DIR) && \
		if [ "$$platform" = "windows-amd64" ]; then \
			zip -r $(PROJECT_NAME)-$$platform.zip $$platform/; \
		else \
			tar -czvf $(PROJECT_NAME)-$$platform.tar.gz $$platform/; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✓ 打包完成：$(DIST_DIR)/$(COLOR_RESET)"
	@ls -lh $(DIST_DIR)/

.PHONY: package-single
package-single: build ## 打包当前平台
	@echo "$(COLOR_BLUE)[Package] 打包当前平台...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)
	@cp -r $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/
	@cp -r $(SCRIPTS_DIR) $(DIST_DIR)/
	@cp -r $(CONFIG_EXAMPLES) $(DIST_DIR)/configs/ 2>/dev/null || true
	@cd $(DIST_DIR) && tar -czvf $(PROJECT_NAME)-latest.tar.gz $(BINARY_NAME) $(SCRIPTS_DIR) configs/
	@echo "$(COLOR_GREEN)✓ 打包完成：$(DIST_DIR)/$(PROJECT_NAME)-latest.tar.gz$(COLOR_RESET)"

# ==============================================================================
# 清理
# ==============================================================================

.PHONY: clean
clean: ## 清理构建产物
	@echo "$(COLOR_BLUE)[Clean] 清理构建产物...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(OUTPUT_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ 清理完成$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## 完全清理（包括依赖）
	@echo "$(COLOR_BLUE)[Clean] 完全清理...$(COLOR_RESET)"
	@rm -rf vendor/
	@echo "$(COLOR_GREEN)✓ 完全清理完成$(COLOR_RESET)"

# ==============================================================================
# 运行
# ==============================================================================

.PHONY: run
run: build ## 运行程序
	@echo "$(COLOR_BLUE)[Run] 运行程序...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) --help

.PHONY: run-precheck
run-precheck: build ## 运行预检功能
	@echo "$(COLOR_BLUE)[Run] 运行预检...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) precheck --config configs/vars.yaml

.PHONY: run-generate
run-generate: build ## 运行生成功能
	@echo "$(COLOR_BLUE)[Run] 运行生成...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) generate --config configs/vars.yaml

.PHONY: dev
dev: ## 开发模式运行（go run）
	@echo "$(COLOR_BLUE)[Dev] 开发模式运行...$(COLOR_RESET)"
	@$(GO) run main.go --help

# ==============================================================================
# Docker 相关
# ==============================================================================

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "$(COLOR_BLUE)[Docker] 构建镜像...$(COLOR_RESET)"
	docker build -t $(PROJECT_NAME):$(VERSION) .
	@echo "$(COLOR_GREEN)✓ Docker 镜像构建完成：$(PROJECT_NAME):$(VERSION)$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "$(COLOR_BLUE)[Docker] 运行容器...$(COLOR_RESET)"
	docker run --rm -v $(PWD):/app $(PROJECT_NAME):$(VERSION) --help

.PHONY: docker-clean
docker-clean: ## 清理 Docker 资源
	@echo "$(COLOR_BLUE)[Docker] 清理资源...$(COLOR_RESET)"
	docker rmi $(PROJECT_NAME):$(VERSION) 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ Docker 资源清理完成$(COLOR_RESET)"

# ==============================================================================
# 文档
# ==============================================================================

.PHONY: docs
docs: ## 生成文档
	@echo "$(COLOR_BLUE)[Docs] 生成文档...$(COLOR_RESET)"
	@godoc -http=:6060 &
	@echo "$(COLOR_GREEN)✓ 文档服务器已启动：http://localhost:6060$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)提示：按 Ctrl+C 停止文档服务器$(COLOR_RESET)"

.PHONY: readme-check
readme-check: ## 检查 README 完整性
	@echo "$(COLOR_BLUE)[Docs] 检查 README...$(COLOR_RESET)"
	@if [ ! -f README.md ]; then \
		echo "$(COLOR_RED)✗ README.md 不存在$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)✓ README.md 存在$(COLOR_RESET)"

# ==============================================================================
# CI/CD
# ==============================================================================

.PHONY: ci
ci: lint test-all build package ## CI 流程：检查 + 测试 + 构建 + 打包

.PHONY: release
release: clean ci ## 发布新版本
	@echo "$(COLOR_BLUE)[Release] 准备发布版本...$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ 版本 $(VERSION) 准备就绪$(COLOR_RESET)"
	@ls -lh $(DIST_DIR)/

# ==============================================================================
# 性能分析
# ==============================================================================

.PHONY: bench
bench: ## 运行性能基准测试
	@echo "$(COLOR_BLUE)[Benchmark] 运行性能测试...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem -run=^None ./...
	@echo "$(COLOR_GREEN)✓ 性能测试完成$(COLOR_RESET)"

.PHONY: bench-short
bench-short: ## 运行简短性能测试
	@echo "$(COLOR_BLUE)[Benchmark] 运行简短性能测试...$(COLOR_RESET)"
	@$(GO) test -bench=. -run=^None ./...
	@echo "$(COLOR_GREEN)✓ 简短性能测试完成$(COLOR_RESET)"

# ==============================================================================
# 快速命令
# ==============================================================================

.PHONY: init
init: deps-python ## 初始化项目
	@echo "$(COLOR_BLUE)[Init] 初始化项目...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ 项目初始化完成$(COLOR_RESET)"

.PHONY: setup
setup: deps-all build ## 完整设置（依赖 + 构建）
	@echo "$(COLOR_GREEN)✓ 设置完成！可以开始使用了$(COLOR_RESET)"

.PHONY: check
check: ## 检查环境
	@echo "$(COLOR_BLUE)[Check] 检查环境...$(COLOR_RESET)"
	@echo "Go 版本："
	@$(GO) version
	@echo "Python 版本："
	@$(PYTHON) --version
	@echo "项目目录：$(PWD)"
	@echo "$(COLOR_GREEN)✓ 环境检查完成$(COLOR_RESET)"

# ==============================================================================
# 对比工具
# ==============================================================================

.PHONY: compare
compare: ## 对比 Go 和 Ansible 的生成物
	@echo "$(COLOR_BLUE)[Compare] 运行对比检测...$(COLOR_RESET)"
	@./scripts/compare_outputs.sh
	@echo ""
	@echo "$(COLOR_GREEN)✓ 对比完成，查看报告：COMPARISON_LATEST.md$(COLOR_RESET)"

.PHONY: compare-quick
compare-quick: ## 快速对比（仅显示摘要）
	@echo "$(COLOR_BLUE)[Compare] 快速对比...$(COLOR_RESET)"
	@./scripts/compare_outputs.sh 2>&1 | grep -A 50 "对比完成总结"

.PHONY: compare-clean
compare-clean: ## 清理对比文件
	@echo "$(COLOR_BLUE)[Clean] 清理对比文件...$(COLOR_RESET)"
	@rm -rf comparison/
	@rm -f COMPARISON_REPORT_*.md
	@rm -f COMPARISON_LATEST.md
	@echo "$(COLOR_GREEN)✓ 对比文件已清理$(COLOR_RESET)"

.PHONY: verify
verify: run-generate compare ## 完整验证流程：生成 + 对比
	@echo "$(COLOR_GREEN)✓ 完整验证流程完成$(COLOR_RESET)"
