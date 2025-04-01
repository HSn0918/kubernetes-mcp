package models

import (
	"fmt"
	"time"
)

// PodLogsResponse 定义Pod日志响应结构
type PodLogsResponse struct {
	Pod          string    `json:"pod"`
	Namespace    string    `json:"namespace"`
	Container    string    `json:"container,omitempty"`
	Previous     bool      `json:"previous"`
	Timestamps   bool      `json:"timestamps"`
	TailLines    int       `json:"tailLines"`
	LineCount    int       `json:"lineCount"`
	TotalLines   int       `json:"totalLines"`
	Truncated    bool      `json:"truncated,omitempty"`
	LogSize      uint64    `json:"logSize"`
	LogSizeHuman string    `json:"logSizeHuman"`
	Logs         string    `json:"logs"`
	RetrievedAt  time.Time `json:"retrievedAt"`
}

// LogAnalysisResponse 定义日志分析响应结构
type LogAnalysisResponse struct {
	Pod           string      `json:"pod"`
	Namespace     string      `json:"namespace"`
	Container     string      `json:"container,omitempty"`
	LinesAnalyzed int         `json:"linesAnalyzed"`
	Previous      bool        `json:"previous"`
	ErrorCount    int         `json:"errorCount"`
	WarningCount  int         `json:"warningCount"`
	ErrorPattern  string      `json:"errorPattern,omitempty"`
	Prompt        string      `json:"prompt,omitempty"`
	Analysis      LogAnalysis `json:"analysis"`
	RetrievedAt   time.Time   `json:"retrievedAt"`
}

// LogAnalysis 定义日志分析结果结构
type LogAnalysis struct {
	Summary          string     `json:"summary"`
	Errors           []LogEvent `json:"errors,omitempty"`
	Warnings         []LogEvent `json:"warnings,omitempty"`
	TimeDistribution TimeStats  `json:"timeDistribution"`
	KeyInsights      []string   `json:"keyInsights,omitempty"`
	Recommendations  []string   `json:"recommendations,omitempty"`
}

// LogEvent 定义日志事件结构
type LogEvent struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Message   string    `json:"message"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"firstSeen,omitempty"`
	LastSeen  time.Time `json:"lastSeen,omitempty"`
}

// TimeStats 定义时间统计结构
type TimeStats struct {
	StartTime      time.Time `json:"startTime,omitempty"`
	EndTime        time.Time `json:"endTime,omitempty"`
	Duration       string    `json:"duration"`
	PeakTimeSpan   string    `json:"peakTimeSpan,omitempty"`
	PeakEventCount int       `json:"peakEventCount,omitempty"`
}

// NewLogAnalysisResponseFromResult 将LogAnalysisResult转换为LogAnalysisResponse
func NewLogAnalysisResponseFromResult(
	podName, namespace, container string,
	result *LogAnalysisResult,
	lineCount int,
	previous bool,
	errorPattern string,
	prompt string,
) *LogAnalysisResponse {

	// 创建LogEvent数组
	errors := make([]LogEvent, 0)
	for errorMsg, count := range result.TopErrors {
		errors = append(errors, LogEvent{
			Message: errorMsg,
			Count:   count,
		})
	}

	// 创建警告数组 (简单模拟，因为LogAnalysisResult没有直接存储warnings)
	warnings := make([]LogEvent, 0)

	// 创建时间统计
	timeStats := TimeStats{
		Duration: result.ProcessingDuration.String(),
	}
	if !result.TimeRange[0].IsZero() {
		timeStats.StartTime = result.TimeRange[0]
	}
	if !result.TimeRange[1].IsZero() {
		timeStats.EndTime = result.TimeRange[1]
	}

	// 创建分析结果
	summary := fmt.Sprintf("分析了%d行日志，发现%d个错误，%d个警告",
		lineCount, result.ErrorCount, result.WarningCount)

	// 提取关键洞察
	insights := []string{
		fmt.Sprintf("日志时间跨度：%s", result.TimeRange[1].Sub(result.TimeRange[0]).String()),
	}

	// 根据错误数生成建议
	recommendations := []string{}
	if result.ErrorCount > 0 {
		recommendations = append(recommendations, "查看并解决出现频率最高的错误")
	}

	analysis := LogAnalysis{
		Summary:          summary,
		Errors:           errors,
		Warnings:         warnings,
		TimeDistribution: timeStats,
		KeyInsights:      insights,
		Recommendations:  recommendations,
	}

	// 创建完整响应
	response := &LogAnalysisResponse{
		Pod:           podName,
		Namespace:     namespace,
		Container:     container,
		LinesAnalyzed: lineCount,
		Previous:      previous,
		ErrorCount:    result.ErrorCount,
		WarningCount:  result.WarningCount,
		ErrorPattern:  errorPattern,
		Prompt:        prompt,
		Analysis:      analysis,
		RetrievedAt:   time.Now(),
	}

	return response
}
