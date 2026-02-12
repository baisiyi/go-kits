package rollwriter

import (
	"io"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

const (
	_  = iota             // 忽略 0 值
	KB = 1 << (10 * iota) // 1 << 10 = 1024
	MB                    // 1 << 20
	GB                    // 1 << 30
	TB                    // 1 << 40
	PB                    // 1 << 50
	EB                    // 1 << 60
)

// WriteSyncer 定义了日志写入器需要实现的行为
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// OptionFunc 是配置选项的函数类型
type OptionFunc func(*Options)

// Options 存储轮转日志的配置选项
type Options struct {
	timeFormat    string        // 时间格式
	maxAge        time.Duration // 日志默认保留时间（Hour）
	rotationAge   time.Duration // 日志轮转时间（Hour）
	rotationSize  int64         // 日志轮转容量（Byte）
	rotationCount uint          // 日志文件数量
}

// WithTimeFormat 设置时间格式
func WithTimeFormat(format string) OptionFunc {
	return func(o *Options) {
		o.timeFormat = format
	}
}

// WithMaxAge 设置日志文件的最大保留时间
func WithMaxAge(days int) OptionFunc {
	return func(o *Options) {
		o.maxAge = time.Duration(days) * 24 * time.Hour
	}
}

// WithMaxAgeDuration 以 duration 方式设置日志文件的最大保留时间
func WithMaxAgeDuration(d time.Duration) OptionFunc {
	return func(o *Options) {
		o.maxAge = d
	}
}

// WithRotationAge 设置日志轮转的时间间隔（小时）
func WithRotationAge(hours int) OptionFunc {
	return func(o *Options) {
		o.rotationAge = time.Duration(hours) * time.Hour
	}
}

// WithRotationAgeDuration 以 duration 方式设置日志轮转的时间间隔
func WithRotationAgeDuration(d time.Duration) OptionFunc {
	return func(o *Options) {
		o.rotationAge = d
	}
}

// WithRotationSize 设置单个日志文件的最大字节数
func WithRotationSize(size int64) OptionFunc {
	return func(o *Options) {
		o.rotationSize = size
	}
}

// WithRotationSizeMB 设置单个日志文件的最大字节数(MB)
func WithRotationSizeMB(size int64) OptionFunc {
	return func(o *Options) {
		o.rotationSize = size * MB
	}
}

// WithRotationCount 设置最大保留的文件数量
func WithRotationCount(count uint) OptionFunc {
	return func(o *Options) {
		o.rotationCount = count
	}
}

// NewRollWriter 创建一个新的日志轮转写入器
func NewRollWriter(filePath string, opt ...OptionFunc) (WriteSyncer, error) {
	opts := &Options{
		timeFormat:    ".%Y%m%d%H%M",
		maxAge:        7 * 24 * time.Hour, // 默认保留 7 天
		rotationAge:   24 * time.Hour,     // 默认每天轮转
		rotationSize:  100 * MB,           // 默认 100MB 轮转
		rotationCount: 0,                  // 默认不限制数量
	}
	for _, o := range opt {
		o(opts)
	}

	// 构建 rotatelogs 选项
	options := []rotatelogs.Option{
		rotatelogs.WithLinkName(filePath),
		rotatelogs.WithRotationTime(opts.rotationAge),
		rotatelogs.WithRotationSize(opts.rotationSize),
	}

	// MaxAge 和 RotationCount 不能同时设置，优先使用 MaxAge
	if opts.maxAge > 0 {
		options = append(options, rotatelogs.WithMaxAge(opts.maxAge))
	}
	if opts.rotationCount > 0 {
		options = append(options, rotatelogs.WithRotationCount(opts.rotationCount))
	}

	rl, err := rotatelogs.New(filePath+opts.timeFormat, options...)
	if err != nil {
		return nil, err
	}

	return &wrapper{rl}, nil
}

// wrapper 包装 rotatelogs.RotateLogs 以实现 WriteSyncer 接口
type wrapper struct {
	*rotatelogs.RotateLogs
}

func (w *wrapper) Write(p []byte) (n int, err error) {
	return w.RotateLogs.Write(p)
}

func (w *wrapper) Sync() error {
	return w.RotateLogs.Close()
}
