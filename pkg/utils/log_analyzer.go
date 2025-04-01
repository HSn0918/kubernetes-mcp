package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hsn0918/kubernetes-mcp/pkg/models"
)

// logAnalyzer 日志分析器结构体
type logAnalyzer struct {
	pattern models.LogPattern
}

// NewLogAnalyzer 创建一个新的日志分析器
func NewLogAnalyzer() *logAnalyzer {
	return &logAnalyzer{
		pattern: DefaultLogPattern(),
	}
}

// NewLogAnalyzerWithPattern 创建一个使用自定义错误模式的日志分析器
func NewLogAnalyzerWithPattern(customErrorPattern string) *logAnalyzer {
	pattern := DefaultLogPattern()
	if customErrorPattern != "" {
		pattern.ErrorPattern = customErrorPattern
	}
	return &logAnalyzer{
		pattern: pattern,
	}
}

// SetCustomErrorPattern 设置自定义错误模式
func SetCustomErrorPattern(pattern string) models.LogPattern {
	defaultPattern := DefaultLogPattern()
	if pattern != "" {
		defaultPattern.ErrorPattern = pattern
	}
	return defaultPattern
}

// AnalyzeLogs 分析日志行并返回结果
func (a *logAnalyzer) AnalyzeLogs(logLines []string) *models.LogAnalysisResult {
	startTime := time.Now()
	result := models.NewLogAnalysisResult()

	// 编译正则
	errorRegex := regexp.MustCompile(a.pattern.ErrorPattern)
	warningRegex := regexp.MustCompile(a.pattern.WarningPattern)
	infoRegex := regexp.MustCompile(a.pattern.InfoPattern)
	levelRegex := regexp.MustCompile(a.pattern.LevelPattern)
	responseTimeRegex := regexp.MustCompile(a.pattern.ResponseTimePattern)
	statusCodeRegex := regexp.MustCompile(a.pattern.StatusCodePattern)
	userAgentRegex := regexp.MustCompile(a.pattern.UserAgentPattern)
	resourceRegex := regexp.MustCompile(a.pattern.ResourcePattern)
	timestampRegex := regexp.MustCompile(a.pattern.TimestampPattern)

	var firstTimestamp, lastTimestamp time.Time
	hasTimestamp := false

	// 统计每小时错误数量
	hourlyErrors := make(map[string]int)

	// 响应时间分类
	timeCategories := DefaultTimeCategories()

	for i, line := range logLines {
		// 提取时间戳
		timestampMatch := timestampRegex.FindString(line)
		if timestampMatch != "" {
			parsedTime, err := time.Parse(time.RFC3339, timestampMatch)
			if err == nil {
				if !hasTimestamp {
					firstTimestamp = parsedTime
					lastTimestamp = parsedTime
					hasTimestamp = true
				} else {
					if parsedTime.Before(firstTimestamp) {
						firstTimestamp = parsedTime
					}
					if parsedTime.After(lastTimestamp) {
						lastTimestamp = parsedTime
					}
				}

				// 统计小时分布
				hourKey := parsedTime.Format("2006-01-02 15")
				if errorRegex.MatchString(line) {
					hourlyErrors[hourKey]++
				}
			}
		}

		// 检测错误
		if errorRegex.MatchString(line) {
			result.ErrorCount++

			// 提取错误消息（尝试找到错误的主要部分）
			errorMsg := ExtractErrorMessage(line)
			result.TopErrors[errorMsg]++

			// 查找周围上下文
			contextStart := Max(0, i-2)
			contextEnd := Min(len(logLines), i+3)
			context := strings.Join(logLines[contextStart:contextEnd], "\n")

			// 只保存前10个不同类型的错误上下文
			if len(result.TopPatterns) < 10 {
				contextHash := GetContextHash(context)
				if _, exists := result.TopPatterns[contextHash]; !exists {
					result.TopPatterns[contextHash] = 1
				} else {
					result.TopPatterns[contextHash]++
				}
			}
		}

		// 检测警告
		if warningRegex.MatchString(line) {
			result.WarningCount++
		}

		// 检测信息日志
		if infoRegex.MatchString(line) {
			result.InfoCount++
		}

		// 识别日志级别
		levelMatches := levelRegex.FindStringSubmatch(line)
		if len(levelMatches) > 1 {
			level := strings.ToLower(levelMatches[1])
			result.LogLevels[level]++
		}

		// 提取响应时间
		responseTimeMatches := responseTimeRegex.FindStringSubmatch(line)
		if len(responseTimeMatches) > 2 {
			if responseTime, err := strconv.Atoi(responseTimeMatches[2]); err == nil {
				result.ResponseTimes = append(result.ResponseTimes, responseTime)

				// 对响应时间进行分类
				for _, category := range timeCategories {
					if category.Threshold < 0 || responseTime < category.Threshold {
						result.ResponseTimeStats[category.Name]++
						break
					}
				}
			}
		}

		// 提取HTTP状态码
		statusCodeMatches := statusCodeRegex.FindStringSubmatch(line)
		if len(statusCodeMatches) > 2 {
			if statusCode, err := strconv.Atoi(statusCodeMatches[2]); err == nil {
				result.StatusCodes[statusCode]++
			}
		}

		// 提取用户代理
		userAgentMatches := userAgentRegex.FindStringSubmatch(line)
		if len(userAgentMatches) > 1 {
			userAgent := strings.TrimSpace(userAgentMatches[1])
			if len(userAgent) > 10 { // 只保留有效的用户代理
				result.UserAgents[userAgent]++
			}
		}

		// 提取资源使用情况
		resourceMatches := resourceRegex.FindAllStringSubmatch(line, -1)
		for _, match := range resourceMatches {
			if len(match) > 2 {
				resourceType := strings.ToLower(match[1])
				if resourceValue, err := strconv.ParseFloat(match[2], 64); err == nil {
					resourceValueInt := int(resourceValue)
					if resourceValueInt > 0 {
						result.ResourceUsage[resourceType] = append(result.ResourceUsage[resourceType], resourceValueInt)
					}
				}
			}
		}
	}

	// 设置时间范围
	if hasTimestamp {
		result.TimeRange[0] = firstTimestamp
		result.TimeRange[1] = lastTimestamp
	}

	// 保存时间分布
	for hour, count := range hourlyErrors {
		result.TimeBased[hour] = count
	}

	result.ProcessingDuration = time.Since(startTime)
	return result
}

// AnalyzeLogsWithPrompt 根据提供的提示分析日志行并返回结果
func (a *logAnalyzer) AnalyzeLogsWithPrompt(logLines []string, prompt string) *models.LogAnalysisResult {
	result := a.AnalyzeLogs(logLines)

	// 设置分析提示
	if prompt != "" {
		result.AnalysisPrompt = prompt
	}

	return result
}
