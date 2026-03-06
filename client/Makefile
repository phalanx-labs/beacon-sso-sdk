# 变量定义
PROTO_FILE ?= proto/beacon/sso/v1/public.proto
BASE_GO_MODULE_DIR := /Users/xiaolfeng/ProgramProjects/Cooperate/bamboo-service/bamboo-base/plugins/grpc
XBASE_LINK := proto/link/base.proto

.PHONY: proto proto-all proto-init

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
	buf generate --path proto/beacon/sso/v1/public.proto
	buf generate --path proto/beacon/sso/v1/auth.proto
	@echo "✅ 所有 proto 文件生成完成"