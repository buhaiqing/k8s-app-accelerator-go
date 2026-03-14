#!/usr/bin/env bash

# Jenkins Jobs 配置生成对比脚本
# 用于比较 Go 实现和 Ansible 实现的输出是否一致

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=================================================="
echo "Jenkins Jobs 配置生成对比脚本"
echo "=================================================="
echo ""

# 定义目录（Go 项目的输出目录）
GO_OUTPUT_DIR="${SCRIPT_DIR}/../output/go-jenkins"
ANSIBLE_OUTPUT_DIR="${SCRIPT_DIR}/../output/ansible-jenkins"

# 清理旧输出
echo "🧹 清理旧的输出目录..."
rm -rf "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}"
mkdir -p "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}"

# 1. 运行 Go 实现（复用 Ansible 的 vars.yaml）
echo ""
echo "🚀 运行 Go 实现..."
cd "${SCRIPT_DIR}/.."
go run cmd/main.go jenkins generate \
    --base-dir . \
    --config "/Users/bohaiqing/work/git/k8s_app_acelerator/jenkins_jobs/vars.yaml" \
    --output "${GO_OUTPUT_DIR}"

echo "✅ Go 实现完成"

# 2. 运行 Ansible 实现
echo ""
echo "🚀 运行 Ansible 实现..."
cd /Users/bohaiqing/work/git/k8s_app_acelerator/jenkins_jobs
ansible-playbook playbook.yaml \
    --extra-vars "rootdir=${ANSIBLE_OUTPUT_DIR}"

echo "✅ Ansible 实现完成"

# 3. 比较输出
echo ""
echo "=================================================="
echo "📊 比较输出结果"
echo "=================================================="
echo ""

# 使用 diff 比较
if diff -rq "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}" > /dev/null 2>&1; then
    echo "✅ 恭喜！Go 和 Ansible 生成的输出完全一致！"
    exit 0
else
    echo "⚠️  发现差异，详细比较如下："
    echo ""
    
    # 列出不同的文件
    echo "📁 不同的文件:"
    diff -rq "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}" || true
    echo ""
    
    # 显示具体差异
    echo "📝 详细差异:"
    diff -r "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}" || true
fi

echo ""
echo "=================================================="
echo "提示：可以使用以下命令查看具体文件差异"
echo "  diff ${GO_OUTPUT_DIR}/<file> ${ANSIBLE_OUTPUT_DIR}/<file>"
echo "=================================================="
