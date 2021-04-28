package logger

var DefaultLogger loggerInterface

type loggerInterface interface {
	Debug(...interface{}) error
	Info(...interface{}) error
	Warn(...interface{}) error
	Error(...interface{}) error
}
