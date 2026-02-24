# 获取版本号（去除 v 前缀）
VERSION := $(shell cat version | sed 's/^v//')

# 获取当前时间戳（格式：YYYYMMDDHHMM）
TIMESTAMP := $(shell date +"%Y%m%d%H%M")

# 完整 tag 名称
TAG_NAME := v$(VERSION)-$(TIMESTAMP)

.DEFAULT_GOAL := help

.PHONY: help proto tidy proto-init connect-install tag tag-upload release

# 显示帮助信息
help:
	@echo "BeaconSsoSDK - 可用命令"
	@echo ""
	@echo "发布命令:"
	@echo "  make tag        	- 创建带有时间戳的 tag（不推送）"
	@echo "                   	  格式: v{version}-{YYYYMMDDHHMM}"
	@echo "                   	  示例: v1.0.0-202602191755"
	@echo "  make tag-upload 	- 单独上传 tag"
	@echo "  make release    	- 创建 tag 并推送到远程仓库"
	@echo ""

# 安装 Connect-Go 代码生成器
connect-install:
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@echo "✅ Connect-Go 代码生成器安装完成"

# 初始化 proto 符号链接
proto-init:
	@mkdir -p $(dir $(XBASE_LINK))
	@ln -sf $(BASE_GO_MODULE_DIR)/proto/base.proto $(XBASE_LINK)
	@echo "符号链接已创建: $(XBASE_LINK) -> $(BASE_GO_MODULE_DIR)/proto/base.proto"

# 生成 proto（自动初始化符号链接）
proto: proto-init
	buf generate --path $(PROTO_FILE)

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
