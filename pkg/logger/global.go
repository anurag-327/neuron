package logger

var globalLogger Logger = NewFallbackLogger()

func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

func GetGlobalLogger() Logger {
	return globalLogger
}
