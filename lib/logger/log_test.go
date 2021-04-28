package logger

import "testing"

// go test -v -run='Test_LogPtrObjField'  lib/logger/*.go

type loggerPtrObj struct {
	FieldA   string
	Children *loggerPtrObj
}

//如何确保嵌套的指针字段日志也能打印出来
func Test_LogPtrObjField(t *testing.T) {
	GetLogger().Debug(&loggerPtrObj{FieldA: "xxxx"})
}
