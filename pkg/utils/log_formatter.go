package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
)

// FormatLogAnalysis 格式化日志分析结果
func FormatLogAnalysis(podName, namespace, container string, analysis *models.LogAnalysisResult, totalLines int, prompt string) string {
	// JSON格式对AI更友好
	if strings.Contains(strings.ToLower(prompt), "json") {
		return formatLogAnalysisJSON(podName, namespace, container, analysis, totalLines, prompt)
	}

	// 否则使用可读性更高的文本格式
	return formatLogAnalysisText(podName, namespace, container, analysis, totalLines, prompt)
}

// formatLogAnalysisJSON 以JSON格式输出分析结果，方便AI处理
func formatLogAnalysisJSON(podName, namespace, container string, analysis *models.LogAnalysisResult, totalLines int, prompt string) string {
	// 构建结构化数据
	result := map[string]interface{}{
		"metadata": map[string]interface{}{
			"pod":                podName,
			"namespace":          namespace,
			"container":          container,
			"totalLines":         totalLines,
			"prompt":             prompt,
			"analysisPrompt":     analysis.AnalysisPrompt,
			"processingDuration": analysis.ProcessingDuration.String(),
		},
		"summary": map[string]interface{}{
			"logLevels": map[string]interface{}{
				"error":   analysis.ErrorCount,
				"warning": analysis.WarningCount,
				"info":    analysis.InfoCount,
				"other":   analysis.LogLevels,
			},
			"timeRange": []string{
				analysis.TimeRange[0].Format(time.RFC3339),
				analysis.TimeRange[1].Format(time.RFC3339),
			},
		},
		"details": map[string]interface{}{
			"topErrors":   analysis.TopErrors,
			"statusCodes": analysis.StatusCodes,
			"timeBased":   analysis.TimeBased,
		},
	}

	// 添加响应时间统计（如果有）
	if len(analysis.ResponseTimes) > 0 {
		// 计算统计数据
		var sum, min, max int
		min = analysis.ResponseTimes[0]
		max = analysis.ResponseTimes[0]

		for _, t := range analysis.ResponseTimes {
			sum += t
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}

		avg := sum / len(analysis.ResponseTimes)

		result["responseTimes"] = map[string]interface{}{
			"stats": map[string]interface{}{
				"average": avg,
				"min":     min,
				"max":     max,
				"count":   len(analysis.ResponseTimes),
			},
			"distribution": analysis.ResponseTimeStats,
		}
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error generating JSON output: %v", err)
	}

	return string(jsonData)
}

