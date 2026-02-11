package log

import (
	"errors"
	"fmt"
	"sync"

	"github.com/baisiyi/go-kits/plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*
策略层，替换不同的logger实现
*/

func init() {
	RegisterWriter(OutputConsole, DefaultConsoleWriterFactory)
	RegisterWriter(OutputFile, DefaultFileWriterFactory)
	Register(defaultLoggerName, NewZapLog(defaultConfig))
	plugin.Register(defaultLoggerName, DefaultLogFactory)
}

const (
	defaultLoggerName = "default"
	pluginType        = "log"
)

var (
	// DefaultLogger the default Logger. The initial output is console. When frame start, it is
	// over write by configuration.
	DefaultLogger Logger
	// DefaultLogFactory is the default log loader. Users may replace it with their own
	// implementation.
	DefaultLogFactory = &Factory{}

	mu      sync.RWMutex
	loggers = make(map[string]Logger)
)

// Register registers Logger. It supports multiple Logger implementation.
func Register(name string, logger Logger) {
	mu.Lock()
	defer mu.Unlock()
	if logger == nil {
		panic("log: Register logger is nil")
	}
	if _, dup := loggers[name]; dup && name != defaultLoggerName {
		panic("log: Register called twiced for logger name " + name)
	}
	loggers[name] = logger
	if name == defaultLoggerName {
		DefaultLogger = logger
	}
}

// GetDefaultLogger gets the default Logger.
// To configure it, set key in configuration file to default.
// The console output is the default value.
func GetDefaultLogger() Logger {
	mu.RLock()
	l := DefaultLogger
	mu.RUnlock()
	return l
}

// Get returns the Logger implementation by log name.
// log.Debug use DefaultLogger to print logs. You may also use log.Get("name").Debug.
func Get(name string) Logger {
	mu.RLock()
	l := loggers[name]
	mu.RUnlock()
	return l
}

type Decoder struct {
	OutputConfig *OutputConfig
	Core         zapcore.Core
	ZapLevel     zap.AtomicLevel
}

// Decode 作用：配置plugin，解耦plugin的配置实例和参数实例，参数实例只要实现了Decoder接口，即可在Decode方法中，将参数实例赋值给plugin的配置实例
// 如： FileWriterFactory 中，FileWriterFactory 需要配置OutputConfig，但是传入配置是Decoder
// (d Decoder) Decode(cfg interface{}) error 是 FileWriterFactory 和 ConsoleWriterFactory 使用的配置工具
func (d Decoder) Decode(cfg interface{}) error {
	output, ok := cfg.(**OutputConfig)
	if !ok {
		return fmt.Errorf("decoder config type:%T invalid, not **OutputConfig", cfg)
	}
	*output = d.OutputConfig
	return nil
}

// Factory 使用config配置生成logger
type Factory struct{}

func (f *Factory) Type() string {
	return "log"
}

func (f *Factory) Setup(name string, dec plugin.Decoder) error {
	if dec == nil {
		return errors.New("log config decoder empty")
	}
	cfg, callerSkip, err := f.setupConfig(dec)
	if err != nil {
		return err
	}
	logger := NewZapLogWithCallerSkip(cfg, callerSkip)
	if logger == nil {
		return errors.New("new zap logger fail")
	}
	Register(name, logger)
	return nil
}

func (f *Factory) setupConfig(configDec plugin.Decoder) (Config, int, error) {
	cfg := Config{}
	if err := configDec.Decode(&cfg); err != nil {
		return nil, 0, err
	}
	if len(cfg) == 0 {
		return nil, 0, errors.New("log config output empty")
	}
	// If caller skip is not configured, use 2 as default.
	callerSkip := 2
	for i := 0; i < len(cfg); i++ {
		if cfg[i].CallerSkip != 0 {
			callerSkip = cfg[i].CallerSkip
		}
	}
	return cfg, callerSkip, nil
}
