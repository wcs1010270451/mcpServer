package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 日志记录器
type Logger struct {
	level LogLevel
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{level: INFO}
	
	// 配置标准log包
	log.SetOutput(os.Stdout)
	log.SetFlags(0) // 不添加时间戳
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// SetLevelFromString 从字符串设置日志级别
func SetLevelFromString(levelStr string) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		SetLevel(DEBUG)
	case "INFO":
		SetLevel(INFO)
	case "WARN":
		SetLevel(WARN)
	case "ERROR":
		SetLevel(ERROR)
	case "FATAL":
		SetLevel(FATAL)
	default:
		SetLevel(INFO)
	}
}

// 内部日志记录方法
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", levelName, message)
}

// 公共方法
func Debug(format string, args ...interface{}) {
	defaultLogger.logf(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.logf(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.logf(WARN, format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.logf(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.logf(FATAL, format, args...)
	os.Exit(1)
}

// 结构化日志方法
func InfoWithFields(message string, fields map[string]interface{}) {
	fieldStr := formatFields(fields)
	Info("%s %s", message, fieldStr)
}

func ErrorWithFields(message string, fields map[string]interface{}) {
	fieldStr := formatFields(fields)
	Error("%s %s", message, fieldStr)
}

func formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	
	var parts []string
	for key, value := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	return fmt.Sprintf("| %s", strings.Join(parts, " "))
}
