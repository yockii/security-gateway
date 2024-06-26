package proxy

import (
	"github.com/rifflock/lfshook"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"security-gateway/pkg/config"
)

var log *logger.Logger

func init() {
	log = logger.New()
	log.SetLevel(logger.InfoLevel)

	d := config.GetString("logger.dir")
	if d == "" {
		d = "logs"
	}
	p, _ := filepath.Abs(d)
	p = path.Join(p, "proxy_trace")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if os.MkdirAll(p, os.ModePerm) != nil {
			logger.Warn("创建代理跟踪日志文件夹失败!")
			return
		}
	}

	rotatedNum := config.GetInt("logger.proxyBackups")
	if rotatedNum <= 0 {
		rotatedNum = 100
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
	log.AddHook(lfHook)
}

func writer(logPath, level string, rotatedNum int) io.Writer {
	logFullPath := path.Join(logPath, level)
	logier := &lumberjack.Logger{
		Filename:   logFullPath + ".log",
		MaxSize:    100, // megabytes
		MaxBackups: rotatedNum,
		MaxAge:     30, //days
		Compress:   true,
	}
	return logier
}
