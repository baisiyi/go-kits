# log - 日志库

基于 Uber Zap 的高性能结构化日志库，简单易用，优雅整洁。

## 特性

- 基于 Zap 的高性能日志
- 控制台和文件输出支持
- 日志轮转（按时间、大小、数量）
- JSON 和 Console 两种格式
- 支持结构化日志
- 灵活的 Options 配置模式

## 快速开始

### 简单用法

默认初始化（控制台输出，info 级别）：

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
)

func main() {
    log.Infof("Hello, %s!", "world")
    log.Debugf("Debug: %d", 42)
    log.Warnf("Warning: %s", "be careful")
    log.Errorf("Error: %v", "something failed")
}
```

### 自定义配置

使用 Options 模式配置日志：

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
)

func main() {
    // 初始化日志
    log.Init(
        log.WithLevel("debug"),
        log.WithFile("./logs/app.log"),
        log.WithMaxSize(100),    // 单文件最大 100MB
        log.WithMaxAge(7),       // 保留 7 天
        log.WithJSONFormatter(), // JSON 格式
    )

    // 使用
    log.Infof("server started on port %d", 8080)
}
```

## 结构化日志

支持结构化日志，告别拼接字符串：

```go
// 基础结构化日志
log.Info("user login",
    log.String("user_id", "12345"),
    log.String("ip", "192.168.1.100"),
)

// 多种类型支持
log.Info("request",
    log.String("path", "/api/users"),
    log.Int("status", 200),
    log.Int64("duration_ms", 150),
    log.Float64("price", 99.99),
    log.Bool("success", true),
    log.Any("metadata", map[string]interface{}{"key": "value"}),
)

// 子 Logger（带名称）
dbLog := log.Named("db")
dbLog.Error("connection failed", log.String("error", "timeout"))

// 上下文
ctx := log.With(log.String("trace_id", "abc123"))
ctx.Info("request handled")
```

## Options 配置

| Option | 说明 | 默认值 |
|--------|------|--------|
| `WithLevel(level)` | 日志级别 | "info" |
| `WithFile(filename)` | 文件输出 | 控制台 |
| `WithMaxSize(size)` | 单文件最大大小(MB) | - |
| `WithMaxAge(days)` | 文件保留天数 | - |
| `WithMaxBackups(count)` | 最大备份数 | - |
| `WithRotationTime(minutes)` | 轮转间隔(分钟) | - |
| `WithJSONFormatter()` | JSON 格式 | console |
| `WithConsoleFormatter()` | 控制台格式 | - |
| `WithColor()` | 彩色输出 | - |

### 完整示例

```go
log.Init(
    log.WithLevel("debug"),
    log.WithFile("./logs/app.log"),
    log.WithMaxSize(100),
    log.WithMaxAge(7),
    log.WithMaxBackups(10),
    log.WithRotationTime(60),
    log.WithJSONFormatter(),
)
```

## Logger 接口

可自定义 Logger 实现：

```go
type Logger interface {
    // 结构化日志
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
```

### 自定义 Logger 示例

```go
type MyLogger struct{}

func (l *MyLogger) Debug(msg string, fields ...log.Field) { /* ... */ }
func (l *MyLogger) Info(msg string, fields ...log.Field) { /* ... */ }
func (l *MyLogger) Warn(msg string, fields ...log.Field) { /* ... */ }
func (l *MyLogger) Error(msg string, fields ...log.Field) { /* ... */ }
func (l *MyLogger) Fatal(msg string, fields ...log.Field) { /* ... */ }
func (l *MyLogger) Panic(msg string, fields ...log.Field) { /* ... */ }

func (l *MyLogger) Debugf(format string, args ...interface{}) { /* ... */ }
func (l *MyLogger) Infof(format string, args ...interface{}) { /* ... */ }
func (l *MyLogger) Warnf(format string, args ...interface{}) { /* ... */ }
func (l *MyLogger) Errorf(format string, args ...interface{}) { /* ... */ }

func (l *MyLogger) With(fields ...log.Field) log.Logger { return l }
func (l *MyLogger) Named(name string) log.Logger { return l }
func (l *MyLogger) Sync() error { return nil }

// 使用自定义 Logger
log.SetDefault(&MyLogger{})
```

## 全局 API

```go
// 初始化
func Init(opts ...Option)

// 设置默认 Logger
func SetDefault(logger Logger)

// 获取默认 Logger
func GetDefaultLogger() Logger

// 格式化日志
func Infof(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Debugf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
func Panicf(format string, args ...interface{})

// 结构化日志
func Info(msg string, fields ...Field)
func Error(msg string, fields ...Field)
func Warn(msg string, fields ...Field)
func Debug(msg string, fields ...Field)
func Fatal(msg string, fields ...Field)
func Panic(msg string, fields ...Field)

// 上下文
func With(fields ...Field) Logger
func Named(name string) Logger

// 同步
func Sync() error
```

## Field 构造函数

```go
log.String(key, value)
log.Int(key, value)
log.Int64(key, value)
log.Uint(key, value)
log.Uint64(key, value)
log.Float64(key, value)
log.Bool(key, value)
log.Duration(key, value)
log.Time(key, value)
log.ByteString(key, value)
log.Any(key, value)
```

## 日志级别

| 级别 | 说明 |
|------|------|
| debug | 调试信息 |
| info | 普通信息（默认） |
| warn | 警告信息 |
| error | 错误信息 |
| fatal | 致命错误（会退出进程） |

## 配置结构（高级用法）

对于复杂配置，可以直接使用 Config：

```go
cfg := log.Config{{
    Writer:    log.OutputFile,
    Formatter: log.FormatterJson,
    Level:     "debug",
    WriteConfig: log.WriteConfig{
        Filename:    "./logs/app.log",
        MaxSize:     100,
        MaxAge:      7,
        MaxBackups:  10,
        RotationTime: 60,
    },
}}

logger := log.NewZapLog(cfg)
log.SetDefault(logger)
```
