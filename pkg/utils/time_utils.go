package utils

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
)

// FormatTimeAgo 将时间格式化为人类可读的"多久以前"形式，使用中文
func FormatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		seconds := int(duration.Seconds())
		return fmt.Sprintf("%d秒前", seconds)
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d小时前", hours)
	}

	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d天前", days)
}

// FormatTimeAgoEN 将时间格式化为人类可读的"多久以前"形式，使用英文
func FormatTimeAgoEN(t time.Time) string {
	return humanize.Time(t)
}

// FormatBytes 将字节数格式化为人类可读形式 (KB, MB, GB等)
func FormatBytes(bytes uint64) string {
	return humanize.Bytes(bytes)
}

// FormatNumber 将数字格式化为带千位分隔符的形式
func FormatNumber(num int64) string {
	return humanize.Comma(num)
}

// FormatDuration 将持续时间格式化为人类可读形式
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%d小时", hours)
		}
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		if hours == 0 {
			return fmt.Sprintf("%d天", days)
		}
		return fmt.Sprintf("%d天%d小时", days, hours)
	}
}

// Pluralize 根据数量返回单复数形式的英文单词
func Pluralize(count int, singular, plural string) string {
	return english.Plural(count, singular, plural)
}