// formatLogAnalysisText 以人类可读文本格式输出分析结果
func formatLogAnalysisText(podName, namespace, container string, analysis *models.LogAnalysisResult, totalLines int, prompt string) string {
	var sb strings.Builder
	separator := "----------------------------------------------------------------------"

	// 1. 打印头部和摘要信息
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(" KUBERNETES POD LOGS ANALYSIS\n")
	sb.WriteString(separator)
	sb.WriteString("\n")

	// 基本信息
	sb.WriteString(fmt.Sprintf("Pod: %s | Namespace: %s", podName, namespace))
	if container != "" {
		sb.WriteString(fmt.Sprintf(" | Container: %s", container))
	}
	sb.WriteString(fmt.Sprintf("\nLines Analyzed: %s\n", humanize.Comma(int64(totalLines))))

	// 显示用户提供的prompt
	if prompt != "" || analysis.AnalysisPrompt != "" {
		promptText := prompt
		if promptText == "" {
			promptText = analysis.AnalysisPrompt
		}
		sb.WriteString(fmt.Sprintf("\n分析焦点: %s\n", promptText))
	}

	// 时间范围（增加更多上下文信息）
	if !analysis.TimeRange[0].IsZero() && !analysis.TimeRange[1].IsZero() {
		duration := analysis.TimeRange[1].Sub(analysis.TimeRange[0])
		durationText := humanize.RelTime(analysis.TimeRange[0], analysis.TimeRange[1], "", "")

		// 增加更多详细的时间信息
		sb.WriteString(fmt.Sprintf("Time Range: %s to %s\n",
			analysis.TimeRange[0].Format("2006-01-02 15:04:05"),
			analysis.TimeRange[1].Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("Duration: %s (%.1f hours / %.1f days)\n",
			durationText,
			duration.Hours(),
			duration.Hours()/24))
	}

	// 摘要统计
	sb.WriteString("\n总结:\n")

	// 日志级别统计（优化可视化）
	totalLogLevels := analysis.ErrorCount + analysis.WarningCount + analysis.InfoCount
	if totalLogLevels > 0 {
		errorPct := Percentage(analysis.ErrorCount, totalLogLevels)
		warnPct := Percentage(analysis.WarningCount, totalLogLevels)
		infoPct := Percentage(analysis.InfoCount, totalLogLevels)

		sb.WriteString(fmt.Sprintf("- 日志级别分布:\n"))
		sb.WriteString(fmt.Sprintf("  - 错误: %s (%.1f%%)\n",
			humanize.Comma(int64(analysis.ErrorCount)), errorPct))
		sb.WriteString(fmt.Sprintf("  - 警告: %s (%.1f%%)\n",
			humanize.Comma(int64(analysis.WarningCount)), warnPct))
		sb.WriteString(fmt.Sprintf("  - 信息: %s (%.1f%%)\n",
			humanize.Comma(int64(analysis.InfoCount)), infoPct))

		// 添加日志级别比例可视化
		sb.WriteString("  - 可视化: [")
		// 错误
		errChars := int(errorPct / 5)
		if analysis.ErrorCount > 0 && errChars == 0 {
			errChars = 1
		}
		sb.WriteString(strings.Repeat("E", errChars))

		// 警告
		warnChars := int(warnPct / 5)
		if analysis.WarningCount > 0 && warnChars == 0 {
			warnChars = 1
		}
		sb.WriteString(strings.Repeat("W", warnChars))

		// 信息
		infoChars := int(infoPct / 5)
		if analysis.InfoCount > 0 && infoChars == 0 {
			infoChars = 1
		}
		sb.WriteString(strings.Repeat("I", infoChars))
		sb.WriteString("]")

		sb.WriteString("\n")
	}

	// 其他日志级别统计
	if len(analysis.LogLevels) > 0 {
		var otherLevels []string
		for level, count := range analysis.LogLevels {
			if level != "error" && level != "warn" && level != "info" {
				otherLevels = append(otherLevels, fmt.Sprintf("%s=%s",
					level, humanize.Comma(int64(count))))
			}
		}
		if len(otherLevels) > 0 {
			sb.WriteString("  - 其他日志级别: " + strings.Join(otherLevels, ", ") + "\n")
		}
	}

	sb.WriteString(fmt.Sprintf("- 分析耗时: %s\n", humanize.RelTime(time.Now().Add(-analysis.ProcessingDuration), time.Now(), "", "")))

	// 响应时间统计
	if len(analysis.ResponseTimes) > 0 {
		// 计算响应时间统计数据
		var sum, min, max int
		min = analysis.ResponseTimes[0]
		max = analysis.ResponseTimes[0]

		for _, t := range analysis.ResponseTimes {
			sum += t
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}

		avg := sum / len(analysis.ResponseTimes)

		sb.WriteString("\n响应时间统计:\n")
		sb.WriteString(fmt.Sprintf("- 平均: %dms | 最小: %dms | 最大: %dms | 样本数: %s\n",
			avg, min, max, humanize.Comma(int64(len(analysis.ResponseTimes)))))

		// 响应时间分布
		if len(analysis.ResponseTimeStats) > 0 {
			sb.WriteString("- 响应时间分布:\n")
			timeCategories := DefaultTimeCategories()
			for _, category := range timeCategories {
				if count, exists := analysis.ResponseTimeStats[category.Name]; exists && count > 0 {
					sb.WriteString(fmt.Sprintf("  - %s: %s (%.1f%%)\n",
						category.Name,
						humanize.Comma(int64(count)),
						Percentage(count, len(analysis.ResponseTimes))))
				}
			}
		}
	}

	// HTTP状态码统计
	if len(analysis.StatusCodes) > 0 {
		sb.WriteString("\nHTTP状态码分布:\n")

		// 计算总状态码数
		totalStatusCodes := 0
		for _, count := range analysis.StatusCodes {
			totalStatusCodes += count
		}

		// 状态码范围分组
		successCount := 0
		clientErrorCount := 0
		serverErrorCount := 0
		redirectCount := 0
		otherCount := 0

		for code, count := range analysis.StatusCodes {
			if code >= 200 && code < 300 {
				successCount += count
			} else if code >= 400 && code < 500 {
				clientErrorCount += count
			} else if code >= 500 && code < 600 {
				serverErrorCount += count
			} else if code >= 300 && code < 400 {
				redirectCount += count
			} else {
				otherCount += count
			}
		}

		// 输出主要状态码组
		if successCount > 0 {
			sb.WriteString(fmt.Sprintf("- 成功 (2xx): %s (%.1f%%)\n",
				humanize.Comma(int64(successCount)),
				Percentage(successCount, totalStatusCodes)))
		}
		if redirectCount > 0 {
			sb.WriteString(fmt.Sprintf("- 重定向 (3xx): %s (%.1f%%)\n",
				humanize.Comma(int64(redirectCount)),
				Percentage(redirectCount, totalStatusCodes)))
		}
		if clientErrorCount > 0 {
			sb.WriteString(fmt.Sprintf("- 客户端错误 (4xx): %s (%.1f%%)\n",
				humanize.Comma(int64(clientErrorCount)),
				Percentage(clientErrorCount, totalStatusCodes)))
		}
		if serverErrorCount > 0 {
			sb.WriteString(fmt.Sprintf("- 服务器错误 (5xx): %s (%.1f%%)\n",
				humanize.Comma(int64(serverErrorCount)),
				Percentage(serverErrorCount, totalStatusCodes)))
		}
		if otherCount > 0 {
			sb.WriteString(fmt.Sprintf("- 其他: %s (%.1f%%)\n",
				humanize.Comma(int64(otherCount)),
				Percentage(otherCount, totalStatusCodes)))
		}

		// 输出具体的错误状态码
		errorCodes := make([]int, 0)
		for code := range analysis.StatusCodes {
			if code >= 400 {
				errorCodes = append(errorCodes, code)
			}
		}

		if len(errorCodes) > 0 {
			sort.Ints(errorCodes)
			sb.WriteString("- 状态码: ")
			for i, code := range errorCodes {
				count := analysis.StatusCodes[code]
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%d(%s)", code, humanize.Comma(int64(count))))
			}
			sb.WriteString("\n")
		}
	}

	// 排序错误以获取最常见的错误
	if len(analysis.TopErrors) > 0 {
		sb.WriteString("\n常见错误类型:\n")

		// 转换为排序切片
		type kv struct {
			Key   string
			Value int
		}
		var sortedErrors []kv
		for k, v := range analysis.TopErrors {
			// 截断过长的错误消息
			errMsg := k
			sortedErrors = append(sortedErrors, kv{errMsg, v})
		}
		sort.Slice(sortedErrors, func(i, j int) bool {
			return sortedErrors[i].Value > sortedErrors[j].Value
		})

		// 显示前10个最常见的错误
		count := Min(10, len(sortedErrors))
		for i := 0; i < count; i++ {
			item := sortedErrors[i]
			sb.WriteString(fmt.Sprintf("- %s: %s次 (%.1f%%)\n",
				item.Key,
				humanize.Comma(int64(item.Value)),
				Percentage(item.Value, analysis.ErrorCount)))
		}
	}

	// 显示时间分布
	if len(analysis.TimeBased) > 0 {
		sb.WriteString("\n错误时间分布: ")

		// 转换为排序切片
		type kv struct {
			Key   string
			Value int
		}
		var sortedHours []kv
		for k, v := range analysis.TimeBased {
			sortedHours = append(sortedHours, kv{k, v})
		}
		sort.Slice(sortedHours, func(i, j int) bool {
			return sortedHours[i].Key < sortedHours[j].Key
		})

		// 计算总错误数
		totalErrors := 0
		for _, item := range sortedHours {
			totalErrors += item.Value
		}

		// 压缩显示时间分布
		for i, item := range sortedHours {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s时(%s)",
				item.Key,
				humanize.Comma(int64(item.Value))))
		}
		sb.WriteString("\n")
	}

	// 资源使用情况
	if len(analysis.ResourceUsage) > 0 {
		sb.WriteString("\n资源使用情况:\n")

		for resource, values := range analysis.ResourceUsage {
			if len(values) > 0 {
				// 计算统计数据
				sum := 0
				min := values[0]
				max := values[0]

				for _, v := range values {
					sum += v
					if v < min {
						min = v
					}
					if v > max {
						max = v
					}
				}

				avg := sum / len(values)

				sb.WriteString(fmt.Sprintf("- %s:\n", resource))
				sb.WriteString(fmt.Sprintf("  - 平均: %d\n", avg))
				sb.WriteString(fmt.Sprintf("  - 最小: %d\n", min))
				sb.WriteString(fmt.Sprintf("  - 最大: %d\n", max))
				sb.WriteString(fmt.Sprintf("  - 样本数: %s\n", humanize.Comma(int64(len(values)))))
			}
		}
	}

	// 3. 打印结束
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n--- 分析结束 ---\n")
	sb.WriteString(separator)
	sb.WriteString("\n")

	return sb.String()
}
