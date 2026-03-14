#!/usr/bin/env bash
# CMDB SQL 生成器对比测试脚本
# 用于验证 Go 版本生成的 SQL 配置文件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 配置目录（使用 GitLab Cfg 项目的配置目录，CMDB 复用这些配置）
BASE_DIR="/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg"
GO_OUTPUT_DIR="./output/cmdb-go"
ANSIBLE_BASE_DIR="${BASE_DIR}/../cmdb"
ANSIBLE_OUTPUT_DIR="${ANSIBLE_BASE_DIR}/out"

echo "================================================"
echo "CMDB SQL 生成器对比测试"
echo "================================================"
echo ""
echo "工作目录：$SCRIPT_DIR"
echo "基础目录：$BASE_DIR"
echo ""

# 清理旧输出
echo "[1/5] 清理输出目录..."
rm -rf "${GO_OUTPUT_DIR}"
if [ -d "${ANSIBLE_OUTPUT_DIR}" ]; then
    # 如果 Ansible 输出已存在，不删除它
    echo "  ✓ Ansible 输出目录已存在，保留"
else
    mkdir -p "${ANSIBLE_OUTPUT_DIR}"
fi
mkdir -p "${GO_OUTPUT_DIR}"

# 运行 Go 版本
echo "[2/5] 运行 Go 版本生成器..."
cd "$SCRIPT_DIR/.."
go run cmd/main.go cmdb \
    --base-dir "$BASE_DIR" \
    --vars vars-test.yaml \
    --resources resources.yaml \
    -o "$GO_OUTPUT_DIR"

if [ $? -ne 0 ]; then
    echo "✗ Go 版本生成失败"
    exit 1
fi
echo "✓ Go 版本生成成功"

# 复制 Ansible 版本的输出（如果存在）
echo "[3/5] 准备 Ansible 版本输出..."
if [ -d "${ANSIBLE_OUTPUT_DIR}" ]; then
    echo "✓ Ansible 版本输出已存在"
else
    echo "⚠️  Ansible 版本输出不存在，尝试生成..."
    
    # 检查 playbook 是否存在
    if [ -f "${ANSIBLE_BASE_DIR}/p1.yml" ]; then
        echo "🚀 运行 Ansible 生成器..."
        cd "${ANSIBLE_BASE_DIR}"
        
        # 执行 ansible-playbook
        if ansible-playbook p1.yml; then
            echo "✓ Ansible 版本生成成功"
        else
            echo "⚠️  Ansible 版本生成失败，跳过对比"
        fi
        
        cd "$SCRIPT_DIR/.."
    else
        echo "⚠️  Playbook 文件不存在 (${ANSIBLE_BASE_DIR}/p1.yml)，跳过对比"
    fi
fi

# 检查 Go 版本是否生成了文件
echo "[4/5] 检查输出文件..."
GO_FILE_COUNT=$(find "$GO_OUTPUT_DIR" -name "*.sql" | wc -l | tr -d ' ')

# 检查 Ansible 输出目录中的 SQL 文件（如果目录存在）
if [ -d "${ANSIBLE_OUTPUT_DIR}" ]; then
    ANSIBLE_FILE_COUNT=$(find "${ANSIBLE_OUTPUT_DIR}" -name "*.sql" 2>/dev/null | wc -l | tr -d ' ')
else
    ANSIBLE_FILE_COUNT=0
fi

echo "  Go 版本生成 $GO_FILE_COUNT 个文件"
echo "  Ansible 版本生成 $ANSIBLE_FILE_COUNT 个文件"

if [ "$GO_FILE_COUNT" -eq 0 ]; then
    echo "✗ Go 版本未生成任何 SQL 文件"
    exit 1
fi

echo "✓ Go 版本输出检查通过"

# 对比输出（如果有 Ansible 版本）
echo "[5/5] 对比输出结果..."
echo ""

if [ -d "$ANSIBLE_OUTPUT_DIR" ] && [ "$ANSIBLE_FILE_COUNT" -gt 0 ]; then
    # 使用 diff 比较两个目录
    DIFF_RESULT=0
    diff -rq "$GO_OUTPUT_DIR" "$ANSIBLE_OUTPUT_DIR" > /dev/null 2>&1 || DIFF_RESULT=$?
    
    if [ $DIFF_RESULT -eq 0 ]; then
        echo "✅ 恭喜！Go 和 Ansible 生成的输出完全一致！"
    else
        echo "⚠️  发现差异，详细比较如下："
        echo ""
        
        # 列出不同的文件
        echo "📁 不同的文件:"
        echo "================================================"
        diff -rq "$GO_OUTPUT_DIR" "$ANSIBLE_OUTPUT_DIR" || true
        echo ""
        
        # 只显示 SQL 文件的差异
        echo "📝 详细差异内容:"
        echo "================================================"
        
        find "$GO_OUTPUT_DIR" -name "*.sql" -type f | while read go_file; do
            relative_path="${go_file#$GO_OUTPUT_DIR/}"
            ansible_file="$ANSIBLE_OUTPUT_DIR/$relative_path"
            
            if [ -f "$ansible_file" ]; then
                # 比较文件
                if ! diff -q "$go_file" "$ansible_file" > /dev/null 2>&1; then
                    echo ""
                    echo "--- $relative_path ---"
                    echo ""
                    # 显示详细差异，限制输出长度
                    diff -u "$ansible_file" "$go_file" | head -100 || true
                    echo ""
                fi
            else
                echo ""
                echo "⚠️  仅在 Go 版本中存在：$relative_path"
            fi
        done
        
        # 查找仅在 Ansible 版本中存在的文件
        find "$ANSIBLE_OUTPUT_DIR" -name "*.sql" -type f | while read ansible_file; do
            relative_path="${ansible_file#$ANSIBLE_OUTPUT_DIR/}"
            go_file="$GO_OUTPUT_DIR/$relative_path"
            
            if [ ! -f "$go_file" ]; then
                echo ""
                echo "⚠️  仅在 Ansible 版本中存在：$relative_path"
            fi
        done
        
        echo ""
        echo "================================================"
        echo "提示：可以使用以下命令查看具体文件差异"
        echo "  diff -u $ANSIBLE_OUTPUT_DIR/<file> $GO_OUTPUT_DIR/<file>"
        echo "================================================"
        
        # 不退出，让用户看到完整的差异
    fi
else
    echo "⚠️  没有 Ansible 版本输出进行对比"
fi

echo ""
echo "================================================"
echo "✓ Go 版本 CMDB SQL 生成器功能验证完成！"
echo "================================================"
echo ""
