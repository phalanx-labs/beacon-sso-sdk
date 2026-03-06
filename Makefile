# 变量定义
PROTO_FILE ?= client/proto/beacon/sso/v1/public.proto
BASE_GO_MODULE_DIR := /Users/xiaolfeng/ProgramProjects/Cooperate/bamboo-service/bamboo-base/plugins/grpc
XBASE_LINK := client/proto/link/base.proto

# 获取版本号（去除 v 前缀）
VERSION := $(shell cat version | sed 's/^v//')

# 获取当前时间戳（格式：YYYYMMDDHHMM）
TIMESTAMP := $(shell date +"%Y%m%d%H%M")

# 完整 tag 名称
TAG_NAME := v$(VERSION)-$(TIMESTAMP)

.DEFAULT_GOAL := help

.PHONY: help proto proto-all proto-init connect-install tidy tag tag-upload release

# 显示帮助信息
help:
	@echo "BeaconSsoSdk - 可用命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make connect-install - 安装 Connect-Go 代码生成器"
	@echo "  make proto           - 生成指定 proto 的客户端代码"
	@echo "                        示例: make proto PROTO_FILE=proto/beacon/sso/v1/public.proto"
	@echo "  make proto-all       - 生成所有 proto 的客户端代码"
	@echo "  make proto-init      - 初始化 proto 符号链接"
	@echo "  make tidy            - 整理 go.mod 依赖"
	@echo ""
	@echo "发布命令:"
	@echo "  make tag             - 创建带有时间戳的 tag（不推送）"
	@echo "                        格式: v{version}-{YYYYMMDDHHMM}"
	@echo "  make tag-upload      - 单独上传 tag"
	@echo "  make release         - 创建 tag 并推送到远程仓库"
	@echo ""

# 安装 Connect-Go 代码生成器
connect-install:
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "✅ Connect-Go 和 Protobuf 代码生成器安装完成"

# 初始化 proto 符号链接
proto-init:
	@mkdir -p $(dir $(XBASE_LINK))
	@if [ ! -d "$(BASE_GO_MODULE_DIR)" ]; then \
		echo "错误: 找不到 bamboo-base-go 模块，请先运行 go mod download"; \
		exit 1; \
	fi
	@ln -sf $(BASE_GO_MODULE_DIR)/proto/base.proto $(XBASE_LINK)
	@echo "符号链接已创建: $(XBASE_LINK) -> $(BASE_GO_MODULE_DIR)/plugins/grpc/proto/base.proto"

# 生成 proto（自动初始化符号链接）
proto: proto-init
	buf generate --path $(PROTO_FILE)

# 生成所有 proto 文件
proto-all: proto-init
	buf generate --path client/proto/beacon/sso/v1/public.proto
	buf generate --path client/proto/beacon/sso/v1/auth.proto
	@echo "✅ 所有 proto 文件生成完成"

tidy:
	go mod tidy

# 创建 tag（仅本地）
tag:
	@echo "创建 tag: $(TAG_NAME)"
	git tag -a $(TAG_NAME) -m "Release $(TAG_NAME)"
	@echo "✅ Tag $(TAG_NAME) 创建成功"

tag-upload:
	@echo "推送 tag 到远程仓库..."
	git push --tags
	@echo "✅ Tag $(TAG_NAME) 推送成功！"

# 创建 tag 并推送
release: tag tag-upload
