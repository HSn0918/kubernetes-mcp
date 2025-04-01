package utils

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
)

// DefaultLogPattern 返回默认的日志分析模式
func DefaultLogPattern() models.LogPattern {
	return models.LogPattern{
		ErrorPattern:        `(?i)(error|exception|failed|failure|fatal|panic|crash|timeout|refused|denied|rejected)`,
		WarningPattern:      `(?i)(warn|warning|caution|attention)`,
		InfoPattern:         `(?i)(info|information|notice)`,
		LevelPattern:        `(?i)(debug|info|warn|error|fatal|trace)`,
		ResponseTimePattern: `(?i)(took|elapsed|duration|latency|response time)[\s:=]+(\d+)(?:ms|milliseconds)?`,
		StatusCodePattern:   `(?i)(status|code|http)[:\s=]+(\d{3})`,
		UserAgentPattern:    `(?i)(?:user-agent|browser)[:\s="']+([^"'\s]+\s[^"'\n]{3,40})`,
		ResourcePattern:     `(?i)(cpu|memory|mem)[:\s=]+(\d+\.?\d*)(%|Mi|m|MB)?`,
		TimestampPattern:    `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z`,
	}
}

// DefaultTimeCategories 返回默认的响应时间分类
func DefaultTimeCategories() []models.TimeCategory {
	return []models.TimeCategory{
		{Name: "快速 (<100ms)", Threshold: 100},
		{Name: "正常 (100-500ms)", Threshold: 500},
		{Name: "较慢 (500-1000ms)", Threshold: 1000},
		{Name: "慢 (>1000ms)", Threshold: -1}, // -1表示无上限
	}
}

// DefaultStatusCodeGroups 返回默认的HTTP状态码分组
func DefaultStatusCodeGroups() []models.StatusCodeGroup {
	return []models.StatusCodeGroup{
		{Name: "2xx (成功)", Start: 200, End: 299},
		{Name: "3xx (重定向)", Start: 300, End: 399},
		{Name: "4xx (客户端错误)", Start: 400, End: 499},
		{Name: "5xx (服务器错误)", Start: 500, End: 599},
		{Name: "其他", Start: 0, End: 999}, // 捕获所有其他状态码
	}
}
