/*
最外层的调用入口
*/

package log

func Infof(format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}
