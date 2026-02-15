package database

import (
	"context"
	"time"

	"github.com/baisiyi/go-kits/log"
	"gorm.io/gorm/logger"
)

type GormLoggerAdapter struct {
	logger        log.Logger
	logLevel      logger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger 创建适配器
func NewGormLogger(l log.Logger, slowThreshold time.Duration, level int) *GormLoggerAdapter {
	return &GormLoggerAdapter{
		logger:        l,
		slowThreshold: slowThreshold,
		logLevel:      logger.LogLevel(level),
	}
}

// LogMode 实现 gorm 接口: 设置日志级别
func (l *GormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 实现 gorm 接口
func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.logger.Infof(msg, data...)
	}
}

// Warn 实现 gorm 接口
func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.logger.Warnf(msg, data...)
	}
}

// Error 实现 gorm 接口
func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.logger.Errorf(msg, data...)
	}
}

// Trace 实现 gorm 接口: 这是最关键的方法，处理 SQL 打印、慢查询和错误
func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc() // 获取 SQL 语句和受影响行数

	// 1. 记录错误 (Error)
	if err != nil && l.logLevel >= logger.Error {
		l.logger.Errorf("[DB_ERR] %s | Elapsed: %v | Rows: %d | SQL: %s", err, elapsed, rows, sql)
		return
	}

	// 2. 记录慢查询 (Warn)
	if l.slowThreshold != 0 && elapsed > l.slowThreshold && l.logLevel >= logger.Warn {
		l.logger.Warnf("[DB_SLOW] Elapsed: %v > %v | Rows: %d | SQL: %s", elapsed, l.slowThreshold, rows, sql)
		return
	}

	// 3. 记录普通 SQL (Info)
	if l.logLevel >= logger.Info {
		l.logger.Infof("[DB_SQL] Elapsed: %v | Rows: %d | SQL: %s", elapsed, rows, sql)
	}
}
