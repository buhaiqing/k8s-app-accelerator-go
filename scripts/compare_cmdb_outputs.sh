#!/usr/bin/env bash
# CMDB SQL 生成器对比测试脚本
# 用于验证 Go 版本生成的 SQL 配置文件

################################################################################
# 核心配置变量（用户只需修改这里）
################################################################################

# 配置目录（复用 GitLab Cfg 项目的配置文件，与 Ansible 保持一致）
BASE_DIR="${BASE_DIR:-/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg}"
CMDB_DIR="${CMDB_DIR:-/Users/bohaiqing/work/git/k8s_app_acelerator/cmdb}"
VARS_FILE="${VARS_FILE:-vars-test.yaml}"
RESOURCES_FILE="${RESOURCES_FILE:-resources.yaml}"

# 输出目录
GO_OUTPUT_DIR="${GO_OUTPUT_DIR:-./output/cmdb}"
ANSIBLE_OUTPUT_DIR="${ANSIBLE_OUTPUT_DIR:-${CMDB_DIR}/out}"

# Ansible Playbook 文件
ANSIBLE_PLAYBOOK="${CMDB_DIR}/p1.yml"

# 对比报告目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/comparison"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

################################################################################
# 版本和颜色定义
################################################################################
VERSION="1.0.0"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 统计变量
TOTAL_FILES_GO=0
TOTAL_FILES_ANSIBLE=0
IDENTICAL_FILES=0
DIFFERENT_FILES=0
TOLERATED_FILES=0
MISSING_IN_GO=0
MISSING_IN_ANSIBLE=0

################################################################################
# 智能文件对比函数
# 支持忽略：
# 1. 末尾空行差异
# 2. 行尾空格差异
# 注意：此函数不处理随机值差异，随机值差异由 is_tolerated_diff 处理
# 返回：0 表示一致，1 表示有差异
################################################################################
smart_compare_files() {
    local file1="$1"
    local file2="$2"

    if [ ! -f "$file1" ]; then
        return 1
    fi
    if [ ! -f "$file2" ]; then
        return 1
    fi

    local tmp1=$(mktemp)
    local tmp2=$(mktemp)

    # 预处理文件1：去除末尾空行和行尾空格
    sed 's/[[:space:]]*$//' "$file1" | sed '/^[[:space:]]*$/d' > "$tmp1" 2>/dev/null || cat "$file1" > "$tmp1"

    # 预处理文件2：去除末尾空行和行尾空格
    sed 's/[[:space:]]*$//' "$file2" | sed '/^[[:space:]]*$/d' > "$tmp2" 2>/dev/null || cat "$file2" > "$tmp2"

    # 直接对比预处理后的文件（不标准化随机值）
    local result=0
    cmp -s "$tmp1" "$tmp2" || result=$?

    # 清理临时文件
    rm -f "$tmp1" "$tmp2"

    return ${result}
}

# 检查是否是容许差异（仅随机数不同）
# 返回：0 表示仅随机数不同（容许），1 表示有其他差异
is_tolerated_diff() {
    local file1="$1"
    local file2="$2"

    if [ ! -f "$file1" ] || [ ! -f "$file2" ]; then
        return 1
    fi

    local tmp1=$(mktemp)
    local tmp2=$(mktemp)

    # 预处理：去除末尾空行和行尾空格
    sed 's/[[:space:]]*$//' "$file1" | sed '/^[[:space:]]*$/d' > "$tmp1" 2>/dev/null || cat "$file1" > "$tmp1"
    sed 's/[[:space:]]*$//' "$file2" | sed '/^[[:space:]]*$/d' > "$tmp2" 2>/dev/null || cat "$file2" > "$tmp2"

    # 标准化随机值
    sed -E 's/(production|int|test)-rds-[0-9]+/\1-rds-XXXX/g' "$tmp1" | \
    sed -E 's/(production|int|test)-pg-[0-9]+/\1-pg-XXXX/g' | \
    sed -E 's/(production|int|test)-mongo-[0-9]+/\1-mongo-XXXX/g' | \
    sed -E 's/(rdsdb|pddb|dds)[0-9]+/\1XXXX/g' > "$tmp1.tmp"
    mv "$tmp1.tmp" "$tmp1"

    sed -E 's/(production|int|test)-rds-[0-9]+/\1-rds-XXXX/g' "$tmp2" | \
    sed -E 's/(production|int|test)-pg-[0-9]+/\1-pg-XXXX/g' | \
    sed -E 's/(production|int|test)-mongo-[0-9]+/\1-mongo-XXXX/g' | \
    sed -E 's/(rdsdb|pddb|dds)[0-9]+/\1XXXX/g' > "$tmp2.tmp"
    mv "$tmp2.tmp" "$tmp2"

    local result=0
    cmp -s "$tmp1" "$tmp2" || result=$?

    rm -f "$tmp1" "$tmp2"
    return ${result}
}

