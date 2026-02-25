package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/baisiyi/go-kits/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DBConfig struct {
	DSN             Connect       `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	LogLevel        int           `mapstructure:"log_level" yaml:"log_level"` // 1:Silent, 2:Error, 3:Warn, 4:Info
	SlowThreshold   time.Duration `mapstructure:"slow_threshold" yaml:"slow_threshold"`
}

type Connect struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Name        string `mapstructure:"name"`
	TablePrefix string `mapstructure:"table_prefix"`
}

// ToDSN 将 Connect 转换为 MySQL DSN 字符串
func (c *Connect) ToDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Name)
}

// Client 封装了 GORM 实例，不对外直接暴露 *gorm.DB，而是通过 GetDB() 获取
type Client struct {
	db *gorm.DB
}

var (
	clientInstance *Client
	once           sync.Once
	initErr        error
)

// Init 初始化数据库连接 (单例模式)
// 即使多次调用，也只会初始化一次
func Init(cfg *DBConfig, svcLogger log.Logger) (*Client, error) {
	// 使用 sync.Once 确保线程安全的单例创建
	once.Do(func() {
		clientInstance, initErr = newClient(cfg, svcLogger)
	})

	if initErr != nil {
		return nil, initErr
	}
	return clientInstance, nil
}

// GetInstance 获取已经初始化的单例
func GetInstance() *Client {
	if clientInstance == nil {
		log.Errorf("Database client has not been initialized. Call Init() first.")
	}
	return clientInstance
}

// 内部构造函数
func newClient(cfg *DBConfig, svcLogger log.Logger) (*Client, error) {
	// A. 配置 Logger
	newLogger := NewGormLogger(
		svcLogger,
		cfg.SlowThreshold,
		cfg.LogLevel,
	)

	// B. GORM 配置
	gormConfig := &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名不加 s
		},
		// 禁用自动事务可以提升 30%+ 性能（如果你的业务逻辑已经在 repo 层手动控制事务）
		// SkipDefaultTransaction: true,
	}

	// C. 建立连接
	db, err := gorm.Open(mysql.Open(cfg.DSN.ToDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}

	// D. 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// E. 立即执行一次 Ping (Fail Fast)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	return &Client{db: db}, nil
}

// GetDB 获取 GORM 实例
// 建议必须传入 Context，以便支持 Trace 和 Timeout
func (c *Client) GetDB(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx)
}

// Health 健康检查
func (c *Client) Health(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close 优雅关闭
func (c *Client) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	log.Infof("Closing database connection pool...")
	return sqlDB.Close()
}
