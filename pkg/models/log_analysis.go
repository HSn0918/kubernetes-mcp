package models

import (
	"time"
)

// LogAnalysisResult 存储日志分析结果
type LogAnalysisResult struct {
	ErrorCount         int
	WarningCount       int
	InfoCount          int // 信息日志计数
	TimeRange          [2]time.Time
	TopErrors          map[string]int
	TopPatterns        map[string]int
	ErrorDistribution  map[string]int
	TimeBased          map[string]int
	LogLevels          map[string]int   // 不同日志级别统计
	ResponseTimes      []int            // 响应时间记录（毫秒）
	ResponseTimeStats  map[string]int   // 响应时间统计
	StatusCodes        map[int]int      // HTTP状态码统计
	UserAgents         map[string]int   // 用户代理统计
	ResourceUsage      map[string][]int // 资源使用统计 (CPU/内存)
	ProcessingDuration time.Duration
	AnalysisPrompt     string // 用户提供的分析提示
}

// NewLogAnalysisResult 创建新的日志分析结果实例
func NewLogAnalysisResult() *LogAnalysisResult {
	return &LogAnalysisResult{
		TopErrors:         make(map[string]int),
		TopPatterns:       make(map[string]int),
		ErrorDistribution: make(map[string]int),
		TimeBased:         make(map[string]int),
		LogLevels:         make(map[string]int),
		ResponseTimes:     make([]int, 0),
		ResponseTimeStats: make(map[string]int),
		StatusCodes:       make(map[int]int),
		UserAgents:        make(map[string]int),
		ResourceUsage:     make(map[string][]int),
	}
}

// LogPattern 日志分析使用的正则表达式模式
type LogPattern struct {
	ErrorPattern        string
	WarningPattern      string
	InfoPattern         string
	LevelPattern        string
	ResponseTimePattern string
	StatusCodePattern   string
	UserAgentPattern    string
	ResourcePattern     string
	TimestampPattern    string
}

// TimeCategory 响应时间分类
type TimeCategory struct {
	Name      string
	Threshold int
}

// StatusCodeGroup 表示HTTP状态码分组
type StatusCodeGroup struct {
	Name  string
	Start int
	End   int
}