# 获取容许差异的详情
get_tolerated_details() {
    local file1="$1"
    local file2="$2"
    local details=""

    # 提取随机值差异（带引号）
    local randoms1
    local randoms2
    local db_ids1
    local db_ids2

    randoms1=$(grep -oE "'(production|int|test)-(rds|pg|mongo)-[0-9]+'" "$file1" 2>/dev/null | tr -d "'" | sort -u)
    randoms2=$(grep -oE "'(production|int|test)-(rds|pg|mongo)-[0-9]+'" "$file2" 2>/dev/null | tr -d "'" | sort -u)

    db_ids1=$(grep -oE "'(rdsdb|pddb|dds)[0-9]+'" "$file1" 2>/dev/null | tr -d "'" | sort -u)
    db_ids2=$(grep -oE "'(rdsdb|pddb|dds)[0-9]+'" "$file2" 2>/dev/null | tr -d "'" | sort -u)

    if [ -n "$randoms1" ] || [ -n "$db_ids1" ]; then
        details="随机ID差异: "
        # 添加file1的随机值（去除换行，最多显示3个）
        if [ -n "$randoms1" ]; then
            local randoms1_short=$(echo "$randoms1" | head -n 3 | tr '\n' ' ' | sed 's/ $//')
            details+="[$randoms1_short"
            local randoms1_count=$(echo "$randoms1" | wc -l)
            if [ "$randoms1_count" -gt 3 ]; then
                details+="... 共${randoms1_count}个"
            fi
            details+="]"
        fi
        if [ -n "$db_ids1" ]; then
            local db_ids1_short=$(echo "$db_ids1" | head -n 3 | tr '\n' ' ' | sed 's/ $//')
            details+="[$db_ids1_short"
            local db_ids1_count=$(echo "$db_ids1" | wc -l)
            if [ "$db_ids1_count" -gt 3 ]; then
                details+="... 共${db_ids1_count}个"
            fi
            details+="]"
        fi
        details+=" vs "
        # 添加file2的随机值（去除换行，最多显示3个）
        if [ -n "$randoms2" ]; then
            local randoms2_short=$(echo "$randoms2" | head -n 3 | tr '\n' ' ' | sed 's/ $//')
            details+="[$randoms2_short"
            local randoms2_count=$(echo "$randoms2" | wc -l)
            if [ "$randoms2_count" -gt 3 ]; then
                details+="... 共${randoms2_count}个"
            fi
            details+="]"
        fi
        if [ -n "$db_ids2" ]; then
            local db_ids2_short=$(echo "$db_ids2" | head -n 3 | tr '\n' ' ' | sed 's/ $//')
            details+="[$db_ids2_short"
            local db_ids2_count=$(echo "$db_ids2" | wc -l)
            if [ "$db_ids2_count" -gt 3 ]; then
                details+="... 共${db_ids2_count}个"
            fi
            details+="]"
        fi
    fi

    echo "$details"
}

################################################################################
# 显示带颜色的消息
################################################################################
print_msg() {
    local color=$1
    local msg=$2
    echo -e "${color}${msg}${NC}"
}

print_success() {
    print_msg "$GREEN" "✓ $1"
}

print_error() {
    print_msg "$RED" "✗ $1"
}

print_warning() {
    print_msg "$YELLOW" "⚠️  $1"
}

print_info() {
    print_msg "$CYAN" "ℹ $1"
}

