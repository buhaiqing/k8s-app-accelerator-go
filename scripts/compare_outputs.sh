#!/bin/bash

################################################################################
# K8s App Accelerator - 统一对比工具
# 
# 功能：自动比对 GitLab Cfg 模块的 Go 实现和 Ansible 实现的生成物一致性
# 版本：v1.3.0 - 统一对比工具（整合 ArgoCD + Jenkins + CMS）
# 
# 用法:
#   ./scripts/compare_outputs.sh                                    # 使用默认路径
#   ./scripts/compare_outputs.sh --no-auto-generate-go             # 禁用自动生成
#   ./scripts/compare_outputs.sh /path/to/ansible/output           # 指定 Ansible 输出
#   ./scripts/compare_outputs.sh /path/to/ansible /path/to/go      # 指定两个输出
#   ./scripts/compare_outputs.sh --help                             # 显示帮助
#
# Ansible 生成命令:
#   cd /Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg
#   ansible-playbook bootstrap-test.yml -e '@vars-test.yaml'
#
# Go 生成命令:
#   cd /Users/bohaiqing/opensource/git/k8s-app-accelerator-go
#   go run main.go gitlab-cfg generate --base-dir "${ANSIBLE_ROOT}" --config vars-test.yaml --skip-precheck
#
# 对比范围:
#   - CMS K8s 配置 (deployment, service, config, hpa, job 等)
#   - ArgoCD Application 配置
#   - Jenkins Job 配置
#
# 作者：K8s App Accelerator Team
# 日期：2026-03-15
# 
# 变更历史:
#   v1.3.0 - 整合 ArgoCD 和 Jenkins 对比逻辑，统一为单一入口脚本
#   v1.2.0 - 添加环境变量支持和改进帮助系统
################################################################################

set -e  # 遇到错误立即退出

# ============================================
# 核心配置（用户只需修改这里）
# 支持通过环境变量覆盖默认值
# ============================================
ANSIBLE_ROOT="${ANSIBLE_ROOT:-/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg}"
VARS_FILE="${VARS_FILE:-${ANSIBLE_ROOT}/vars-test.yaml}"
MAPPING_FILE="${MAPPING_FILE:-${ANSIBLE_ROOT}/mapping.yaml}"
RESOURCES_FILE="${RESOURCES_FILE:-${ANSIBLE_ROOT}/resources.yaml}"
BOOTSTRAP_FILE="${BOOTSTRAP_FILE:-${ANSIBLE_ROOT}/bootstrap-test.yml}"

# ============================================
# 自动计算的变量（无需修改）
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Go 项目根目录
GO_ROOT="${PROJECT_ROOT}"

# 输出目录（默认值）
ANSIBLE_OUTPUT_DEFAULT="${ANSIBLE_ROOT}/output"
GO_OUTPUT_DEFAULT="${GO_ROOT}/output"

# 对比和报告目录
COMPARISON_DIR="${PROJECT_ROOT}/comparison"
REPORT_DIR="${PROJECT_ROOT}"

# ============================================
# 版本和颜色定义
# ============================================
VERSION="1.2.0"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 统计变量
TOTAL_FILES_ANSIBLE=0
TOTAL_FILES_GO=0
IDENTICAL_FILES=0
TOLERATED_FILES=0
DIFFERENT_FILES=0
MISSING_FILES=0

# 自动生成 Go 输出（默认开启）
AUTO_GENERATE_GO=true

# ============================================
# 帮助信息
# ============================================
show_help() {
    cat << EOF
K8s App Accelerator - GitLab Cfg 对比工具 v${VERSION}

用法:
  \$0 [OPTIONS] [ANSIBLE_OUTPUT_DIR] [GO_OUTPUT_DIR]

选项:
  -h, --help                 显示此帮助信息
  --no-auto-generate-go      Go 输出为空时不自动生成

环境变量（可通过环境变量覆盖默认配置）:
  ANSIBLE_ROOT               Ansible 项目根目录
                             默认：/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg
  VARS_FILE                  配置文件路径
                             默认：\${ANSIBLE_ROOT}/vars-test.yaml
  MAPPING_FILE               Mapping 文件路径
                             默认：\${ANSIBLE_ROOT}/mapping.yaml
  RESOURCES_FILE             Resources 文件路径
                             默认：\${ANSIBLE_ROOT}/resources.yaml
  BOOTSTRAP_FILE             Bootstrap 文件路径
                             默认：\${ANSIBLE_ROOT}/bootstrap-test.yml

参数:
  ANSIBLE_OUTPUT_DIR         Ansible 生成的输出目录路径
  GO_OUTPUT_DIR              Go 生成的输出目录路径

示例:
  # 使用默认配置运行
  \$0

  # 禁用自动生成
  \$0 --no-auto-generate-go

  # 自定义 Ansible 根目录
  ANSIBLE_ROOT=/path/to/gitlab_cfg \$0

  # 自定义配置文件（如生产环境）
  VARS_FILE=\${ANSIBLE_ROOT}/vars-prod.yaml \$0

  # 完整自定义所有配置
  ANSIBLE_ROOT=/path/to/gitlab_cfg \\
    VARS_FILE=\${ANSIBLE_ROOT}/vars-prod.yaml \\
    MAPPING_FILE=\${ANSIBLE_ROOT}/mapping-prod.yaml \\
    \$0

默认路径:
  Ansible: \${ANSIBLE_OUTPUT_DEFAULT}
  Go:      \${GO_OUTPUT_DEFAULT}

输出文件:
  - 对比数据：\${COMPARISON_DIR}/
  - 详细报告：\${REPORT_DIR}/COMPARISON_REPORT_[时间戳].md
EOF
}

