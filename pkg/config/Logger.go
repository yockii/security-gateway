package config

import (
	"fmt"
	"github.com/rifflock/lfshook"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	key                    = "publish-manage"
	TimeStampFormat        = "2006-01-02 15:04:05"
	maximumCallerDepth int = 25
	minimumCallerDepth int = 4
)

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logger.Warn("加载时区失败, 使用默认时区")
		loc = time.Local
	}
	time.Local = loc
}

func InitialLogger() {
	initLoggerDefault()

	// logger
	var logLevel logger.Level = logger.InfoLevel
	if err := logLevel.UnmarshalText([]byte(DefaultInstance.GetString("logger.level"))); err != nil {
		logger.Warnf("设置日志级别失败: %v, 将使用默认[info]级别", err)
	}
	logger.SetLevel(logLevel)
	logger.SetReportCaller(true)

	logger.AddHook(&callerHook{})
	setLoggerRotateHook()
}

func initLoggerDefault() {
	DefaultInstance.SetDefault("logger.level", "debug")
	DefaultInstance.SetDefault("moduleName", "main")
}

func setLoggerRotateHook() {
	loggerDir := GetString("logger.dir")
	if loggerDir == "" {
		loggerDir = "logs"
	}
	p, _ := filepath.Abs(loggerDir)

	p = path.Join(p, GetString("moduleName"))
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if os.MkdirAll(p, os.ModePerm) != nil {
			logger.Warn("创建日志文件夹失败!")
			return
		}
	}

	rotatedNum := GetInt("logger.backups")
	if rotatedNum <= 0 {
		rotatedNum = 7
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logger.DebugLevel: writer(p, "debug", rotatedNum),
		logger.InfoLevel:  writer(p, "info", rotatedNum),
		logger.WarnLevel:  writer(p, "warn", rotatedNum),
		logger.ErrorLevel: writer(p, "error", rotatedNum),
	}, &logger.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logger.AddHook(lfHook)

	return
}

func writer(logPath, level string, rotatedNum int) io.Writer {
	logFullPath := path.Join(logPath, level)

	logier := &lumberjack.Logger{
		Filename:   logFullPath + ".log",
		MaxSize:    GetInt("logger.maxSize", 100), // megabytes
		MaxBackups: rotatedNum,
		MaxAge:     GetInt("logger.maxAge", 30), //days
		Compress:   GetBool("logger.compress"),
	}

	return logier
}

// ///////////////// caller //////////////
type callerHook struct{}

func (c *callerHook) Levels() []logger.Level {
	return []logger.Level{
		logger.PanicLevel,
		logger.FatalLevel,
		logger.ErrorLevel,
	}
}

func (c *callerHook) Fire(entry *logger.Entry) error {
	entry.Data["caller"] = c.caller(entry)
	return nil
}

func (c *callerHook) caller(entry *logger.Entry) string {
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	var stackInfo []string
	for f, again := frames.Next(); again; f, again = frames.Next() {
		if strings.Contains(f.Function, key) {
			stackInfo = append(stackInfo, fmt.Sprintf("%s:%d\n", f.Function, f.Line))
		}
	}
	//sort.Sort(sort.Reverse(sort.StringSlice(stackInfo)))
	return strings.Join(stackInfo, "\n")
}
