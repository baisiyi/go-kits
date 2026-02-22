# plugin - 插件系统

go-kits 的插件系统提供了一种简单的方式来注册和获取插件工厂，支持按类型和名称进行组织，并支持插件依赖管理、生命周期回调等功能。

## 核心接口

### Factory

插件工厂接口，用于创建和管理插件实例。

```go
type Factory interface {
    // Type 返回插件类型
    Type() string
    // Setup 加载插件配置
    Setup(name string, dec Decoder) error
}
```

### Decoder

配置解码器接口，用于解析插件配置。

```go
type Decoder interface {
    Decode(v any) error
}
```

### Depender

强依赖接口。如果插件 A 强依赖插件 B，则 B 必须存在且 A 会在 B 初始化之后才初始化。

```go
type Depender interface {
    DependsOn() []string
}
```

### FlexDepender

弱依赖接口。如果插件 A 弱依赖插件 B，当 B 存在时 A 会在 B 初始化之后才初始化。

```go
type FlexDepender interface {
    FlexDependsOn() []string
}
```

### Closer

插件关闭接口，用于释放资源。

```go
type Closer interface {
    Close() error
}
```

### FinishNotifier

插件初始化完成通知接口。当所有插件加载完成后会调用此接口。

```go
type FinishNotifier interface {
    OnFinish(name string) error
}
```

## API

### Register

注册一个插件工厂。

```go
func Register(name string, f Factory)
```

**示例：**

```go
type MyPluginFactory struct{}

func (f *MyPluginFactory) Type() string {
    return "myplugin"
}

func (f *MyPluginFactory) Setup(name string, dec plugin.Decoder) error {
    var cfg map[string]interface{}
    return dec.Decode(&cfg)
}

plugin.Register("my_plugin", &MyPluginFactory{})
```

### Get

根据类型和名称获取插件工厂。

```go
func Get(typ string, name string) Factory
```

**示例：**

```go
factory := plugin.Get("myplugin", "my_plugin")
if factory != nil {
    // 使用 factory
}
```

### Config

插件配置类型，结构为 `map[string]map[string]yaml.Node`，支持从 YAML 文件加载插件配置。

```go
type Config map[string]map[string]yaml.Node
```

### SetupClosables

加载并初始化所有插件，返回一个关闭函数（按初始化逆序关闭插件）。

```go
func (c Config) SetupClosables() (close func() error, err error)
```

### YamlNodeDecoder

YAML 节点解码器，用于解析 YAML 配置文件。

```go
type YamlNodeDecoder struct {
    Node *yaml.Node
}

func (d *YamlNodeDecoder) Decode(cfg interface{}) error
```

## 使用示例

### 完整的插件示例

```go
package main

import (
    "fmt"
    "github.com/baisiyi/go-kits/plugin"
)

// PluginConfig 插件配置结构
type PluginConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

// DatabasePlugin 数据库插件
type DatabasePlugin struct{}

func (p *DatabasePlugin) Type() string {
    return "database"
}

func (p *DatabasePlugin) Setup(name string, dec plugin.Decoder) error {
    var cfg PluginConfig
    if err := dec.Decode(&cfg); err != nil {
        return err
    }
    fmt.Printf("Database plugin '%s' connected to %s:%d\n", name, cfg.Host, cfg.Port)
    return nil
}

func main() {
    // 注册插件
    plugin.Register("mysql", &DatabasePlugin{})

    // 模拟配置（通常从配置文件读取）
    var cfg PluginConfig = PluginConfig{
        Host: "localhost",
        Port: 3306,
    }

    // 获取并初始化插件
    factory := plugin.Get("database", "mysql")
    if factory != nil {
        factory.Setup("mysql", &configDecoder{cfg: cfg})
    }
}

// configDecoder 实现 plugin.Decoder
type configDecoder struct {
    cfg interface{}
}

func (d *configDecoder) Decode(v interface{}) error {
    cfgPtr, ok := v.(*PluginConfig)
    if !ok {
        return fmt.Errorf("invalid config type")
    }
    *cfgPtr = d.cfg.(PluginConfig)
    return nil
}
```

### YAML 配置加载示例

通过 YAML 配置文件加载插件：

```go
package main

import (
    "fmt"

    "gopkg.in/yaml.v3"

    "github.com/baisiyi/go-kits/plugin"
)

// DatabasePlugin 数据库插件
type DatabasePlugin struct{}

func (p *DatabasePlugin) Type() string {
    return "database"
}

func (p *DatabasePlugin) Setup(name string, dec plugin.Decoder) error {
    var cfg struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    }
    if err := dec.Decode(&cfg); err != nil {
        return err
    }
    fmt.Printf("Database plugin '%s' connected to %s:%d\n", name, cfg.Host, cfg.Port)
    return nil
}

func main() {
    // 注册插件
    plugin.Register("mysql", &DatabasePlugin{})

    // YAML 配置
    yamlConfig := `
database:
  mysql:
    host: localhost
    port: 3306
`

    // 解析 YAML 配置
    var config plugin.Config
    if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
        panic(err)
    }

    // 加载所有插件
    closeFunc, err := config.SetupClosables()
    if err != nil {
        panic(err)
    }
    defer closeFunc()
}
```

### 插件依赖示例

```go
// 实现 Depender 接口，声明强依赖
type DependentPlugin struct{}

func (p *DependentPlugin) Type() string {
    return "dependent"
}

func (p *DependentPlugin) Setup(name string, dec plugin.Decoder) error {
    // ...
    return nil
}

func (p *DependentPlugin) DependsOn() []string {
    // 强依赖 database-mysql 插件
    return []string{"database-mysql"}
}
```

## 内部结构

插件系统使用两级 map 存储：

```go
var plugins = make(map[string]map[string]Factory) // type => name => factory
```

这种设计允许：
- 按类型组织插件（如 database, log, cache 等）
- 同类型下可以有多个不同名称的插件实现

## 注意事项

1. **全局状态**：插件注册表是全局的，测试时需要注意隔离
2. **重复注册**：同名插件会被覆盖，请确保唯一性
3. **类型安全**：Get 返回的 Factory 需要根据实际类型进行类型断言
