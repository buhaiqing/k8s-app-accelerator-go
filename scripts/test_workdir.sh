#!/bin/bash
# 测试工作目录功能

set -e

echo "╔════════════════════════════════════════════════════╗"
echo "║       测试工作目录功能 - cms-service 示例         ║"
echo "╚════════════════════════════════════════════════════╝"
echo ""

# 设置测试目录
TEST_WORKDIR="/tmp/k8s_gen_test_$$"
echo "创建测试工作目录：$TEST_WORKDIR"
mkdir -p "$TEST_WORKDIR"

# 清理函数
cleanup() {
    echo ""
    echo "清理测试目录..."
    rm -rf "$TEST_WORKDIR"
}

# 注册清理函数
trap cleanup EXIT

# 1. 创建 bootstrap.yml
echo "1. 创建 bootstrap.yml..."
cat > "$TEST_WORKDIR/bootstrap.yml" <<EOF
roles:
  - cms-service
  - fms-service
EOF
echo "   ✓ bootstrap.yml 创建完成"

# 2. 创建 vars.yaml
echo "2. 创建 vars.yaml..."
cat > "$TEST_WORKDIR/vars.yaml" <<EOF
rootdir: $TEST_WORKDIR/output
project: test-project
profiles:
  - int
  - production

ssl_secret_name: test-ssl-secret

apollo:
  site: https://apollo.example.com
  customerid: "12345"
  env: prod
  alias: default
  token: test-token

argocd:
  site: https://argocd.example.com
  cluster: test-cluster

jenkins:
  site: https://jenkins.example.com

stack:
  cms-service: baas
  fms-service: baas

toolset_git_base_url: https://git.example.com
toolset_git_group: my-group
toolset_git_project: test-project

cluster_id: test-cluster-id
jira_id: TEST-123
EOF
echo "   ✓ vars.yaml 创建完成"

# 3. 创建 resources.yaml
echo "3. 创建 resources.yaml..."
cat > "$TEST_WORKDIR/resources.yaml" <<EOF
rds:
  - name: default
    datasource_url: rm-test.mysql.rds.aliyuncs.com
    datasource_db: testdb
    datasource_port: "3306"
    datasource_username: admin
    datasource_password: test-password

redis:
  - name: default
    redisIp: r-test.redis.rds.aliyuncs.com
    redisPort: "6379"
    redisDb: "0"
    redisPassword: test-redis-password

oss:
  - name: default
    oss_roleArn: acs:ram::123456789:role/oss-role
    oss_stsEndpoint: sts.cn-hangzhou.aliyuncs.com
    bucketName: test-bucket
    ossEndpoint: oss-cn-hangzhou.aliyuncs.com
    ossimageBaseUrl: https://test-bucket.oss-cn-hangzhou.aliyuncs.com
    ossinternalEndpoint: oss-cn-hangzhou-internal.aliyuncs.com
    oss_accessKeyId: test-access-key
    oss_accessKeySecret: test-access-key-secret
EOF
echo "   ✓ resources.yaml 创建完成"

# 4. 创建 mapping.yaml
echo "4. 创建 mapping.yaml..."
cat > "$TEST_WORKDIR/mapping.yaml" <<EOF
mappings:
  cms-service: cms
  fms-service: fms
EOF
echo "   ✓ mapping.yaml 创建完成"

# 5. 创建简单的模板（用于测试）
echo "5. 创建测试模板..."
mkdir -p "$TEST_WORKDIR/templates/cms/templates/base"
cat > "$TEST_WORKDIR/templates/cms/templates/base/deployment.yaml.j2" <<'TEMPLATE'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ app }}
  namespace: {{ profile | default("default") }}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: {{ app }}
  template:
    metadata:
      labels:
        app: {{ app }}
    spec:
      containers:
      - name: {{ app }}
        image: {{ app }}:latest
        env:
        - name: APOLLO_APP_ID
          value: {{ app }}
        - name: PROFILE
          value: {{ profile | upper }}
TEMPLATE
echo "   ✓ 测试模板创建完成"

# 6. 运行生成器
echo ""
echo "6. 运行配置生成器..."
echo "   命令：./build/k8s-gen generate --workdir $TEST_WORKDIR"
echo ""

cd "$(dirname "$0")/.."
./build/k8s-gen generate \
  --workdir "$TEST_WORKDIR" \
  --output output

if [ $? -eq 0 ]; then
    echo ""
    echo "   ✓ 生成成功！"
else
    echo ""
    echo "   ✗ 生成失败！"
    exit 1
fi

# 7. 查看生成的文件
echo ""
echo "7. 查看生成的文件结构:"
echo "   └── output/"
if [ -d "$TEST_WORKDIR/output" ]; then
    find "$TEST_WORKDIR/output" -type f | while read file; do
        echo "      ├── $(basename $file)"
    done
else
    echo "      (空)"
fi

# 8. 显示生成的内容示例
echo ""
echo "8. 生成的 deployment.yaml 内容示例:"
DEPLOYMENT_FILE=$(find "$TEST_WORKDIR/output" -name "deployment.yaml" | head -n 1)
if [ -f "$DEPLOYMENT_FILE" ]; then
    echo "   ┌─────────────────────────────────────"
    head -n 20 "$DEPLOYMENT_FILE" | sed 's/^/   │ /'
    echo "   └─────────────────────────────────────"
else
    echo "   (未找到 deployment.yaml)"
fi

echo ""
echo "╔════════════════════════════════════════════════════╗"
echo "║              测试完成！✅                          ║"
echo "╚════════════════════════════════════════════════════╝"
echo ""
echo "提示：测试目录将在退出时自动清理"
echo ""
