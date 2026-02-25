package log

const (
	OutputConsole = "console"
	OutputFile    = "file"

	FormatterConsole = "console"
	FormatterJson    = "json"

	DefaultLogFileName = "ap.log"
)

var defaultConfig = []OutputConfig{{
	Writer:    OutputConsole,
	Formatter: FormatterConsole,
	Level:     "info",
}}

type Config []OutputConfig

type OutputConfig struct {
	// Writer is the output of log, such as console or file.
	Writer      string      `yaml:"writer" mapstructure:"writer"`
	WriteConfig WriteConfig `yaml:"writer_config" mapstructure:"writer_config"`

	// Formatter is the format of log, such as console or json.
	Formatter    string       `yaml:"formatter" mapstructure:"formatter"`
	FormatConfig FormatConfig `yaml:"formatter_config" mapstructure:"formatter_config"`

	// Level controls the log level, like debug, info or error.
	Level string `yaml:"level" mapstructure:"level"`

	// CallerSkip controls the nesting depth of log function.
	CallerSkip int `yaml:"caller_skip" mapstructure:"caller_skip"`

	// EnableColor determines if the output is colored. The default value is false.
	EnableColor bool `yaml:"enable_color" mapstructure:"enable_color"`
}

// WriteConfig is the local file config.
type WriteConfig struct {
	// LogPath is the log path like /usr/local/trpc/log/.
	LogPath string `yaml:"log_path"`
	// Filename is the file name like trpc.log.
	Filename string `yaml:"filename"`
	// MaxAge is the max expire times(day).
	MaxAge int `yaml:"max_age"`
	// MaxBackups is the max backup files.
	MaxBackups uint `yaml:"max_backups"`
	// MaxSize is the max size of log file(MB).
	MaxSize int64 `yaml:"max_size"`
	// RotationTime is the rotation time interval (minute).
	RotationTime int `yaml:"rotation_time"`
	// TimeFormat is the time format for log file name.
	// Default is ".%Y%m%d%H%M" (精确到分钟).
	// Examples:
	//   - ".%Y%m%d" -> app.log.20260211
	//   - ".%Y%m%d%H" -> app.log.2026021122
	TimeFormat string `yaml:"time_format"`
}

type FormatConfig struct {
	// TimeFmt is the time format of log output, default as "2006-01-02 15:04:05.000" on empty.
	TimeFmt string `yaml:"time_fmt"`

	// TimeKey is the time key of log output, default as "T".
	TimeKey string `yaml:"time_key"`
	// LevelKey is the level key of log output, default as "L".
	LevelKey string `yaml:"level_key"`
	// NameKey is the name key of log output, default as "N".
	NameKey string `yaml:"name_key"`
	// CallerKey is the caller key of log output, default as "C".
	CallerKey string `yaml:"caller_key"`
	// FunctionKey is the function key of log output, default as "", which means not to print
	// function name.
	FunctionKey string `yaml:"function_key"`
	// MessageKey is the message key of log output, default as "M".
	MessageKey string `yaml:"message_key"`
	// StackTraceKey is the stack trace key of log output, default as "S".
	StacktraceKey string `yaml:"stacktrace_key"`
}
