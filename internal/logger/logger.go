package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"mq-toolkit/pkg/types"
	"os"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// Logger 日志记录器
type Logger struct {
	level      Level
	output     io.Writer
	mu         sync.RWMutex
	entries    []types.LogEntry
	maxEntries int
	listeners  []LogListener
}

// LogListener 日志监听器
type LogListener func(entry types.LogEntry)

// New 创建新的日志记录器
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		entries:    make([]types.LogEntry, 0),
		maxEntries: 1000, // 默认保留最近1000条日志
		listeners:  make([]LogListener, 0),
	}
}

// NewDefault 创建默认日志记录器
func NewDefault() *Logger {
	return New(LevelInfo, os.Stdout)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// SetMaxEntries 设置最大日志条目数
func (l *Logger) SetMaxEntries(max int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxEntries = max

	// 如果当前条目数超过最大值，删除旧的条目
	if len(l.entries) > max {
		l.entries = l.entries[len(l.entries)-max:]
	}
}

// AddListener 添加日志监听器
func (l *Logger) AddListener(listener LogListener) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.listeners = append(l.listeners, listener)
}

// log 记录日志
func (l *Logger) log(level Level, source, message string, extra map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	entry := types.LogEntry{
		Level:     levelNames[level],
		Message:   message,
		Timestamp: time.Now(),
		Source:    source,
		Extra:     extra,
	}

	// 添加到内存中的日志条目
	l.entries = append(l.entries, entry)

	// 如果超过最大条目数，删除最旧的条目
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[1:]
	}

	// 通知监听器
	for _, listener := range l.listeners {
		go listener(entry)
	}

	// 输出到指定目标
	if l.output != nil {
		var output string
		data, err := json.Marshal(entry)
		if err != nil {
			output = fmt.Sprintf("[%s] %s: %s (JSON marshal error: %v)\n",
				entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Level, entry.Message, err)
		} else {
			output = string(data) + "\n"
		}

		l.output.Write([]byte(output))
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(source, message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(LevelDebug, source, message, extraData)
}

// Info 记录信息日志
func (l *Logger) Info(source, message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(LevelInfo, source, message, extraData)
}

// Warn 记录警告日志
func (l *Logger) Warn(source, message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(LevelWarn, source, message, extraData)
}

// Error 记录错误日志
func (l *Logger) Error(source, message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(LevelError, source, message, extraData)
}

// Fatal 记录致命错误日志并退出程序
func (l *Logger) Fatal(source, message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(LevelFatal, source, message, extraData)
	os.Exit(1)
}

// GetEntries 获取日志条目
func (l *Logger) GetEntries() []types.LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 返回副本以避免并发问题
	entries := make([]types.LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// GetEntriesByLevel 按级别获取日志条目
func (l *Logger) GetEntriesByLevel(level string) []types.LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []types.LogEntry
	for _, entry := range l.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetEntriesBySource 按来源获取日志条目
func (l *Logger) GetEntriesBySource(source string) []types.LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []types.LogEntry
	for _, entry := range l.entries {
		if entry.Source == source {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Clear 清空日志条目
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]types.LogEntry, 0)
}

// 全局日志记录器
var globalLogger *Logger

// init 初始化全局日志记录器
func init() {
	globalLogger = NewDefault()
}

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger 获取全局日志记录器
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewDefault()
	}
	return globalLogger
}

// 全局日志函数
func Debug(source, message string, extra ...map[string]interface{}) {
	globalLogger.Debug(source, message, extra...)
}

func Info(source, message string, extra ...map[string]interface{}) {
	globalLogger.Info(source, message, extra...)
}

func Warn(source, message string, extra ...map[string]interface{}) {
	globalLogger.Warn(source, message, extra...)
}

func Error(source, message string, extra ...map[string]interface{}) {
	globalLogger.Error(source, message, extra...)
}

func Fatal(source, message string, extra ...map[string]interface{}) {
	globalLogger.Fatal(source, message, extra...)
}
