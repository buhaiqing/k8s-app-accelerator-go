#!/bin/bash
# ==============================================================================
# 文件名称：compare_argocd_outputs.sh
# 功能描述：比对 ArgoCD Golang实现与 Ansible 实现的生成物一致性
# 使用示例：bash scripts/compare_argocd_outputs.sh
# ==============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 打印分隔线
print_separator() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# ==============================================================================
# 配置参数
# ==============================================================================

# Ansible 输出目录
ANSIBLE_OUTPUT_DIR="${ANSIBLE_OUTPUT_DIR:-/Users/bohaiqing/work/git/k8s_app_acelerator/argocd/out}"

# Go 输出目录（需要先生成）
GO_OUTPUT_DIR="${GO_OUTPUT_DIR:-/Users/bohaiqing/opensource/git/k8s-app-accelerator-go/output/argo-app}"

# 对比报告目录
COMPARISON_DIR="${COMPARISON_DIR:-/Users/bohaiqing/opensource/git/k8s-app-accelerator-go/comparison/argocd}"

# 时间戳
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# ==============================================================================
# 前置检查
# ==============================================================================

print_separator
echo "🔍 ArgoCD 配置一致性对比工具 v1.0.0"
print_separator

print_info "检测时间: $(date '+%Y-%m-%d %H:%M:%S')"
print_info "Ansible 输出目录：$ANSIBLE_OUTPUT_DIR"
print_info "Go 输出目录：$GO_OUTPUT_DIR"
print_info "对比报告目录：$COMPARISON_DIR"

# 检查目录是否存在
check_directory() {
    local dir=$1
    local name=$2
    
    if [ ! -d "$dir" ]; then
        print_error "$name 目录不存在：$dir"
        return 1
    fi
    print_success "$name 目录已就绪：$dir"
    return 0
}

print_separator
echo "检查前置条件"
print_separator

check_directory "$ANSIBLE_OUTPUT_DIR" "Ansible 输出" || exit 1
check_directory "$GO_OUTPUT_DIR" "Go 输出" || exit 1

# 创建对比报告目录
mkdir -p "$COMPARISON_DIR"

# ==============================================================================
# 统计文件数量
# ==============================================================================

print_separator
echo "统计文件清单"
print_separator

# 获取 Ansible 生成的文件列表
ansible_files=$(find "$ANSIBLE_OUTPUT_DIR" -type f -name "*.yaml" | sort)
ansible_count=$(echo "$ansible_files" | wc -l | tr -d ' ')

# 获取 Go 生成的文件列表
go_files=$(find "$GO_OUTPUT_DIR" -type f -name "*.yaml" | sort)
go_count=$(echo "$go_files" | wc -l | tr -d ' ')

print_info "Ansible 生成文件总数：$ansible_count"
print_info "Go 生成文件总数：$go_count"

# 保存文件列表
echo "$ansible_files" > "$COMPARISON_DIR/ansible_files_${TIMESTAMP}.txt"
echo "$go_files" > "$COMPARISON_DIR/go_files_${TIMESTAMP}.txt"

print_info "文件列表已保存:"
print_info "  - Ansible: $COMPARISON_DIR/ansible_files_${TIMESTAMP}.txt"
print_info "  - Go: $COMPARISON_DIR/go_files_${TIMESTAMP}.txt"

# ==============================================================================
# 提取相对路径并对比
# ==============================================================================

print_separator
echo "提取相对路径"
print_separator

# 提取 Ansible 文件的相对路径（相对于输出目录）
cd "$ANSIBLE_OUTPUT_DIR"
ansible_relative=$(find . -type f -name "*.yaml" | sed 's|^\./||' | sort)
cd - > /dev/null

# 提取 Go 文件的相对路径（相对于输出目录）
cd "$GO_OUTPUT_DIR"
go_relative=$(find . -type f -name "*.yaml" | sed 's|^\./||' | sort)
cd - > /dev/null

# 保存相对路径列表
echo "$ansible_relative" > "$COMPARISON_DIR/ansible_relative_${TIMESTAMP}.txt"
echo "$go_relative" > "$COMPARISON_DIR/go_relative_${TIMESTAMP}.txt"

# ==============================================================================
# 文件清单对比
# ==============================================================================

print_separator
echo "对比文件清单"
print_separator

# 找出 Ansible 有但 Go 没有的文件
ansible_only=$(echo "$ansible_relative" | sort > /tmp/ansible_sorted.txt && echo "$go_relative" | sort > /tmp/go_sorted.txt && comm -23 /tmp/ansible_sorted.txt /tmp/go_sorted.txt || echo "")
ansible_only_count=$(echo "$ansible_only" | grep -c '.' 2>/dev/null || echo 0)

