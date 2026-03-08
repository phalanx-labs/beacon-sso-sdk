package bSdkStartup

import (
	xRegNode "github.com/bamboo-services/bamboo-base-go/major/register/node"
)

// startupNode 启动注册节点定义
type startupNode struct {
	name string
	node xRegNode.RegNodeList
}

// NewStartupConfig 初始化并返回 SSO SDK 所有依赖注入节点列表。
//
// 该函数聚合了以下注册节点，用于在应用启动时批量注册 SSO 相关的上下文键值：
//   - `oAuthConfig`: OAuth2 核心配置（ClientID、Endpoint 等）
//   - `oAuthRedirectURI`: OAuth2 重定向地址
//   - `ssoClient`: SsoClient gRPC 客户端
//
// 参数:
//   - exclude: 要排除的注册节点名称列表（可选），支持: "oAuthConfig", "oAuthRedirectURI", "ssoClient"
//
// 返回值:
//   - 包含所有未被排除的注册节点的切片。
func NewStartupConfig(exclude ...string) []xRegNode.RegNodeList {
	// 构建排除集合
	excludeSet := make(map[string]struct{}, len(exclude))
	for _, name := range exclude {
		excludeSet[name] = struct{}{}
	}

	// 定义所有注册节点
	nodes := []startupNode{
		{name: "oAuthConfig", node: oAuthConfig()},
		{name: "oAuthRedirectURI", node: oAuthRedirectURI()},
		{name: "ssoClient", node: ssoClient()},
	}

	// 过滤并收集注册节点
	regNode := make([]xRegNode.RegNodeList, 0, len(nodes))
	for _, n := range nodes {
		if _, excluded := excludeSet[n.name]; !excluded {
			regNode = append(regNode, n.node)
		}
	}

	return regNode
}
