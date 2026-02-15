# Database

GORM 数据库客户端封装，提供单例模式和自定义日志功能。

## 特性

- **单例模式**: 确保整个应用只有一个数据库连接
- **连接池管理**: 支持配置最大连接数、空闲连接数、连接生命周期
- **自定义日志**: 集成自定义日志包，支持慢查询日志
- **健康检查**: 提供数据库连接健康检查接口
- **优雅关闭**: 支持安全关闭数据库连接

## 安装

```go
import "github.com/baisiyi/go-kits/database"
```

## 快速开始

### 1. 配置数据库连接

```go
cfg := &database.DBConfig{
    DSN: &database.Connect{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Name:     "mydb",
    },
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
    LogLevel:        4, // 1:Silent, 2:Error, 3:Warn, 4:Info
    SlowThreshold:   200 * time.Millisecond,
}
```

### 2. 初始化数据库

```go
// 初始化数据库连接
dbClient, err := database.Init(cfg, logger)
if err != nil {
    log.Fatalf("Failed to init database: %v", err)
}
```

### 3. 获取 GORM 实例

```go
// 在业务代码中获取 GORM 实例
ctx := context.Background()
db := dbClient.GetDB(ctx)

// 执行数据库操作
var users []User
result := db.Where("age > ?", 18).Find(&users)
```

### 4. 健康检查

```go
func checkDatabaseHealth() error {
    ctx := context.Background()
    return dbClient.Health(ctx)
}
```

### 5. 优雅关闭

```go
// 在应用关闭时调用
defer func() {
    if err := dbClient.Close(); err != nil {
        log.Printf("Error closing database: %v", err)
    }
}()
```

## 配置说明

### DBConfig

| 字段 | 类型 | 说明 |
|------|------|------|
| DSN | *Connect | 数据库连接配置 |
| MaxOpenConns | int | 最大打开连接数 |
| MaxIdleConns | int | 最大空闲连接数 |
| ConnMaxLifetime | time.Duration | 连接最大生命周期 |
| ConnMaxIdleTime | time.Duration | 空闲连接最大存活时间 |
| LogLevel | int | 日志级别 (1:Silent, 2:Error, 3:Warn, 4:Info) |
| SlowThreshold | time.Duration | 慢查询阈值 |

### Connect

| 字段 | 类型 | 说明 |
|------|------|------|
| Host | string | 数据库主机地址 |
| Port | int | 数据库端口 |
| Username | string | 用户名 |
| Password | string | 密码 |
| Name | string | 数据库名称 |
| TablePrefix | string | 表前缀 (可选) |

## 日志格式

数据库日志会输出以下信息:

```bash
# 普通 SQL
[DB_SQL] Elapsed: 1.5ms | Rows: 10 | SQL: SELECT * FROM users

# 慢查询
[DB_SLOW] Elapsed: 500ms > 200ms | Rows: 1000 | SQL: SELECT * FROM large_table

# 错误
[DB_ERR] database connection timeout | Elapsed: 5s | Rows: 0 | SQL: SELECT ...
```

## 使用示例

### YAML 配置

```yaml
database:
  dsn:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    name: "mydb"
    table_prefix: "app_"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  conn_max_idle_time: 5m
  log_level: 4
  slow_threshold: 200ms
```

### 读取配置并初始化

```go
import (
    "github.com/baisiyi/go-kits/database"
    "github.com/spf13/viper"
)

func initDatabase() (*database.Client, error) {
    var cfg database.DBConfig
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return database.Init(&cfg, log.DefaultLogger)
}
```

### 事务操作

```go
func transferMoney(ctx context.Context, db *gorm.DB, from, to int64, amount float64) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 扣款
        if err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from).Error; err != nil {
            return err
        }
        // 加款
        if err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to).Error; err != nil {
            return err
        }
        return nil
    })
}
```

## 注意事项

1. **单例模式**: `Init()` 多次调用只会初始化一次，如果需要重新初始化，需要重启应用
2. **上下文传递**: 建议在所有数据库操作中传入 Context，以便支持超时和取消
3. **连接池配置**: 根据应用负载调整 `MaxOpenConns` 和 `SlowThreshold`
4. **日志级别**: 生产环境建议使用 `LogLevel: 2` (Error) 以减少日志输出

## 许可证

MIT
