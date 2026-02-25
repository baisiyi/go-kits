package log

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/baisiyi/go-kits/log/rollwriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Levels is the map from string to zapcore.Level.
var Levels = map[string]zapcore.Level{
	"":      zapcore.DebugLevel,
	"trace": zapcore.DebugLevel,
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

type ZapLogger struct {
	logger *zap.Logger
}

// WriterFactory creates a zapcore.Core.
type WriterFactory interface {
	Setup(name string, dec Decoder) error
}

// WriterFactoryFunc is an adapter to allow the use of
// ordinary functions as WriterFactory.
type WriterFactoryFunc func(name string, dec Decoder) error

// Setup calls fn(name, dec)
func (fn WriterFactoryFunc) Setup(name string, dec Decoder) error {
	return fn(name, dec)
}

var (
	factoryMu   sync.RWMutex
	factories  = make(map[string]WriterFactory)
)

func init() {
	RegisterWriter(OutputConsole, WriterFactoryFunc(defaultConsoleWriterFactory))
	RegisterWriter(OutputFile, WriterFactoryFunc(defaultFileWriterFactory))
}

// RegisterWriter registers a writer factory.
func RegisterWriter(name string, factory WriterFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	factories[name] = factory
}

// GetWriter gets a registered writer factory.
func GetWriter(name string) WriterFactory {
	factoryMu.RLock()
	f := factories[name]
	factoryMu.RUnlock()
	return f
}

// Decoder decode config to OutputConfig.
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

func NewZapLog(c Config) Logger {
	return NewZapLogWithCallerSkip(c, 2)
}

// NewZapLogWithCallerSkip creates a trpc default Logger from zap.
func NewZapLogWithCallerSkip(cfg Config, callerSkip int) Logger {
	var cores []zapcore.Core
	for _, c := range cfg {
		writer := GetWriter(c.Writer)
		if writer == nil {
			panic("log: writer core: " + c.Writer + " no registered")
		}
		var decoder Decoder
		decoder.OutputConfig = &c
		if err := writer.Setup(c.Writer, decoder); err != nil {
			panic("log: writer core: " + c.Writer + " setup fail: " + err.Error())
		}
		cores = append(cores, decoder.Core)
	}
	return &ZapLogger{
		logger: zap.New(
			zapcore.NewTee(cores...),
			zap.AddCallerSkip(callerSkip),
			zap.AddCaller(),
		),
	}
}

func newEncoder(c *OutputConfig) zapcore.Encoder {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        GetLogEncoderKey("T", c.FormatConfig.TimeKey),
		LevelKey:       GetLogEncoderKey("L", c.FormatConfig.LevelKey),
		NameKey:        GetLogEncoderKey("N", c.FormatConfig.NameKey),
		CallerKey:      GetLogEncoderKey("C", c.FormatConfig.CallerKey),
		FunctionKey:    GetLogEncoderKey(zapcore.OmitKey, c.FormatConfig.FunctionKey),
		MessageKey:     GetLogEncoderKey("M", c.FormatConfig.MessageKey),
		StacktraceKey:  GetLogEncoderKey("S", c.FormatConfig.StacktraceKey),
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     NewTimeEncoder(c.FormatConfig.TimeFmt),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if c.EnableColor {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if newFormatEncoder, ok := formatEncoders[c.Formatter]; ok {
		return newFormatEncoder(encoderCfg)
	}
	// Defaults to console encoder.
	return zapcore.NewConsoleEncoder(encoderCfg)
}

var formatEncoders = map[string]NewFormatEncoder{
	FormatterConsole: zapcore.NewConsoleEncoder,
	FormatterJson:    zapcore.NewJSONEncoder,
}

// NewFormatEncoder is the function type for creating a format encoder out of an encoder config.
type NewFormatEncoder func(zapcore.EncoderConfig) zapcore.Encoder

// RegisterFormatEncoder registers a NewFormatEncoder with the specified formatName key.
// The existing formats include "console" and "json", but you can override these format encoders
// or provide a new custom one.
func RegisterFormatEncoder(formatName string, newFormatEncoder NewFormatEncoder) {
	formatEncoders[formatName] = newFormatEncoder
}

func newConsoleCore(c *OutputConfig) (zapcore.Core, zap.AtomicLevel) {
	lvl := zap.NewAtomicLevelAt(Levels[c.Level])
	return zapcore.NewCore(
		newEncoder(c),
		zapcore.Lock(os.Stdout),
		lvl), lvl
}

func newFileCore(c *OutputConfig) (zapcore.Core, zap.AtomicLevel, error) {

	opts := []rollwriter.OptionFunc{
		rollwriter.WithMaxAge(c.WriteConfig.MaxAge),
		rollwriter.WithRotationAgeDuration(time.Duration(c.WriteConfig.RotationTime) * time.Minute),
		rollwriter.WithRotationSizeMB(c.WriteConfig.MaxSize),
		rollwriter.WithRotationCount(c.WriteConfig.MaxBackups),
	}

	// 使用配置的 TimeFormat，未配置时使用默认值（由 rollwriter 内部处理）
	if c.WriteConfig.TimeFormat != "" {
		opts = append(opts, rollwriter.WithTimeFormat(c.WriteConfig.TimeFormat))
	}

	if c.WriteConfig.Filename == "" {
		c.WriteConfig.Filename = DefaultLogFileName
	}

	writer, err := rollwriter.NewRollWriter(c.WriteConfig.Filename, opts...)
	if err != nil {
		return nil, zap.AtomicLevel{}, err
	}

	// write mode.
	var ws zapcore.WriteSyncer
	ws = zapcore.AddSync(writer)
	// log level.
	lvl := zap.NewAtomicLevelAt(Levels[c.Level])
	return zapcore.NewCore(
		newEncoder(c),
		ws, lvl,
	), lvl, nil
}

// NewTimeEncoder creates a time format encoder.
func NewTimeEncoder(format string) zapcore.TimeEncoder {
	switch format {
	case "":
		return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendByteString(defaultTimeFormat(t))
		}
	case "seconds":
		return zapcore.EpochTimeEncoder
	case "milliseconds":
		return zapcore.EpochMillisTimeEncoder
	case "nanoseconds":
		return zapcore.EpochNanosTimeEncoder
	default:
		return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(format))
		}
	}
}

