# log - 日志库

基于 Uber Zap 的高性能结构化日志库，提供便捷的日志功能和灵活的配置选项。

## 特性

- 基于 Zap 的高性能日志
- 控制台和文件输出支持
- 日志轮转（按时间、大小、数量）
- JSON 和 Console 两种格式
- 多种日志级别
- 便捷的全局函数
- 支持自定义 Logger 实现

## 快速开始

### 便捷函数

最简单的方式是使用全局便捷函数：

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

### 创建 Logger

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
)

func main() {
    // 创建配置
    cfg := log.Config{{
        Writer:    log.OutputConsole,
        Formatter: "console",
        Level:     "debug",
    }}

    // 创建 logger
    logger := log.NewZapLog(cfg)

    // 使用 logger
    logger.Infof("Hello, %s!", "world")
}
```

## 配置说明

### OutputConfig

主配置结构：

```go
type OutputConfig struct {
    Writer       string          // 输出类型: console, file
    WriteConfig  WriteConfig     // 文件输出配置
    Formatter    string          // 格式: console, json
    FormatConfig FormatConfig    // 格式详细配置
    Level        string          // 日志级别
    CallerSkip   int             // 调用者跳过的层数
    EnableColor  bool            // 是否启用颜色
}
```

### WriteConfig

文件输出配置：

```go
type WriteConfig struct {
    LogPath      string // 日志目录
    Filename     string // 日志文件名
    MaxAge       int    // 日志最大保留天数
    MaxBackups   uint   // 最大备份数量
    MaxSize      int64  // 单文件最大大小 (MB)
    RotationTime int    // 轮转时间间隔 (小时)
}
```

### FormatConfig

日志格式配置：

```go
type FormatConfig struct {
    TimeFmt       string // 时间格式
    TimeKey       string // 时间字段名，默认 "T"
    LevelKey      string // 级别字段名，默认 "L"
    NameKey       string // 名称字段名，默认 "N"
    CallerKey     string // 调用者字段名，默认 "C"
    FunctionKey   string // 函数名字段名，默认 ""
    MessageKey    string // 消息字段名，默认 "M"
    StacktraceKey string // 堆栈字段名，默认 "S"
}
```

## 日志级别

支持的日志级别（按优先级从低到高）：

| 级别 | 说明 |
|------|------|
| trace | 最详细级别，等同于 debug |
| debug | 调试信息 |
| info | 普通信息 |
| warn | 警告信息 |
| error | 错误信息 |
| fatal | 致命错误 |

## YAML 配置示例

```yaml
log:
  - writer: console
    formatter: console
    level: debug
    enable_color: true
    caller_skip: 1
    formatter_config:
      time_fmt: ""
      time_key: "T"
      level_key: "L"
      message_key: "M"

  - writer: file
    formatter: json
    level: info
    writer_config:
      log_path: /var/log/myapp
      filename: app.log
      max_age: 7
      max_backups: 10
      max_size: 100
      rotation_time: 24
```

## 日志轮转 (rollwriter)

日志轮转功能由 `log/rollwriter` 包提供，支持以下选项：

```go
import "github.com/baisiyi/go-kits/log/rollwriter"

// 按时间格式
rollwriter.WithTimeFormat(".%Y%m%d")

// 保留天数
rollwriter.WithMaxAge(7)

// 保留天数（Duration）
rollwriter.WithMaxAgeDuration(168 * time.Hour)

// 轮转间隔（小时）
rollwriter.WithRotationAge(24)

// 轮转间隔（Duration）
rollwriter.WithRotationAgeDuration(24 * time.Hour)

// 单文件大小（字节）
rollwriter.WithRotationSize(100 * 1024 * 1024)

