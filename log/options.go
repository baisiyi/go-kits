package log

// Option 日志配置选项
type Option interface {
	apply(cfg *[]OutputConfig)
}

type optionFunc func(cfg *[]OutputConfig)

func (f optionFunc) apply(cfg *[]OutputConfig) {
	f(cfg)
}

// WithLevel 设置日志级别
func WithLevel(level string) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].Level = level
		}
	})
}

// WithFile 设置文件输出
func WithFile(filename string) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].Writer = OutputFile
			(*cfg)[i].WriteConfig.Filename = filename
		}
	})
}

// WithMaxSize 设置单个日志文件最大大小(MB)
func WithMaxSize(maxSize int64) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].WriteConfig.MaxSize = maxSize
		}
	})
}

// WithMaxAge 设置日志文件保留天数
func WithMaxAge(maxAge int) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].WriteConfig.MaxAge = maxAge
		}
	})
}

// WithMaxBackups 设置最大备份文件数
func WithMaxBackups(maxBackups uint) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].WriteConfig.MaxBackups = maxBackups
		}
	})
}

// WithRotationTime 设置日志轮转时间(分钟)
func WithRotationTime(minutes int) Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].WriteConfig.RotationTime = minutes
		}
	})
}

// WithJSONFormatter 设置JSON格式
func WithJSONFormatter() Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].Formatter = FormatterJson
		}
	})
}

// WithConsoleFormatter 设置控制台格式
func WithConsoleFormatter() Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].Formatter = FormatterConsole
		}
	})
}

// WithColor 启用彩色输出
func WithColor() Option {
	return optionFunc(func(cfg *[]OutputConfig) {
		for i := range *cfg {
			(*cfg)[i].EnableColor = true
		}
	})
}
