package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	return uuid.New().String()
}

// GenerateShortID 生成短ID（8位）
func GenerateShortID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.0fns", float64(d.Nanoseconds()))
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// SanitizeString 清理字符串（移除控制字符）
func SanitizeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 { // 可打印字符
			result.WriteRune(r)
		}
	}
	return result.String()
}

// IsValidTopic 验证主题名称是否有效
func IsValidTopic(topic string) bool {
	if topic == "" {
		return false
	}
	// 基本验证：不能包含空格和特殊字符
	for _, r := range topic {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return false
		}
	}
	return true
}

// ParseConnectionString 解析连接字符串
func ParseConnectionString(connStr string) (host string, port int, err error) {
	parts := strings.Split(connStr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid connection string format")
	}

	host = strings.TrimSpace(parts[0])
	if host == "" {
		return "", 0, fmt.Errorf("host cannot be empty")
	}

	var portInt int
	if _, err := fmt.Sscanf(parts[1], "%d", &portInt); err != nil {
		return "", 0, fmt.Errorf("invalid port number: %s", parts[1])
	}

	if portInt <= 0 || portInt > 65535 {
		return "", 0, fmt.Errorf("port number out of range: %d", portInt)
	}

	return host, portInt, nil
}

// MaskPassword 掩码密码
func MaskPassword(password string) string {
	if password == "" {
		return ""
	}
	if len(password) <= 2 {
		return strings.Repeat("*", len(password))
	}
	return password[:1] + strings.Repeat("*", len(password)-2) + password[len(password)-1:]
}

// Contains 检查切片是否包含元素
func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Remove 从切片中移除元素
func Remove[T comparable](slice []T, item T) []T {
	var result []T
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// Unique 去重切片
func Unique[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	var result []T
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// MaxInt 返回两个整数中的最大值
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt 返回两个整数中的最小值
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两个int64中的最大值
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个int64中的最小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
