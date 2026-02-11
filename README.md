# go-kits

一个 Go 语言工具库，提供插件系统和日志库等常用组件。

## 特性

### 插件系统 (plugin)
- 简单易用的插件注册与获取机制
- 支持插件类型和名称的多级映射
- 基于工厂模式的插件抽象

### 日志库 (log)
- 基于 Uber Zap 的高性能结构化日志
- 支持控制台和文件输出
- 日志轮转支持（按时间、文件大小、文件数量）
- 支持自定义日志格式（JSON/Console）
- 多种日志级别（debug, info, warn, error, fatal）
- 便捷的全局日志函数

## 安装

```bash
go get github.com/baisiyi/go-kits
```

## 快速开始

### 插件系统

```go
package main

import (
    "fmt"
    "github.com/baisiyi/go-kits/plugin"
)

// 定义插件配置结构
type MyConfig struct {
    Name string `yaml:"name"`
}

// 定义插件工厂
type MyFactory struct{}

func (f *MyFactory) Type() string {
    return "myplugin"
}

func (f *MyFactory) Setup(name string, dec plugin.Decoder) error {
    var cfg MyConfig
    if err := dec.Decode(&cfg); err != nil {
        return err
    }
    fmt.Printf("Plugin %s configured with name: %s\n", name, cfg.Name)
    return nil
}

func main() {
    // 注册插件
    plugin.Register("my_plugin", &MyFactory{})

    // 获取插件
    factory := plugin.Get("myplugin", "my_plugin")
    if factory != nil {
        fmt.Println("Plugin found!")
    }
}
```

### 日志库

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
)

func main() {
    // 使用便捷函数
    log.Info("Hello, %s!", "world")
    log.Error("This is an error: %v", "something failed")
    log.Debug("Debug message: %d", 123)
    log.Warn("Warning: %s", "be careful")
}
```

### 配置日志

```go
package main

import (
    "github.com/baisiyi/go-kits/log"
)

func main() {
    // 通过配置创建日志
    cfg := log.Config{{
        Writer:    log.OutputConsole,
        Formatter: "console",
        Level:     "debug",
    }}
    logger := log.NewZapLog(cfg)
    logger.Infof("Configured logger: %s", "test")
}
```

## 目录结构

```
go-kits/
├── plugin/               # 插件系统
│   ├── plugin.go        # 插件注册与获取
│   └── plugin_test.go   # 单元测试
├── log/                 # 日志库
│   ├── log.go          # 便捷日志函数
│   ├── logger.go       # Logger 接口定义
│   ├── logger_factory.go # Logger 工厂和注册
│   ├── writer_factory.go # 输出 Writer 工厂
│   ├── zaplogger.go     # Zap 日志实现
│   ├── config.go        # 配置结构
│   └── rollwriter/      # 日志轮转
│       └── roll_writer.go
└── README.md
```

## 依赖

- [go.uber.org/zap](https://github.com/uber-go/zap) - 高性能日志库
- [github.com/lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) - 日志轮转
