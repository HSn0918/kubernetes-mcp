package utils

import (
	"bufio"
	"regexp"
	"strings"
)

// ExtractErrorMessage 从日志行中提取错误消息
func ExtractErrorMessage(logLine string) string {
	// 尝试提取: error: message 或 ERROR: message 格式
	errorMsgRegex := regexp.MustCompile(`(?i)(?:error|exception|fail\w*|fatal)[\s:]+(.{10,60})`)
	matches := errorMsgRegex.FindStringSubmatch(logLine)

	if len(matches) > 1 {
		// 清理和限制长度
		msg := strings.TrimSpace(matches[1])
		if len(msg) > 1024 {
			msg = msg[:1021] + "..."
		}
		return msg
	}

	// 如果没有清晰的错误格式，返回截断的一部分日志
	logPart := logLine
	// 移除时间戳部分
	timestampRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z\s*`)
	logPart = timestampRegex.ReplaceAllString(logPart, "")

	if len(logPart) > 1024 {
		logPart = logPart[:1021] + "..."
	}
	return logPart
}

// GetContextHash 为上下文生成一个简单的哈希值
func GetContextHash(context string) string {
	// 简化上下文 - 移除具体时间戳和数字ID等变量部分
	simplified := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z`).ReplaceAllString(context, "TIMESTAMP")
	simplified = regexp.MustCompile(`\b[0-9a-f]{8,}\b`).ReplaceAllString(simplified, "ID")

	// 取前100个字符作为哈希标识
	if len(simplified) > 100 {
		return simplified[:100]
	}
	return simplified
}

// Percentage 计算百分比
func Percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

// FormatPodLogs 格式化Pod日志输出
func FormatPodLogs(summaryDetails string, logs string) string {
	var sb strings.Builder
	separator := "----------------------------------------------------------------------"

	// 1. 打印头部和摘要信息
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(" KUBERNETES POD LOGS\n")
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(summaryDetails)
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n\n")

	// 2. 打印日志内容（带缩进）
	if strings.TrimSpace(logs) == "" {
		sb.WriteString("  --- No logs available with the specified options. ---\n")
	} else {
		scanner := bufio.NewScanner(strings.NewReader(logs))
		for scanner.Scan() {
			sb.WriteString("  ")
			sb.WriteString(scanner.Text())
			sb.WriteString("\n")
		}
	}

	// 3. 打印尾部
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n--- End of Logs ---\n")
	sb.WriteString(separator)
	sb.WriteString("\n")

	return sb.String()
}