# 找出 Go 有但 Ansible 没有的文件
go_only=$(echo "$ansible_relative" | sort > /tmp/ansible_sorted.txt && echo "$go_relative" | sort > /tmp/go_sorted.txt && comm -13 /tmp/ansible_sorted.txt /tmp/go_sorted.txt || echo "")
go_only_count=$(echo "$go_only" | grep -c '.' 2>/dev/null || echo 0)

# 找出共同的文件
common_files=$(echo "$ansible_relative" | sort > /tmp/ansible_sorted.txt && echo "$go_relative" | sort > /tmp/go_sorted.txt && comm -12 /tmp/ansible_sorted.txt /tmp/go_sorted.txt || echo "")
common_count=$(echo "$common_files" | grep -c '.' 2>/dev/null || echo 0)

if [ "$ansible_only_count" -gt 0 ]; then
    print_warning "Ansible 独有文件 ($ansible_only_count 个):"
    echo "$ansible_only" | while read -r file; do
        [ -n "$file" ] && echo "  - $file"
    done
fi

if [ "$go_only_count" -gt 0 ]; then
    print_warning "Go 独有文件 ($go_only_count 个):"
    echo "$go_only" | while read -r file; do
        [ -n "$file" ] && echo "  - $file"
    done
fi

print_info "共同文件数量：$common_count"

# 保存差异信息
{
    echo "Ansible Only Files:"
    echo "$ansible_only"
    echo ""
    echo "Go Only Files:"
    echo "$go_only"
} > "$COMPARISON_DIR/file_list_diff_${TIMESTAMP}.txt"

# ==============================================================================
# 文件内容对比
# ==============================================================================

print_separator
echo "对比文件内容"
print_separator

identical_files=""
different_files=""
missing_files=""

identical_count=0
different_count=0
missing_count=0

# 逐个对比共同文件
echo "$common_files" | while read -r file; do
    [ -z "$file" ] && continue
    
    ansible_file="$ANSIBLE_OUTPUT_DIR/$file"
    go_file="$GO_OUTPUT_DIR/$file"
    
    if [ ! -f "$go_file" ]; then
        print_error "缺失：$file"
        echo "$file" >> "$COMPARISON_DIR/missing_files_${TIMESTAMP}.txt"
        continue
    fi
    
    # 对比文件内容（忽略空白字符和注释差异）
    if diff -q "$ansible_file" "$go_file" > /dev/null 2>&1; then
        print_success "相同：$file"
        echo "$file" >> "$COMPARISON_DIR/identical_files_${TIMESTAMP}.txt"
    else
        print_warning "不同：$file"
        echo "$file" >> "$COMPARISON_DIR/different_files_${TIMESTAMP}.txt"
        
        # 保存详细差异
        echo "===== 文件：$file =====" >> "$COMPARISON_DIR/content_diff_${TIMESTAMP}.txt"
        diff -u "$ansible_file" "$go_file" >> "$COMPARISON_DIR/content_diff_${TIMESTAMP}.txt" 2>&1 || true
        echo "" >> "$COMPARISON_DIR/content_diff_${TIMESTAMP}.txt"
    fi
done

# 统计结果
identical_count=$(wc -l < "$COMPARISON_DIR/identical_files_${TIMESTAMP}.txt" 2>/dev/null | tr -d ' ' || echo 0)
different_count=$(wc -l < "$COMPARISON_DIR/different_files_${TIMESTAMP}.txt" 2>/dev/null | tr -d ' ' || echo 0)
missing_count=$(wc -l < "$COMPARISON_DIR/missing_files_${TIMESTAMP}.txt" 2>/dev/null | tr -d ' ' || echo 0)

print_separator
echo "内容对比统计"
print_separator

print_info "完全一致的文件：${identical_count:-0}"
[ "${different_count:-0}" -gt 0 ] 2>/dev/null && print_warning "内容有差异的文件：${different_count:-0}"
[ "${missing_count:-0}" -gt 0 ] 2>/dev/null && print_error "缺失的文件：${missing_count:-0}"

# ==============================================================================
# 生成汇总报告
# ==============================================================================

print_separator
echo "生成汇总报告"
print_separator

REPORT_FILE="$COMPARISON_DIR/comparison_report_${TIMESTAMP}.md"

