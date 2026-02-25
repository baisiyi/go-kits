package log

import (
	"go.uber.org/zap"
)

// Logger 日志接口
type Logger interface {
	// 基础日志（结构化）
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Panic(msg string, fields ...Field)

	// 格式化日志
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// 上下文
	With(fields ...Field) Logger
	Named(name string) Logger

	// 同步
	Sync() error
}

// Field 是 zap.Field 的别名，支持结构化日志
type Field = zap.Field

// 常用Field构造函数（直接暴露zap的）
var (
	String     = zap.String
	Int        = zap.Int
	Int64      = zap.Int64
	Uint       = zap.Uint
	Uint64     = zap.Uint64
	Float64    = zap.Float64
	Bool       = zap.Bool
	Duration   = zap.Duration
	Time       = zap.Time
	ByteString = zap.ByteString
	Any        = zap.Any
)
