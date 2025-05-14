package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *UtilityHandler) GetCurrentTime(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Getting current time")

	// 获取当前时间
	currentTime := time.Now().Format(time.RFC3339)

	// 构建响应
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Current Time: %s", currentTime),
			},
		},
	}, nil
}