cat > "$REPORT_FILE" << EOF
# ArgoCD 配置一致性对比报告

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')
**Ansible 目录**: $ANSIBLE_OUTPUT_DIR
**Go 目录**: $GO_OUTPUT_DIR

---

## 📊 对比结果汇总

### 文件统计

| 指标 | 数量 |
|------|------|
| Ansible 文件总数 | $ansible_count |
| Go 文件总数 | $go_count |
| 完全一致的文件 | $identical_count |
| 内容有差异的文件 | $different_count |
| 缺失的文件 | $missing_count |

---

## 📁 详细文件清单

### ✅ 完全一致的文件 ($identical_count 个)

EOF

if [ -f "$COMPARISON_DIR/identical_files_${TIMESTAMP}.txt" ]; then
    while IFS= read -r file; do
        [ -n "$file" ] && echo "- \`$file\`" >> "$REPORT_FILE"
    done < "$COMPARISON_DIR/identical_files_${TIMESTAMP}.txt"
fi

cat >> "$REPORT_FILE" << EOF

### ⚠️ 内容有差异的文件 ($different_count 个)

EOF

if [ -f "$COMPARISON_DIR/different_files_${TIMESTAMP}.txt" ]; then
    while IFS= read -r file; do
        [ -n "$file" ] && echo "- \`$file\`" >> "$REPORT_FILE"
    done < "$COMPARISON_DIR/different_files_${TIMESTAMP}.txt"
else
    echo "无" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << EOF

### ❌ 缺失的文件 ($missing_count 个)

EOF

if [ -f "$COMPARISON_DIR/missing_files_${TIMESTAMP}.txt" ]; then
    while IFS= read -r file; do
        [ -n "$file" ] && echo "- \`$file\`" >> "$REPORT_FILE"
    done < "$COMPARISON_DIR/missing_files_${TIMESTAMP}.txt"
else
    echo "无" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << EOF

---

## 🔗 相关文件

- 相同文件列表：identical_files_${TIMESTAMP}.txt
- 差异文件列表：different_files_${TIMESTAMP}.txt
- 缺失文件列表：missing_files_${TIMESTAMP}.txt
- 详细差异内容：content_diff_${TIMESTAMP}.txt
- 文件清单 (Ansible): ansible_files_${TIMESTAMP}.txt
- 文件清单 (Go): go_files_${TIMESTAMP}.txt

---

**生成工具**: ArgoCD 对比脚本 v1.0.0
EOF

print_success "汇总报告已生成：$REPORT_FILE"

# ==============================================================================
# 最终总结
# ==============================================================================

print_separator
echo "对比完成总结"
print_separator

# 计算一致性比例
if [ "$ansible_count" -gt 0 ] && [ "$identical_count" != "" ] && [ "$ansible_count" != "" ]; then
    consistency_rate=$(awk -v identical="$identical_count" -v total="$ansible_count" 'BEGIN {printf "%.1f", (identical / total) * 100}')
else
    consistency_rate="0.0"
fi

print_info "Ansible 文件总数：$ansible_count"
print_info "Go 文件总数：$go_count"
print_info "完全一致：${identical_count:-0} 个"
[ "${different_count:-0}" -gt 0 ] 2>/dev/null && print_warning "内容差异：${different_count:-0} 个"
[ "${missing_count:-0}" -gt 0 ] 2>/dev/null && print_error "缺失文件：${missing_count:-0} 个"
print_info "一致性比例：${consistency_rate}%"

print_separator

if [ "${different_count:-0}" -eq 0 ] 2>/dev/null && [ "${missing_count:-0}" -eq 0 ] 2>/dev/null; then
    print_success "🎉 完美！Golang 与 Ansible 生成的配置完全一致！"
else
    print_warning "💡 发现差异，请查看详细报告："
    print_warning "  $REPORT_FILE"
    
    if [ "${different_count:-0}" -gt 0 ] 2>/dev/null; then
        print_info "查看差异详情："
        print_info "  cat $COMPARISON_DIR/content_diff_${TIMESTAMP}.txt"
    fi
fi

print_separator

# 创建最新报告的软链接
latest_report="$COMPARISON_DIR/latest_comparison.md"
rm -f "$latest_report"
ln -s "$(basename "$REPORT_FILE")" "$latest_report" 2>/dev/null || cp "$REPORT_FILE" "$latest_report"

print_success "最新报告链接：$latest_report"
print_separator

# 如果有差异，退出时返回非零状态码
if [ "${different_count:-0}" -gt 0 ] 2>/dev/null || [ "${missing_count:-0}" -gt 0 ] 2>/dev/null; then
    exit 1
fi

exit 0
