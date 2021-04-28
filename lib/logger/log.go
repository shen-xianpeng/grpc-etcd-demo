package logger

var DefaultLogger loggerInterface


type logger interface {
	Log(keyvals ...interface{}) error
}

type loggerInterface interface {
	GetStdLogger() logger
	GetErrLogger() logger
	Debug(...interface{}) error
	Info(...interface{}) error
	Warn(...interface{}) error
	Error(...interface{}) error
}