################################################################################
# 显示帮助信息
################################################################################
show_help() {
    cat << EOF
CMDB SQL 生成器对比测试脚本 v${VERSION}

用法:
  $0 [OPTIONS]

选项:
  -h, --help              显示帮助信息
  -v, --verbose          显示详细输出
  -n, --no-generate      跳过自动生成，仅进行对比
  -d, --diff-only        仅显示差异，不自动生成
  --base-dir DIR         指定基础配置目录
  --cmdb-dir DIR         指定 CMDB 目录
  --go-output DIR        指定 Go 输出目录
  --ansible-output DIR   指定 Ansible 输出目录

示例:
  $0                                  # 默认运行
  $0 --no-generate                    # 跳过生成，仅对比
  $0 --base-dir /path/to/gitlab_cfg   # 指定配置目录
  $0 --verbose                         # 显示详细输出

配置文件:
  使用 Ansible 相同的配置文件（vars-test.yaml, resources.yaml）
  确保与 Ansible Playbook (p1.yml) 保持一致
EOF
}

################################################################################
# 主函数
################################################################################
main() {
    local verbose=false
    local no_generate=false
    local diff_only=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            -n|--no-generate)
                no_generate=true
                shift
                ;;
            -d|--diff-only)
                diff_only=true
                shift
                ;;
            --base-dir)
                BASE_DIR="$2"
                shift 2
                ;;
            --cmdb-dir)
                CMDB_DIR="$2"
                shift 2
                ;;
            --go-output)
                GO_OUTPUT_DIR="$2"
                shift 2
                ;;
            --ansible-output)
                ANSIBLE_OUTPUT_DIR="$2"
                shift 2
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    echo "================================================"
    echo "CMDB SQL 生成器对比测试 v${VERSION}"
    echo "================================================"
    echo ""
    echo "📁 配置信息："
    echo "  - 基础配置目录: $BASE_DIR"
    echo "  - CMDB 目录: $CMDB_DIR"
    echo "  - 配置文件: $VARS_FILE, $RESOURCES_FILE"
    echo "  - Go 输出: $GO_OUTPUT_DIR"
    echo "  - Ansible 输出: $ANSIBLE_OUTPUT_DIR"
    echo ""

    # 创建输出目录
    mkdir -p "$GO_OUTPUT_DIR"
    mkdir -p "$(dirname "$ANSIBLE_OUTPUT_DIR")"

    # Step 1: 生成或检查输出
    echo "================================================"
    echo "📦 生成输出文件"
    echo "================================================"
    echo ""

    # 运行 Go 版本
    if [ "$no_generate" = false ] && [ "$diff_only" = false ]; then
        echo "[1/3] 运行 Go 版本生成器..."
        cd "$PROJECT_ROOT"

        if go run cmd/main.go cmdb \
            --base-dir "$BASE_DIR" \
            --vars "$VARS_FILE" \
            --resources "$RESOURCES_FILE" \
            -o "$GO_OUTPUT_DIR" 2>&1; then
            print_success "Go 版本生成成功"
        else
            print_error "Go 版本生成失败"
            exit 1
        fi
        echo ""
    fi

    # 检查/生成 Ansible 版本
    echo "[2/3] 检查 Ansible 版本输出..."

    if [ ! -d "$ANSIBLE_OUTPUT_DIR" ] || [ -z "$(ls -A "$ANSIBLE_OUTPUT_DIR" 2>/dev/null)" ]; then
        print_warning "Ansible 版本输出不存在，尝试生成..."

        if [ -f "$ANSIBLE_PLAYBOOK" ]; then
            cd "$CMDB_DIR"
            if ansible-playbook p1.yml -v 2>&1; then
                print_success "Ansible 版本生成成功"
            else
                print_error "Ansible 版本生成失败"
                exit 1
            fi
            cd "$PROJECT_ROOT"
        else
            print_error "Playbook 文件不存在: $ANSIBLE_PLAYBOOK"
            exit 1
        fi
    else
        print_success "Ansible 版本输出已存在"
    fi
    echo ""

    # Step 2: 对比输出
    echo "================================================"
    echo "🔍 对比输出结果"
    echo "================================================"
    echo ""

    # 获取文件列表
    local go_files=($(find "$GO_OUTPUT_DIR" -name "*.sql" -type f 2>/dev/null | sort))
    local ansible_files=($(find "$ANSIBLE_OUTPUT_DIR" -name "*.sql" -type f 2>/dev/null | sort))

    TOTAL_FILES_GO=${#go_files[@]}
    TOTAL_FILES_ANSIBLE=${#ansible_files[@]}

    echo "📊 文件统计："
    echo "  - Go 版本: $TOTAL_FILES_GO 个文件"
    echo "  - Ansible 版本: $TOTAL_FILES_ANSIBLE 个文件"
    echo ""

    if [ "$TOTAL_FILES_GO" -eq 0 ]; then
        print_error "Go 版本未生成任何 SQL 文件"
        exit 1
    fi

    # 对比文件
    echo "📝 逐文件对比："
    echo "================================================"

    local has_diff=false

    for go_file in "${go_files[@]}"; do
        local filename=$(basename "$go_file")
        local ansible_file="$ANSIBLE_OUTPUT_DIR/$filename"

        if [ -f "$ansible_file" ]; then
            if smart_compare_files "$go_file" "$ansible_file"; then
                echo "  ✓ $filename (一致)"
                ((IDENTICAL_FILES++))
            elif is_tolerated_diff "$go_file" "$ansible_file"; then
                local diff_details=$(get_tolerated_details "$go_file" "$ansible_file")
                if [ -n "$diff_details" ]; then
                    echo "  ✓ $filename (一致，仅 $diff_details)"
                else
                    echo "  ✓ $filename (一致，随机值差异已容许)"
                fi
                ((TOLERATED_FILES++))
            else
                echo "  ✗ $filename (有差异)"
                ((DIFFERENT_FILES++))
                has_diff=true
            fi
        else
            echo "  ⚠️  $filename (仅在 Go 版本中存在)"
            ((MISSING_IN_ANSIBLE++))
            has_diff=true
        fi
    done

    # 查找仅在 Ansible 版本中存在的文件
    for ansible_file in "${ansible_files[@]}"; do
        local filename=$(basename "$ansible_file")
        local go_file="$GO_OUTPUT_DIR/$filename"

        if [ ! -f "$go_file" ]; then
            echo "  ⚠️  $filename (仅在 Ansible 版本中存在)"
            ((MISSING_IN_GO++))
            has_diff=true
        fi
    done

    echo ""

    # Step 3: 显示详细差异
    if [ "$has_diff" = true ]; then
        echo "================================================"
        echo "📄 详细差异内容"
        echo "================================================"
        echo ""

        for go_file in "${go_files[@]}"; do
            local filename=$(basename "$go_file")
            local ansible_file="$ANSIBLE_OUTPUT_DIR/$filename"

            if [ -f "$ansible_file" ]; then
                if ! smart_compare_files "$go_file" "$ansible_file"; then
                    echo "📄 $filename:"
                    echo "----------------------------------------"
                    diff -u "$ansible_file" "$go_file" 2>/dev/null | head -50 || true
                    echo ""
                fi
            else
                echo "📄 $filename (仅在 Go 版本中存在):"
                echo "----------------------------------------"
                head -20 "$go_file" || true
                echo ""
            fi
        done

        # 显示仅在 Ansible 版本中存在的文件
        for ansible_file in "${ansible_files[@]}"; do
            local filename=$(basename "$ansible_file")
            local go_file="$GO_OUTPUT_DIR/$filename"

            if [ ! -f "$go_file" ]; then
                echo "📄 $filename (仅在 Ansible 版本中存在):"
                echo "----------------------------------------"
                head -20 "$ansible_file" || true
                echo ""
            fi
        done
    fi

    # 总结
    echo "================================================"
    echo "📈 对比总结"
    echo "================================================"
    echo ""
    echo "  总计:"
    echo "    - Go 版本文件数: $TOTAL_FILES_GO"
    echo "    - Ansible 版本文件数: $TOTAL_FILES_ANSIBLE"
    echo ""
    echo "  对比结果:"
    echo "    - 一致: $IDENTICAL_FILES"
    echo "    - 容许差异: $TOLERATED_FILES"
    echo "    - 有差异: $DIFFERENT_FILES"
    echo "    - 仅在 Go: $MISSING_IN_ANSIBLE"
    echo "    - 仅在 Ansible: $MISSING_IN_GO"
    echo ""

    if [ "$DIFFERENT_FILES" -eq 0 ] && [ "$MISSING_IN_GO" -eq 0 ] && [ "$MISSING_IN_ANSIBLE" -eq 0 ]; then
        print_success "🎉 Go 和 Ansible 生成的输出完全一致！"
        echo ""
        exit 0
    else
        print_warning "⚠️  发现差异，请检查上述详细输出"
        echo ""
        echo "提示：可以使用以下命令查看具体文件差异"
        echo "  diff -u $ANSIBLE_OUTPUT_DIR/<file> $GO_OUTPUT_DIR/<file>"
        echo ""
        exit 1
    fi
}

# 运行主函数
main "$@"
