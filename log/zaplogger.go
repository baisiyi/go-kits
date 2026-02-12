package log

import (
	"fmt"
	"os"
	"time"

	"github.com/baisiyi/go-kits/log/rollwriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*具体的实现*/

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
	levels []zap.AtomicLevel
	logger *zap.Logger
}

func NewZapLog(c Config) Logger {
	return NewZapLogWithCallerSkip(c, 2)
}

// NewZapLogWithCallerSkip creates a trpc default Logger from zap.
func NewZapLogWithCallerSkip(cfg Config, callerSkip int) Logger {
	var (
		cores  []zapcore.Core
		levels []zap.AtomicLevel
	)
	for _, c := range cfg {
		writer := GetWriter(c.Writer)
		if writer == nil {
			panic("log: writer core: " + c.Writer + " no registered")
		}
		decoder := &Decoder{OutputConfig: &c}
		if err := writer.Setup(c.Writer, decoder); err != nil {
			panic("log: writer core: " + c.Writer + " setup fail: " + err.Error())
		}
		cores = append(cores, decoder.Core)
		levels = append(levels, decoder.ZapLevel)
	}
	return &ZapLogger{
		levels: levels,
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

func (z ZapLogger) Infof(format string, args ...interface{}) {
	if z.logger.Core().Enabled(zapcore.InfoLevel) {
		z.logger.Info(fmt.Sprintf(format, args...))
	}
}

func (z ZapLogger) Errorf(format string, args ...interface{}) {
	if z.logger.Core().Enabled(zapcore.ErrorLevel) {
		z.logger.Error(fmt.Sprintf(format, args...))
	}
}

func (z ZapLogger) Debugf(format string, args ...interface{}) {
	if z.logger.Core().Enabled(zapcore.DebugLevel) {
		z.logger.Debug(fmt.Sprintf(format, args...))
	}
}

func (z ZapLogger) Warnf(format string, args ...interface{}) {
	if z.logger.Core().Enabled(zapcore.WarnLevel) {
		z.logger.Warn(fmt.Sprintf(format, args...))
	}
}
