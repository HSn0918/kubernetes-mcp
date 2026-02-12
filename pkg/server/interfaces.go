package server

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer interface {
	Start() error
	Stop() error
	GetServer() *server.MCPServer
}

type Factory interface {
	CreateServer(config *config.Config) (MCPServer, error)
}