// 最大保留文件数
rollwriter.WithRotationCount(10)
```

**注意：** `MaxAge` 和 `RotationCount` 不能同时设置。

### 使用示例

```go
writer, err := rollwriter.NewRollWriter(
    "/var/log/myapp/app.log",
    rollwriter.WithTimeFormat(".%Y%m%d"),
    rollwriter.WithMaxAge(7),
    rollwriter.WithRotationAge(24),
    rollwriter.WithRotationSize(100*1024*1024),
)
```

## Logger 接口

你可以实现自己的 Logger：

```go
type Logger interface {
    Infof(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Debugf(format string, args ...interface{})
    Warnf(format string, args ...interface{})
}
```

### 注册自定义 Logger

```go
type MyLogger struct{}

func (l *MyLogger) Infof(format string, args ...interface{}) {
    // 实现
}

func (l *MyLogger) Errorf(format string, args ...interface{}) {
    // 实现
}

// 注册
log.Register("my_logger", &MyLogger{})

// 获取使用
logger := log.Get("my_logger")
```

## 框架集成 (Factory)

`log.Factory` 实现了 `plugin.Factory` 接口，可供外部框架统一管理插件生命周期。

### Decoder 配置解码器

`Decoder` 是配置解码工具，用于解耦插件配置与调用方：

```go
type Decoder struct {
    OutputConfig *OutputConfig  // 输出配置
    Core         zapcore.Core   // Zap Core
    ZapLevel     zap.AtomicLevel // 日志级别
}
```

**使用模式：** 框架传入 `*Decoder`，Factory 通过 `Decoder.Decode()` 获取配置：

```go
// ConsoleWriterFactory 的内部实现
cfg := &OutputConfig{}
decoder := &Decoder{OutputConfig: cfg}
err := decoder.Decode(&cfg)  // cfg 指向 OutputConfig
```

### Factory 接口

```go
type Factory struct{}

func (f *Factory) Type() string       // 返回 "log"
func (f *Factory) Setup(name string, dec plugin.Decoder) error
```

### 注册到 plugin 框架

log 包在 `init()` 中已将 `DefaultLogFactory` 注册到 plugin：

```go
func init() {
    // ... 其他初始化
    plugin.Register(defaultLoggerName, DefaultLogFactory)
}
```

### 完整 YAML 配置集成示例

使用 `plugin.Config` 从 YAML 文件加载日志配置：

```go
package main

import (
    "fmt"

    "gopkg.in/yaml.v3"

    "github.com/baisiyi/go-kits/log"
    "github.com/baisiyi/go-kits/plugin"
)

func main() {
    // YAML 配置（通常从文件读取）
    yamlConfig := `
log:
  default:
    - writer: console
      formatter: console
      level: debug
      enable_color: true
      caller_skip: 2
    - writer: file
      formatter: json
      level: info
      writer_config:
        log_path: /var/log/myapp
        filename: app.log
        max_age: 7
        max_backups: 10
        max_size: 100
        rotation_time: 24
`

    // 解析 YAML 配置
    var config plugin.Config
    if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
        panic(err)
    }

    // 加载所有插件（包括 log）
    closeFunc, err := config.SetupClosables()
    if err != nil {
        panic(err)
    }
    defer closeFunc()

    // 使用 log 包
    log.Info("application started")
    log.Infof("server listening on %s:%d", "localhost", 8080)
    log.Debugf("debug info: %v", "some data")
}
```

### 直接使用 plugin 框架初始化

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
    "github.com/baisiyi/go-kits/plugin"
)

func main() {
    // 1. 从配置文件读取配置
    yamlConfig := `
writer: console
formatter: console
level: debug
`
    var cfg log.Config
    yaml.Unmarshal([]byte(yamlConfig), &cfg)

    // 2. 创建 Decoder（框架侧实现）
    decoder := &log.Decoder{OutputConfig: &cfg[0]}

    // 3. 通过 plugin 框架获取 log 工厂
    logFactory := plugin.Get("log", "default")

    // 4. 调用 Setup 初始化日志
    logFactory.Setup("default", decoder)

    // 5. 之后可使用 log.Infof() 等便捷函数
    log.Info("framework integration complete")
}
```

## API 参考

### 便捷函数

```go
func Infof(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Debugf(format string, args ...interface{})
```

### Logger 管理

```go
func Register(name string, logger Logger)
func Get(name string) Logger
func GetDefaultLogger() Logger
```

### Logger 创建

```go
func NewZapLog(c Config) Logger
func NewZapLogWithCallerSkip(cfg Config, callerSkip int) Logger
```

### 自定义编码器

```go
func RegisterFormatEncoder(formatName string, newFormatEncoder NewFormatEncoder)
```
