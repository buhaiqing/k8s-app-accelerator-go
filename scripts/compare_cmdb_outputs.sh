#!/usr/bin/env bash
# CMDB SQL 生成器对比测试脚本
# 用于验证 Go 版本生成的 SQL 配置文件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 配置目录（使用 GitLab Cfg 项目的配置目录，CMDB 复用这些配置）
BASE_DIR="/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg"
GO_OUTPUT_DIR="./output/cmdb-go"

echo "================================================"
echo "CMDB SQL 生成器验证测试"
echo "================================================"
echo ""
echo "工作目录：$SCRIPT_DIR"
echo "基础目录：$BASE_DIR"
echo ""

# 清理旧输出
echo "[1/4] 清理输出目录..."
rm -rf "${GO_OUTPUT_DIR}"
mkdir -p "${GO_OUTPUT_DIR}"

# 运行 Go 版本
echo "[2/4] 运行 Go 版本生成器..."
cd "$SCRIPT_DIR/.."
go run cmd/main.go cmdb \
    --base-dir "$BASE_DIR" \
    --vars vars.yaml \
    --resources resources.yaml \
    -o "$GO_OUTPUT_DIR"

if [ $? -ne 0 ]; then
    echo "✗ Go 版本生成失败"
    exit 1
fi
echo "✓ Go 版本生成成功"

# 检查 Go 版本是否生成了文件
echo "[3/4] 检查 Go 版本输出..."
GO_FILE_COUNT=$(find "$GO_OUTPUT_DIR" -name "*.sql" | wc -l | tr -d ' ')
echo "  Go 版本生成 $GO_FILE_COUNT 个文件"

if [ "$GO_FILE_COUNT" -eq 0 ]; then
    echo "✗ Go 版本未生成任何 SQL 文件"
    exit 1
fi

echo "✓ Go 版本输出检查通过"

# 显示生成的文件列表
echo ""
echo "Go 版本生成的文件:"
find "$GO_OUTPUT_DIR" -name "*.sql" | sort | while read file; do
    echo "  - $(basename "$file")"
done

# 显示文件内容摘要
echo ""
echo "文件内容摘要:"
find "$GO_OUTPUT_DIR" -name "*.sql" | sort | while read file; do
    echo ""
    echo "--- $(basename "$file") ---"
    head -20 "$file"
done

echo ""
echo "================================================"
echo "✓ Go 版本 CMDB SQL 生成器功能验证通过！"
echo "================================================"
echo ""
