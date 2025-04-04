package base

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

// Handler 提供公共功能
type Handler struct {
	Client kubernetes.Client
	Log    logger.Logger
	Scope  interfaces.ResourceScope
	Group  interfaces.APIGroup
}

// NewHandler 创建新的基础处理程序
func NewHandler(client kubernetes.Client, scope interfaces.ResourceScope, group interfaces.APIGroup) Handler {
	return Handler{
		Client: client,
		Log:    logger.GetLogger(),
		Scope:  scope,
		Group:  group,
	}
}

// GetScope 实现ToolHandler接口
func (h *Handler) GetScope() interfaces.ResourceScope {
	return h.Scope
}

// GetAPIGroup 实现ToolHandler接口
func (h *Handler) GetAPIGroup() interfaces.APIGroup {
	return h.Group
}