# 解析参数
POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        --no-auto-generate-go)
            AUTO_GENERATE_GO=false
            shift
            ;;
        --*)
            echo "错误：未知参数：$1" >&2
            show_help
            exit 1
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            shift
            ;;
    esac
done

# 处理位置参数
if [ ${#POSITIONAL_ARGS[@]} -eq 0 ]; then
    # 没有提供参数，使用默认值
    ANSIBLE_OUTPUT="${ANSIBLE_OUTPUT_DEFAULT}"
    GO_OUTPUT="${GO_OUTPUT_DEFAULT}"
elif [ ${#POSITIONAL_ARGS[@]} -eq 1 ]; then
    ANSIBLE_OUTPUT="${POSITIONAL_ARGS[0]}"
    GO_OUTPUT="${GO_OUTPUT_DEFAULT}"
elif [ ${#POSITIONAL_ARGS[@]} -eq 2 ]; then
    ANSIBLE_OUTPUT="${POSITIONAL_ARGS[0]}"
    GO_OUTPUT="${POSITIONAL_ARGS[1]}"
elif [ ${#POSITIONAL_ARGS[@]} -gt 2 ]; then
    echo "错误：参数过多" >&2
    show_help
    exit 1
fi

# ============================================
# 显示配置信息
# ============================================
echo "=================================================="
echo "GitLab Cfg 配置生成对比脚本"
echo "=================================================="
echo ""
echo "📋 配置信息:"
echo "  ANSIBLE_ROOT:     ${ANSIBLE_ROOT}"
echo "  VARS_FILE:        ${VARS_FILE}"
echo "  MAPPING_FILE:     ${MAPPING_FILE}"
echo "  RESOURCES_FILE:   ${RESOURCES_FILE}"
echo "  BOOTSTRAP_FILE:   ${BOOTSTRAP_FILE}"
echo "  Go 输出目录：      ${GO_OUTPUT}"
echo "  Ansible 输出：     ${ANSIBLE_OUTPUT}"
echo ""

print_banner() {
    echo -e "${CYAN}"
    echo "================================================================================"
    echo "                    K8s App Accelerator - 对比检测工具 v${VERSION}"
    echo "================================================================================"
    echo -e "${NC}"
    echo "检测时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
}

print_section() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

show_help() {
    cat << EOF
K8s App Accelerator - Go vs Ansible 生成物对比工具 v${VERSION}

用法:
  $0 [OPTIONS] [ANSIBLE_OUTPUT_DIR] [GO_OUTPUT_DIR]

参数:
  ANSIBLE_OUTPUT_DIR    Ansible 生成的输出目录路径
  GO_OUTPUT_DIR         Go 生成的输出目录路径

选项:
  -h, --help                 显示此帮助信息
  --no-auto-generate-go      Go 输出为空时不自动生成

示例:
  # 使用默认路径（Go 输出为空时自动执行 go run main.go generate）
  $0

  # 禁用自动生成
  $0 --no-auto-generate-go

  # 只指定 Ansible 输出目录
  $0 /path/to/ansible/output

  # 指定两个输出目录
  $0 /path/to/ansible/output /path/to/go/output

默认路径:
  Ansible: ${ANSIBLE_OUTPUT_DEFAULT}
  Go:      ${GO_OUTPUT_DEFAULT}

输出文件:
  - 对比数据：${COMPARISON_DIR}/
  - 详细报告：${REPORT_DIR}/COMPARISON_REPORT_[时间戳].md
EOF
}

auto_generate_go_output() {
    print_section "自动生成 Go 输出"

    print_info "检测到 Go 输出目录为空，准备自动执行生成命令"
    print_info "执行目录：${PROJECT_ROOT}"

    mkdir -p "${COMPARISON_DIR}"
    local gen_log="${COMPARISON_DIR}/go_generate_${TIMESTAMP}.log"

    # Go 代码直接使用 Ansible roles 目录中的模板
    # 路径：{baseDir}/roles/{app}/templates/overlays/{profile}/{stack}/*.j2
    # 无需创建符号链接或复制模板文件

    # 先清理旧的 Go 输出（如果有）
    if [ -d "${GO_OUTPUT_DEFAULT}" ]; then
        print_info "清理旧的 Go 输出目录..."
        rm -rf "${GO_OUTPUT_DEFAULT}"/*
    fi

    local cmd="cd \"${PROJECT_ROOT}\" && go run main.go gitlab-cfg generate --base-dir \"${ANSIBLE_ROOT}\" --config ${VARS_FILE} --mapping ${MAPPING_FILE} --resources ${RESOURCES_FILE} --skip-precheck"

    echo "执行命令:"
    echo "  ${cmd}"
    echo ""

    if bash -c "${cmd}" > "${gen_log}" 2>&1; then
        print_success "Go 生成命令执行成功"
        print_info "生成日志：${gen_log}"
        
        # 显示生成的组件统计
        echo ""
        print_info "生成的组件统计:"
        local cms_count=$(find "${GO_OUTPUT_DEFAULT}/cms" -type f 2>/dev/null | wc -l | tr -d ' ')
        local argo_count=$(find "${GO_OUTPUT_DEFAULT}/argo-app" -type f 2>/dev/null | wc -l | tr -d ' ')
        local jenkins_count=$(find "${GO_OUTPUT_DEFAULT}/jenkins-job" -type f 2>/dev/null | wc -l | tr -d ' ')
        
        echo "  - CMS K8s 配置：${cms_count} 个文件"
        echo "  - ArgoCD Application: ${argo_count} 个文件"
        echo "  - Jenkins Job: ${jenkins_count} 个文件"
    else
        print_error "Go 生成命令执行失败"
        print_info "请查看日志：${gen_log}"
        echo ""
        print_info "你可以手动执行以下命令重试："
        echo "  cd \"${PROJECT_ROOT}\""
        echo "  go run main.go gitlab-cfg generate --base-dir \"${ANSIBLE_ROOT}\" --config ${VARS_FILE} --mapping ${MAPPING_FILE} --resources ${RESOURCES_FILE} --skip-precheck"
        echo ""
        print_info "模板目录说明："
        echo "  Go 使用：{baseDir}/roles/{app}/templates/overlays/{profile}/{stack}/"
        echo "  直接使用 Ansible roles 中的模板文件，无需复制或符号链接"
        exit 1
    fi

    echo ""
}

check_prerequisites() {
    print_section "检查前置条件"
    
    print_info "使用以下路径进行对比:"
    echo "  Ansible 输出: ${ANSIBLE_OUTPUT}"
    echo "  Go 输出:       ${GO_OUTPUT}"
    echo ""
    
    # 检查 Ansible 输出目录
    if [ ! -d "${ANSIBLE_OUTPUT}" ]; then
        print_warning "Ansible 输出目录不存在：${ANSIBLE_OUTPUT}"
        print_info "尝试创建目录..."
        if mkdir -p "${ANSIBLE_OUTPUT}" 2>/dev/null; then
            print_success "目录已创建：${ANSIBLE_OUTPUT}"
        else
            print_error "无法创建目录：${ANSIBLE_OUTPUT}"
            exit 1
        fi
    fi
    
    # 检查 Ansible 输出目录是否有内容
    ANSIBLE_FILE_COUNT=$(find "${ANSIBLE_OUTPUT}" -type f 2>/dev/null | wc -l | tr -d ' ')
    if [ "${ANSIBLE_FILE_COUNT}" -eq 0 ]; then
        print_error "Ansible 输出目录为空：${ANSIBLE_OUTPUT}"
        print_info "请先运行 Ansible 生成配置文件"
        echo ""
        print_info "生成 Ansible 输出的命令:"
        echo "  cd ${ANSIBLE_ROOT}"
        echo "  ansible-playbook ${BOOTSTRAP_FILE} -e '@${VARS_FILE}'"
        exit 1
    fi
    
    print_success "Ansible 输出目录已就绪：${ANSIBLE_OUTPUT}"
    print_info "Ansible 生成的文件数量：${ANSIBLE_FILE_COUNT} 个"
    
    # 显示 Ansible 输出目录结构预览
    echo ""
    print_info "Ansible 输出目录结构预览:"
    if command -v tree &> /dev/null; then
        tree -L 4 "${ANSIBLE_OUTPUT}" 2>/dev/null || find "${ANSIBLE_OUTPUT}" -type d | head -10
    else
        find "${ANSIBLE_OUTPUT}" -type d | head -10
    fi
    
    # 检查 Go 输出目录
    if [ ! -d "${GO_OUTPUT}" ]; then
        print_warning "Go 输出目录不存在：${GO_OUTPUT}"
        print_info "尝试创建目录..."
        if mkdir -p "${GO_OUTPUT}" 2>/dev/null; then
            print_success "目录已创建：${GO_OUTPUT}"
        else
            print_error "无法创建目录：${GO_OUTPUT}"
            exit 1
        fi
    fi
    
    # 检查 Go 输出目录是否有内容
    GO_FILE_COUNT=$(find "${GO_OUTPUT}" -type f 2>/dev/null | wc -l | tr -d ' ')
    if [ "${GO_FILE_COUNT}" -eq 0 ]; then
        print_warning "Go 输出目录为空：${GO_OUTPUT}"

        if [ "${AUTO_GENERATE_GO}" = true ]; then
            auto_generate_go_output
            GO_FILE_COUNT=$(find "${GO_OUTPUT}" -type f 2>/dev/null | wc -l | tr -d ' ')
        fi

        if [ "${GO_FILE_COUNT}" -eq 0 ]; then
            print_error "Go 输出目录仍为空：${GO_OUTPUT}"
            print_info "请手动运行 Go 生成命令"
            exit 1
        fi
    fi
    
    print_success "Go 输出目录已就绪：${GO_OUTPUT}"
    print_info "Go 生成的文件数量：${GO_FILE_COUNT} 个"
    
    # 显示 Go 输出目录结构预览
    echo ""
    print_info "Go 输出目录结构预览:"
    if command -v tree &> /dev/null; then
        tree -L 4 -I 'comparison' "${GO_OUTPUT}" 2>/dev/null || find "${GO_OUTPUT}" -type d | head -10
    else
        find "${GO_OUTPUT}" -type d | head -10
    fi
    
    # 创建对比目录
    print_info "创建对比目录..."
    if [ -d "${COMPARISON_DIR}" ]; then
        print_warning "对比目录已存在，清理旧文件：${COMPARISON_DIR}"
        rm -rf "${COMPARISON_DIR}"
    fi
    if mkdir -p "${COMPARISON_DIR}"; then
        print_success "对比目录已创建：${COMPARISON_DIR}"
    else
        print_error "无法创建对比目录：${COMPARISON_DIR}"
        exit 1
    fi
    
    echo ""
}

compare_directory_structure() {
    print_section "对比目录结构"
    
    echo -e "${CYAN}Ansible 输出目录结构:${NC}"
    echo "─────────────────────────────────────────"
    if command -v tree &> /dev/null; then
        tree -L 3 "${ANSIBLE_OUTPUT}" 2>/dev/null || find "${ANSIBLE_OUTPUT}" -type d | head -20
    else
        find "${ANSIBLE_OUTPUT}" -type d | head -20
    fi
    
    echo ""
    echo -e "${CYAN}Go 输出目录结构:${NC}"
    echo "─────────────────────────────────────────"
    if command -v tree &> /dev/null; then
        tree -L 3 "${GO_OUTPUT}" 2>/dev/null || find "${GO_OUTPUT}" -type d | head -20
    else
        find "${GO_OUTPUT}" -type d | head -20
    fi
    
    echo ""
}

compare_file_counts() {
    print_section "统计文件数量"
    
    TOTAL_FILES_ANSIBLE=$(find "${ANSIBLE_OUTPUT}" -type f | wc -l | tr -d ' ')
    echo -e "${CYAN}Ansible 生成文件总数: ${TOTAL_FILES_ANSIBLE}${NC}"
    
    TOTAL_FILES_GO=$(find "${GO_OUTPUT}" -type f | wc -l | tr -d ' ')
    echo -e "${CYAN}Go 生成文件总数: ${TOTAL_FILES_GO}${NC}"
    
    echo ""
    echo -e "${BLUE}Ansible 文件分类:${NC}"
    find "${ANSIBLE_OUTPUT}" -type f -name "*.yaml" 2>/dev/null | wc -l | xargs printf "  - YAML 文件: %s\n"
    find "${ANSIBLE_OUTPUT}" -type f -name "*.yml" 2>/dev/null | wc -l | xargs printf "  - YML 文件: %s\n"
    find "${ANSIBLE_OUTPUT}" -type f -name "*.env" 2>/dev/null | wc -l | xargs printf "  - ENV 文件: %s\n"
    find "${ANSIBLE_OUTPUT}" -type f -name "*.j2" 2>/dev/null | wc -l | xargs printf "  - J2 文件: %s\n"
    
    echo ""
    echo -e "${BLUE}Go 文件分类:${NC}"
    find "${GO_OUTPUT}" -type f -name "*.yaml" 2>/dev/null | wc -l | xargs printf "  - YAML 文件: %s\n"
    find "${GO_OUTPUT}" -type f -name "*.yml" 2>/dev/null | wc -l | xargs printf "  - YML 文件: %s\n"
    find "${GO_OUTPUT}" -type f -name "*.env" 2>/dev/null | wc -l | xargs printf "  - ENV 文件: %s\n"
    find "${GO_OUTPUT}" -type f -name "*.j2" 2>/dev/null | wc -l | xargs printf "  - J2 文件: %s\n"
    
    echo ""
}

compare_file_lists() {
    print_section "对比文件清单"
    
    local ansible_list="${COMPARISON_DIR}/ansible_files_${TIMESTAMP}.txt"
    local go_list="${COMPARISON_DIR}/go_files_${TIMESTAMP}.txt"
    
    find "${ANSIBLE_OUTPUT}" -type f | sed "s|${ANSIBLE_OUTPUT}/||" | sort > "${ansible_list}"
    find "${GO_OUTPUT}" -type f | sed "s|${GO_OUTPUT}/||" | sort > "${go_list}"
    
    print_info "文件清单已保存:"
    echo "  - Ansible: ${ansible_list}"
    echo "  - Go: ${go_list}"
    
    local diff_file="${COMPARISON_DIR}/file_list_diff_${TIMESTAMP}.txt"
    
    echo ""
    echo -e "${CYAN}文件列表差异:${NC}"
    echo "─────────────────────────────────────────"
    
    if diff -u "${ansible_list}" "${go_list}" > "${diff_file}"; then
        print_success "文件列表完全一致！"
    else
        print_warning "文件列表存在差异"
        echo ""
        cat "${diff_file}"
        
        local missing_in_go=$(grep "^-" "${diff_file}" | grep -v "^---" | wc -l | tr -d ' ')
        local missing_in_ansible=$(grep "^+" "${diff_file}" | grep -v "^+++" | wc -l | tr -d ' ')
        
        echo ""
        echo -e "${YELLOW}Go 实现缺失的文件数: ${missing_in_go}${NC}"
        echo -e "${YELLOW}Go 实现多余的文件数: ${missing_in_ansible}${NC}"
    fi
    
    echo ""
}

# ============================================
# 智能文件对比函数
# 支持忽略：
# 1. 末尾空行差异（1个或多个换行符）
# 2. UUID/随机Hex值（如 dbupgrade-xxx-3e7bd6e）
# ============================================
smart_compare_files() {
    local file1="$1"
    local file2="$2"
    
    # 创建临时文件
    local tmp1=$(mktemp)
    local tmp2=$(mktemp)
    
    # 预处理文件1：去除末尾空行
    sed -E ':a; /^\n*$/d; /\n$/!b; N; ba' "${file1}" > "${tmp1}" 2>/dev/null || cat "${file1}" > "${tmp1}"
    
    # 预处理文件2：去除末尾空行  
    sed -E ':a; /^\n*$/d; /\n$/!b; N; ba' "${file2}" > "${tmp2}" 2>/dev/null || cat "${file2}" > "${tmp2}"
    
    # 去除末尾空行（更简单的方法）
    sed -i '/^[[:space:]]*$/d' "${tmp1}" 2>/dev/null || true
    sed -i '/^[[:space:]]*$/d' "${tmp2}" 2>/dev/null || true
    
    # 去除行尾空格
    sed -i 's/[[:space:]]*$//' "${tmp1}" 2>/dev/null || true
    sed -i 's/[[:space:]]*$//' "${tmp2}" 2>/dev/null || true
    
    # 标准化 UUID/Hex 随机值（用于 job.yaml 中的 name 字段）
    # 匹配模式：dbupgrade-xxx-随机7位hex
    # 例如：dbupgrade-cms-service-3e7bd6e -> dbupgrade-cms-service-XXXXXXX
    local ntmp1=$(mktemp)
    local ntmp2=$(mktemp)
    
    # 标准化 file1 中的随机值
    sed -E 's/(dbupgrade-[a-zA-Z0-9-]+-)[0-9a-f]{7}/\1XXXXXXX/g' "${tmp1}" > "${ntmp1}"
    
    # 标准化 file2 中的随机值
    sed -E 's/(dbupgrade-[a-zA-Z0-9-]+-)[0-9a-f]{7}/\1XXXXXXX/g' "${tmp2}" > "${ntmp2}"
    
    # 对比预处理后的文件
    local result=0
    cmp -s "${ntmp1}" "${ntmp2}" || result=$?
    
    # 清理临时文件
    rm -f "${tmp1}" "${tmp2}" "${ntmp1}" "${ntmp2}"
    
    return ${result}
}

# 获取智能对比的详细差异
smart_diff_files() {
    local file1="$1"
    local file2="$2"
    
    # 创建临时文件
    local tmp1=$(mktemp)
    local tmp2=$(mktemp)
    
    # 预处理：去除末尾空行和行尾空格
    sed '/^[[:space:]]*$/d' "${file1}" | sed 's/[[:space:]]*$//' > "${tmp1}"
    sed '/^[[:space:]]*$/d' "${file2}" | sed 's/[[:space:]]*$//' > "${tmp2}"
    
    # 标准化随机值
    local ntmp1=$(mktemp)
    local ntmp2=$(mktemp)
    sed -E 's/(dbupgrade-[a-zA-Z0-9-]+-)[0-9a-f]{7}/\1XXXXXXX/g' "${tmp1}" > "${ntmp1}"
    sed -E 's/(dbupgrade-[a-zA-Z0-9-]+-)[0-9a-f]{7}/\1XXXXXXX/g' "${tmp2}" > "${ntmp2}"
    
    # 输出差异
    diff -u "${ntmp1}" "${ntmp2}"
    
    # 清理
    rm -f "${tmp1}" "${tmp2}" "${ntmp1}" "${ntmp2}"
}

# ============================================
# 内容对比函数
# ============================================
compare_file_contents() {
    print_section "对比文件内容 (智能模式)"
    print_info "忽略规则: 末尾空行差异、UUID/随机Hex值差异"
    echo ""
    
    local identical_file="${COMPARISON_DIR}/identical_files_${TIMESTAMP}.txt"
    local different_file="${COMPARISON_DIR}/different_files_${TIMESTAMP}.txt"
    local missing_file="${COMPARISON_DIR}/missing_files_${TIMESTAMP}.txt"
    local diff_detail="${COMPARISON_DIR}/content_diff_${TIMESTAMP}.txt"
    local tolerated_file="${COMPARISON_DIR}/tolerated_differences_${TIMESTAMP}.txt"
    
    > "${identical_file}"
    > "${different_file}"
    > "${missing_file}"
    > "${diff_detail}"
    > "${tolerated_file}"
    
    local identical_count=0
    local different_count=0
    local missing_count=0
    local tolerated_count=0
    
    echo -e "${CYAN}开始逐文件对比...${NC}"
    echo ""
    
    # 遍历 Ansible 输出的所有文件
    while IFS= read -r ansible_f; do
        rel_path=$(echo "${ansible_f}" | sed "s|${ANSIBLE_OUTPUT}/||")
        go_f="${GO_OUTPUT}/${rel_path}"
        
        if [ ! -f "${go_f}" ]; then
            echo -e "${RED}缺失${NC}  ${rel_path}"
            echo "${rel_path}" >> "${missing_file}"
            missing_count=$((missing_count + 1))
        else
            # 临时禁用 set -e 以防止 cmp -s 在文件不同时退出
            set +e
            if smart_compare_files "${ansible_f}" "${go_f}"; then
                echo -e "${GREEN}相同${NC}  ${rel_path}"
                echo "${rel_path}" >> "${identical_file}"
                identical_count=$((identical_count + 1))
            else
                # 文件有差异，检查是否是可容忍的差异
                local diff_output
                diff_output=$(smart_diff_files "${ansible_f}" "${go_f}" 2>&1)
                
                if [ -z "${diff_output}" ]; then
                    # 差异仅在末尾空行或随机值，可容忍
                    echo -e "${GREEN}容错${NC}  ${rel_path} (仅末尾空行/随机值差异)"
                    echo "${rel_path}" >> "${tolerated_file}"
                    tolerated_count=$((tolerated_count + 1))
                else
                    echo -e "${YELLOW}差异${NC}  ${rel_path}"
                    echo "${rel_path}" >> "${different_file}"
                    different_count=$((different_count + 1))
                    
                    echo "========== ${rel_path} ==========" >> "${diff_detail}"
                    echo "${diff_output}" >> "${diff_detail}"
                    echo "" >> "${diff_detail}"
                fi
            fi
            # 恢复 set -e
            set -e
        fi
    done < <(find "${ANSIBLE_OUTPUT}" -type f | sort)
    
    # 检查 Go 中多余而 Ansible 没有的文件
    echo ""
    print_info "检查 Go 中多余的文件..."
    while IFS= read -r go_f; do
        rel_path=$(echo "${go_f}" | sed "s|${GO_OUTPUT}/||")
        ansible_f="${ANSIBLE_OUTPUT}/${rel_path}"
        
        if [ ! -f "${ansible_f}" ]; then
            echo -e "${YELLOW}多余${NC}  ${rel_path} (仅 Go 生成的文件)"
            echo "${rel_path}" >> "${different_file}"
            different_count=$((different_count + 1))
        fi
    done < <(find "${GO_OUTPUT}" -type f | sort)
    
    # 更新全局统计变量
    IDENTICAL_FILES=${identical_count}
    TOLERATED_FILES=${tolerated_count}
    DIFFERENT_FILES=${different_count}
    MISSING_FILES=${missing_count}
    
    echo ""
    echo -e "${CYAN}内容对比统计:${NC}"
    echo "─────────────────────────────────────────"
    print_success "完全一致的文件: ${IDENTICAL_FILES}"
    print_success "容错一致的文件: ${TOLERATED_FILES} (仅末尾空行/随机值差异)"
    print_warning "内容有差异的文件: ${DIFFERENT_FILES}"
    print_error "缺失的文件: ${MISSING_FILES}"
    
    echo ""
    print_info "详细对比结果已保存:"
    echo "  - 相同文件列表: ${identical_file}"
    echo "  - 容错一致文件列表: ${tolerated_file}"
    echo "  - 差异文件列表: ${different_file}"
    echo "  - 缺失文件列表: ${missing_file}"
    echo "  - 详细差异内容: ${diff_detail}"
    
    echo ""
}

analyze_key_differences() {
    print_section "关键差异分析"
    
    local diff_list="${COMPARISON_DIR}/different_files_${TIMESTAMP}.txt"
    
    if [ ! -s "${diff_list}" ]; then
        print_success "没有发现内容差异！"
        return
    fi
    
    echo -e "${CYAN}分析文件内容差异的关键问题:${NC}"
    echo "─────────────────────────────────────────"
    
    # 检查 namespace 问题
    echo ""
    print_info "1. namespace 变量渲染检查:"
    if grep -r "jinja2.utils.Namespace" "${GO_OUTPUT}" 2>/dev/null > /dev/null; then
        print_error "发现 Jinja2 namespace 保留字问题"
    else
        print_success "namespace 变量渲染正确"
    fi
    
    # 检查 harbor_project 问题
    echo ""
    print_info "2. harbor_project 变量渲染检查:"
    if grep -r "{{harbor_project}}" "${GO_OUTPUT}" 2>/dev/null > /dev/null; then
        print_error "发现 harbor_project 变量未渲染问题"
        echo "  影响文件:"
        grep -r "{{harbor_project}}" "${GO_OUTPUT}" 2>/dev/null | cut -d: -f1 | sort -u | sed 's/^/    - /'
    else
        print_success "harbor_project 变量已正确渲染"
    fi
    
    # 检查多余空行问题
    echo ""
    print_info "3. 多余空行检查:"
    for f in $(find "${GO_OUTPUT}" -name "hpa.yaml" -o -name "job.yaml" -o -name "config.yaml" 2>/dev/null); do
        if [ -f "${f}" ]; then
            first_line=$(head -1 "${f}")
            if [ -z "${first_line}" ]; then
                print_warning "$(basename ${f}) 开头有多余空行"
            fi
        fi
    done
    
    echo ""
}

check_missing_components() {
    print_section "缺失组件检查"
    
    # 检查 ArgoCD 配置
    echo -e "${CYAN}1. ArgoCD Application 配置:${NC}"
    echo "─────────────────────────────────────────"
    local argo_ansible=$(find "${ANSIBLE_OUTPUT}" -path "*/argo-app/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    local argo_go=$(find "${GO_OUTPUT}" -path "*/argo-app/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${argo_go}" -eq 0 ] && [ "${argo_ansible}" -gt 0 ]; then
        print_error "Go 实现缺少 ArgoCD Application 配置生成"
        print_info "Ansible 生成了 ${argo_ansible} 个 ArgoCD 配置文件"
    elif [ "${argo_go}" -gt 0 ]; then
        print_success "ArgoCD 配置已生成：${argo_go} 个文件"
    else
        print_info "未检测到 ArgoCD 配置（可能不需要）"
    fi
    
    # 检查 Jenkins 配置
    echo ""
    echo -e "${CYAN}2. Jenkins Job 配置:${NC}"
    echo "─────────────────────────────────────────"
    local jenkins_ansible=$(find "${ANSIBLE_OUTPUT}" -path "*/jenkins-job/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    local jenkins_go=$(find "${GO_OUTPUT}" -path "*/jenkins-job/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${jenkins_go}" -eq 0 ] && [ "${jenkins_ansible}" -gt 0 ]; then
        print_error "Go 实现缺少 Jenkins Job 配置生成"
        print_info "Ansible 生成了 ${jenkins_ansible} 个 Jenkins 配置文件"
    elif [ "${jenkins_go}" -gt 0 ]; then
        print_success "Jenkins 配置已生成：${jenkins_go} 个文件"
    else
        print_info "未检测到 Jenkins 配置（可能不需要）"
    fi
    
    # 检查 CMS 配置（核心 K8s 资源）
    echo ""
    echo -e "${CYAN}3. CMS K8s 资源配置:${NC}"
    echo "─────────────────────────────────────────"
    local cms_ansible=$(find "${ANSIBLE_OUTPUT}" -path "*/cms/cms-service/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    local cms_go=$(find "${GO_OUTPUT}" -path "*/cms/cms-service/*" -type f 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${cms_go}" -eq 0 ] && [ "${cms_ansible}" -gt 0 ]; then
        print_error "Go 实现缺少 CMS K8s 资源配置生成"
        print_info "Ansible 生成了 ${cms_ansible} 个 CMS 配置文件"
    elif [ "${cms_go}" -gt 0 ]; then
        print_success "CMS 配置已生成：${cms_go} 个文件"
        
        # 检查关键文件是否都存在
        local expected_files=("deployment.yaml" "service.yaml" "config.yaml" "hpa.yaml" "job.yaml" "kustomization.yaml")
        local missing_files=()
        
        for file in "${expected_files[@]}"; do
            if ! find "${GO_OUTPUT}" -path "*/cms/cms-service/*" -name "${file}" | grep -q .; then
                missing_files+=("${file}")
            fi
        done
        
        if [ ${#missing_files[@]} -gt 0 ]; then
            print_warning "缺少以下关键文件：${missing_files[*]}"
        else
            print_success "所有关键配置文件都已生成"
        fi
    else
        print_info "未检测到 CMS 配置（可能不需要）"
    fi
    
    echo ""
}

generate_summary_report() {
    print_section "生成汇总报告"
    
    local report_file="${REPORT_DIR}/COMPARISON_REPORT_${TIMESTAMP}.md"
    
    cat > "${report_file}" << EOF
# Go vs Ansible 生成物对比报告

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')
**Ansible 目录**: ${ANSIBLE_OUTPUT}
**Go 目录**: ${GO_OUTPUT}

---

## 📊 对比结果汇总

### 文件统计

| 指标 | 数量 |
|------|------|
| Ansible 文件总数 | ${TOTAL_FILES_ANSIBLE} |
| Go 文件总数 | ${TOTAL_FILES_GO} |
| 完全一致的文件 | ${IDENTICAL_FILES} |
| 内容有差异的文件 | ${DIFFERENT_FILES} |
| 缺失的文件 | ${MISSING_FILES} |

---

## 📁 详细文件清单

### ✅ 完全一致的文件 (${IDENTICAL_FILES} 个)

EOF

    if [ -s "${COMPARISON_DIR}/identical_files_${TIMESTAMP}.txt" ]; then
        while IFS= read -r line; do
            echo "- \`${line}\`" >> "${report_file}"
        done < "${COMPARISON_DIR}/identical_files_${TIMESTAMP}.txt"
    else
        echo "无" >> "${report_file}"
    fi
    
    cat >> "${report_file}" << EOF

### ⚠️ 内容有差异的文件 (${DIFFERENT_FILES} 个)

EOF

    if [ -s "${COMPARISON_DIR}/different_files_${TIMESTAMP}.txt" ]; then
        while IFS= read -r line; do
            echo "- \`${line}\`" >> "${report_file}"
        done < "${COMPARISON_DIR}/different_files_${TIMESTAMP}.txt"
    else
        echo "无" >> "${report_file}"
    fi
    
    cat >> "${report_file}" << EOF

### ❌ 缺失的文件 (${MISSING_FILES} 个)

EOF

    if [ -s "${COMPARISON_DIR}/missing_files_${TIMESTAMP}.txt" ]; then
        while IFS= read -r line; do
            echo "- \`${line}\`" >> "${report_file}"
        done < "${COMPARISON_DIR}/missing_files_${TIMESTAMP}.txt"
    else
        echo "无" >> "${report_file}"
    fi
    
    cat >> "${report_file}" << EOF

---

## 🔗 相关文件

- 详细差异: comparison/content_diff_${TIMESTAMP}.txt
- 文件清单 (Ansible): comparison/ansible_files_${TIMESTAMP}.txt
- 文件清单 (Go): comparison/go_files_${TIMESTAMP}.txt

---

**生成工具**: K8s App Accelerator 对比脚本 v${VERSION}
EOF

    print_success "汇总报告已生成: ${report_file}"
    
    cp "${report_file}" "${REPORT_DIR}/COMPARISON_LATEST.md"
    print_info "最新报告链接: ${REPORT_DIR}/COMPARISON_LATEST.md"
    
    echo ""
}

print_final_summary() {
    print_section "对比完成总结"
    
    echo -e "${CYAN}📊 对比统计:${NC}"
    echo "─────────────────────────────────────────"
    echo "  Ansible 文件总数: ${TOTAL_FILES_ANSIBLE}"
    echo "  Go 文件总数:      ${TOTAL_FILES_GO}"
    echo ""
    echo -e "  ${GREEN}✅ 完全一致:      ${IDENTICAL_FILES} 个${NC}"
    echo -e "  ${GREEN}⚠️  容错一致:      ${TOLERATED_FILES} 个 (末尾空行/随机值)${NC}"
    echo -e "  ${YELLOW}⚠️  内容差异:      ${DIFFERENT_FILES} 个${NC}"
    echo -e "  ${RED}❌ 缺失文件:      ${MISSING_FILES} 个${NC}"
    echo ""
    
    if [ "${TOTAL_FILES_GO}" -gt 0 ]; then
        local total_matched=$((IDENTICAL_FILES + TOLERATED_FILES))
        local consistency=$(awk "BEGIN {printf \"%.1f\", (${total_matched}/${TOTAL_FILES_GO})*100}")
        echo -e "  ${GREEN}一致性比例:      ${consistency}%${NC}"
    fi
    
    echo ""
    echo -e "${CYAN}📁 生成的文件:${NC}"
    echo "─────────────────────────────────────────"
    echo "  对比数据目录: ${COMPARISON_DIR}/"
    echo "  报告文件:     ${REPORT_DIR}/COMPARISON_LATEST.md"
    echo ""
    
    echo -e "${CYAN}💡 下一步建议:${NC}"
    echo "─────────────────────────────────────────"
    
    if [ ${MISSING_FILES} -gt 0 ]; then
        print_error "发现 ${MISSING_FILES} 个缺失文件，建议优先实现"
        echo -e "  ${YELLOW}查看缺失文件: cat ${COMPARISON_DIR}/missing_files_${TIMESTAMP}.txt${NC}"
    fi
    
    if [ ${DIFFERENT_FILES} -gt 0 ]; then
        print_warning "发现 ${DIFFERENT_FILES} 个文件内容有差异"
        echo -e "  ${YELLOW}查看详细差异: cat ${COMPARISON_DIR}/content_diff_${TIMESTAMP}.txt${NC}"
    fi
    
    if [ ${MISSING_FILES} -eq 0 ] && [ ${DIFFERENT_FILES} -eq 0 ]; then
        print_success "所有文件完全一致！Go 实现已达到生产就绪状态"
    fi
    
    echo ""
}

################################################################################
# 参数解析和主流程
################################################################################

# 解析参数
POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        --no-auto-generate-go)
            AUTO_GENERATE_GO=false
            shift
            ;;
        --*)
            echo "错误: 未知参数: $1" >&2
            show_help
            exit 1
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            shift
            ;;
    esac
done

# 处理位置参数
if [ ${#POSITIONAL_ARGS[@]} -eq 0 ]; then
    # 没有提供参数，使用默认值
    ANSIBLE_OUTPUT="${ANSIBLE_OUTPUT_DEFAULT}"
    GO_OUTPUT="${GO_OUTPUT_DEFAULT}"
elif [ ${#POSITIONAL_ARGS[@]} -eq 1 ]; then
    ANSIBLE_OUTPUT="${POSITIONAL_ARGS[0]}"
    GO_OUTPUT="${GO_OUTPUT_DEFAULT}"
elif [ ${#POSITIONAL_ARGS[@]} -eq 2 ]; then
    ANSIBLE_OUTPUT="${POSITIONAL_ARGS[0]}"
    GO_OUTPUT="${POSITIONAL_ARGS[1]}"
elif [ ${#POSITIONAL_ARGS[@]} -gt 2 ]; then
    echo "错误: 参数过多" >&2
    show_help
    exit 1
fi

# 主流程
main() {
    print_banner
    
    check_prerequisites
    compare_directory_structure
    compare_file_counts
    compare_file_lists
    compare_file_contents
    analyze_key_differences
    check_missing_components
    generate_summary_report
    print_final_summary
    
    echo -e "${GREEN}✅ 对比检测完成！${NC}"
    echo ""
}

# 运行主流程
main "$@"
