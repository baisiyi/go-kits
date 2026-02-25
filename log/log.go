/*
log包的全局入口
*/

package log

import (
	"fmt"
	"sync"
)

var (
	mu            sync.RWMutex
	defaultLogger Logger
)

func init() {
	// 默认使用控制台输出
	// 注意: 这里不调用 Init()，让用户自行初始化
	// Init() 会在首次使用时自动调用
}

// ensureInit 确保日志已初始化
func ensureInit() {
	mu.RLock()
	initialized := defaultLogger != nil
	mu.RUnlock()
	if !initialized {
		mu.Lock()
		// 双重检查
		if defaultLogger == nil {
			defaultLogger = NewZapLog(defaultConfig)
		}
		mu.Unlock()
	}
}

// Init 初始化日志系统，使用默认配置（控制台输出info级别）
func Init(opts ...Option) {
	cfg := defaultConfig
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	SetDefault(NewZapLog(cfg))
}

// SetDefault 设置默认logger
func SetDefault(logger Logger) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = logger
}

// GetDefaultLogger 获取默认logger
func GetDefaultLogger() Logger {
	ensureInit()
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	return l
}

// Infof 格式化 info 日志
func Infof(format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

// Errorf 格式化 error 日志
func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

// Warnf 格式化 warn 日志
func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

// Debugf 格式化 debug 日志
func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}

// Info 结构化 info 日志
func Info(msg string, fields ...Field) {
	GetDefaultLogger().Info(msg, fields...)
}

// Error 结构化 error 日志
func Error(msg string, fields ...Field) {
	GetDefaultLogger().Error(msg, fields...)
}

// Warn 结构化 warn 日志
func Warn(msg string, fields ...Field) {
	GetDefaultLogger().Warn(msg, fields...)
}

// Debug 结构化 debug 日志
func Debug(msg string, fields ...Field) {
	GetDefaultLogger().Debug(msg, fields...)
}

// Fatal 结构化 fatal 日志
func Fatal(msg string, fields ...Field) {
	GetDefaultLogger().Fatal(msg, fields...)
}

// Panic 结构化 panic 日志
func Panic(msg string, fields ...Field) {
	GetDefaultLogger().Panic(msg, fields...)
}

// With 创建带有上下文的logger
func With(fields ...Field) Logger {
	return GetDefaultLogger().With(fields...)
}

// Named 创建带名称的子logger
func Named(name string) Logger {
	return GetDefaultLogger().Named(name)
}

// Sync 同步日志缓冲
func Sync() error {
	return GetDefaultLogger().Sync()
}

// Fatalf 格式化 fatal 日志
func Fatalf(format string, args ...interface{}) {
	GetDefaultLogger().Fatal(fmt.Sprintf(format, args...))
}

// Panicf 格式化 panic 日志
func Panicf(format string, args ...interface{}) {
	GetDefaultLogger().Panic(fmt.Sprintf(format, args...))
}
