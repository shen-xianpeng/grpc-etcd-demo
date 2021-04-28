package logger

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"

	kitlog "github.com/go-kit/kit/log"
)

type wrapLogger struct {
	stdLogger logger //kitlog.Logger
	errLogger logger
}

func (w *wrapLogger) GetStdLogger() logger {
	return w.stdLogger
}

func (w *wrapLogger) GetErrLogger() logger {
	return w.errLogger
}

func FileLineFuncNameCaller(depth int) kitlog.Valuer {
	return func() interface{} {
		pc, file, line, _ := runtime.Caller(depth)
		funcName := runtime.FuncForPC(pc).Name()
		idx := strings.LastIndexByte(file, '/')
		// using idx+1 below handles both of following cases:
		// idx == -1 because no "/" was found, or
		// idx >= 0 and we want to start at the character after the found "/".
		funcNameDetailList := strings.Split(funcName, "/")
		fn := funcNameDetailList[len(funcNameDetailList)-1]
		return file[idx+1:] + "#" + strconv.Itoa(line) + fmt.Sprintf("(%s)", fn)
	}
}

func TimeFormat() kitlog.Valuer {
	return func() interface{} {
		return time.Now().Format("2006-01-02 15:04:05.999999999")
	}
}

var LogDatetime kitlog.Valuer = TimeFormat()

func InitLogger() loggerInterface {
	if DefaultLogger != nil {
		return DefaultLogger
	}
	var stdKitLogger kitlog.Logger
	{
		stdKitLogger = kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
		//kitLogger := kitlog.NewLogfmtLogger(kitlog.StdlibWriter{})
		//kitLogger = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
		//kitLogger = kitlog.With(kitLogger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)

		stdKitLogger = kitlog.WithPrefix(stdKitLogger, "time", LogDatetime) //加上前缀时间
		//日志输出时的文件和第几行代码, DefaultCaller的depth是3，因为这里打印日志时，包装了一个函数，深度depth=3+1
		stdKitLogger = kitlog.WithPrefix(stdKitLogger, "caller", FileLineFuncNameCaller(3+1))

	}
	var errKitLogger kitlog.Logger
	{
		errKitLogger = kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stderr))
		//kitLogger := kitlog.NewLogfmtLogger(kitlog.StdlibWriter{})
		//kitLogger = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
		//kitLogger = kitlog.With(kitLogger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)

		errKitLogger = kitlog.WithPrefix(errKitLogger, "time", LogDatetime) //加上前缀时间
		//日志输出时的文件和第几行代码, DefaultCaller的depth是3，因为这里打印日志时，包装了一个函数，深度depth=3+1
		errKitLogger = kitlog.WithPrefix(errKitLogger, "caller", FileLineFuncNameCaller(3+1))

	}
	DefaultLogger = &wrapLogger{stdLogger: stdKitLogger, errLogger: errKitLogger}
	return DefaultLogger
}

func formatArgs(args ...interface{}) string {
	argsStringSlice := []string{}
	for _, i := range args {
		v := reflect.ValueOf(i)
		if v.IsValid() == true {
			typ := v.Type()
			if err, ok := i.(error); ok {
				argsStringSlice = append(argsStringSlice, string(err.Error()))
			} else if typ.Kind() == reflect.Ptr {
				argsStringSlice = append(argsStringSlice, fmt.Sprintf("%#v", v.Elem()))
			} else {
				argsStringSlice = append(argsStringSlice, fmt.Sprintf("%#v", i))
			}

		} else {
			argsStringSlice = append(argsStringSlice, fmt.Sprintf("%#v", args))

		}
	}
	return strings.Join(argsStringSlice, " ")

}

//keyvals必须是两个的倍数，key, val, key, val 这样成对出现
func (m *wrapLogger) Debug(keyvals ...interface{}) error {

	return level.Debug(m.stdLogger).Log(keyvals...)
}

func (m *wrapLogger) Info(keyvals ...interface{}) error {

	return level.Info(m.stdLogger).Log(keyvals...)
}

func (m *wrapLogger) Warn(keyvals ...interface{}) error {

	return level.Warn(m.errLogger).Log(keyvals...)
}

func (m *wrapLogger) Error(keyvals ...interface{}) error {

	return level.Error(m.errLogger).Log(keyvals...)
}

//keyvals必须是两个的倍数，key, val, key, val 这样成对出现
func Debug(keyvals ...interface{}) error {

	return level.Debug(DefaultLogger.GetStdLogger()).Log(keyvals...)
}

func Info(keyvals ...interface{}) error {

	return level.Info(DefaultLogger.GetStdLogger()).Log(keyvals...)
}

func Warn(keyvals ...interface{}) error {

	return level.Warn(DefaultLogger.GetErrLogger()).Log(keyvals...)
}

func Error(keyvals ...interface{}) error {
	return level.Error(DefaultLogger.GetErrLogger()).Log(keyvals...)
}

func InfoMsg(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return level.Info(DefaultLogger.GetStdLogger()).Log("msg", msg)
}

func ErrorMsg(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return level.Info(DefaultLogger.GetErrLogger()).Log("msg", msg)
}
