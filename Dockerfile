# K8s App Accelerator Go - Dockerfile
# 多阶段构建，最小化镜像体积

# ==============================================================================
# 阶段 1: 构建 Go 程序
# ==============================================================================
FROM golang:1.25-alpine AS builder-go

# 设置工作目录
WORKDIR /build

# 安装必要的工具
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go main.go
COPY cmd/ cmd/
COPY internal/ internal/

# 编译二进制文件（静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/k8s-gen \
    main.go

# ==============================================================================
# 阶段 2: 准备 Python 环境
# ==============================================================================
FROM python:3.12-alpine AS builder-python

WORKDIR /build

# 复制 Python 脚本
COPY scripts/ scripts/

# 安装 Python 依赖
RUN pip install --no-cache-dir -r scripts/requirements.txt

# ==============================================================================
# 阶段 3: 最终运行镜像
# ==============================================================================
FROM alpine:3.19

# 添加标签信息
LABEL maintainer="DevOps Team"
LABEL description="K8s App Accelerator Go - K8s 配置生成器"
LABEL version="1.0.0"

# 安装运行时依赖
RUN apk add --no-cache \
    python3 \
    py3-pip \
    ca-certificates

# 创建非 root 用户
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder-go /build/k8s-gen /app/k8s-gen
COPY --from=builder-python /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --chown=appuser:appgroup scripts/ /app/scripts/
COPY --chown=appuser:appgroup configs/ /app/configs/

# 设置执行权限
RUN chmod +x /app/k8s-gen

# 切换到非 root 用户
USER appuser

# 设置环境变量
ENV PYTHONUNBUFFERED=1
ENV LOG_LEVEL=INFO

# 卷挂载点（用于挂载配置文件和输出目录）
VOLUME ["/app/output", "/app/configs"]

# 入口点
ENTRYPOINT ["/app/k8s-gen"]

# 默认参数
CMD ["--help"]

# ==============================================================================
# 使用说明
# ==============================================================================
# 
# 构建镜像:
#   docker build -t k8s-app-accelerator-go:1.0.0 .
#
# 运行容器:
#   docker run --rm -v $(pwd)/configs:/app/configs -v $(pwd)/output:/app/output k8s-app-accelerator-go:1.0.0 generate --config /app/configs/vars.yaml
#
# 开发模式:
#   docker run --rm -it -v $(pwd):/app k8s-app-accelerator-go:1.0.0 /bin/sh
#
# ==============================================================================