// defaultTimeFormat returns the default time format "2006-01-02 15:04:05.000",
// which performs better than https://pkg.go.dev/time#Time.AppendFormat.
func defaultTimeFormat(t time.Time) []byte {
	t = t.Local()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	micros := t.Nanosecond() / 1000

	buf := make([]byte, 23)
	buf[0] = byte((year/1000)%10) + '0'
	buf[1] = byte((year/100)%10) + '0'
	buf[2] = byte((year/10)%10) + '0'
	buf[3] = byte(year%10) + '0'
	buf[4] = '-'
	buf[5] = byte((month)/10) + '0'
	buf[6] = byte((month)%10) + '0'
	buf[7] = '-'
	buf[8] = byte((day)/10) + '0'
	buf[9] = byte((day)%10) + '0'
	buf[10] = ' '
	buf[11] = byte((hour)/10) + '0'
	buf[12] = byte((hour)%10) + '0'
	buf[13] = ':'
	buf[14] = byte((minute)/10) + '0'
	buf[15] = byte((minute)%10) + '0'
	buf[16] = ':'
	buf[17] = byte((second)/10) + '0'
	buf[18] = byte((second)%10) + '0'
	buf[19] = '.'
	buf[20] = byte((micros/100000)%10) + '0'
	buf[21] = byte((micros/10000)%10) + '0'
	buf[22] = byte((micros/1000)%10) + '0'
	return buf
}

// GetLogEncoderKey gets user defined log output name, uses defKey if empty.
func GetLogEncoderKey(defKey, key string) string {
	if key == "" {
		return defKey
	}
	return key
}

// 结构化日志方法
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, fields...)
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, fields...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, fields...)
}

func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, fields...)
}

func (z *ZapLogger) Panic(msg string, fields ...Field) {
	z.logger.Panic(msg, fields...)
}

// 格式化日志方法（兼容旧API）- 移除冗余的Enabled检查
func (z *ZapLogger) Infof(format string, args ...interface{}) {
	z.logger.Info(fmt.Sprintf(format, args...))
}

func (z *ZapLogger) Errorf(format string, args ...interface{}) {
	z.logger.Error(fmt.Sprintf(format, args...))
}

func (z *ZapLogger) Debugf(format string, args ...interface{}) {
	z.logger.Debug(fmt.Sprintf(format, args...))
}

func (z *ZapLogger) Warnf(format string, args ...interface{}) {
	z.logger.Warn(fmt.Sprintf(format, args...))
}

// 上下文方法
func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{logger: z.logger.With(fields...)}
}

func (z *ZapLogger) Named(name string) Logger {
	return &ZapLogger{logger: z.logger.Named(name)}
}

// Sync 实现sync接口
func (z *ZapLogger) Sync() error {
	return z.logger.Sync()
}

// defaultConsoleWriterFactory creates a console writer.
func defaultConsoleWriterFactory(name string, dec Decoder) error {
	core, lvl := newConsoleCore(dec.OutputConfig)
	dec.Core = core
	dec.ZapLevel = lvl
	return nil
}

// defaultFileWriterFactory creates a file writer.
func defaultFileWriterFactory(name string, dec Decoder) error {
	core, lvl, err := newFileCore(dec.OutputConfig)
	if err != nil {
		return err
	}
	dec.Core = core
	dec.ZapLevel = lvl
	return nil
}
