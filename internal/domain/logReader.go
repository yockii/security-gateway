package domain

import (
	"bufio"
	"compress/gzip"
	logger "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"security-gateway/pkg/config"
	"strconv"
	"strings"
	"time"
)

type AccessLog struct {
	CustomIp     string `json:"customIp,omitempty"`
	Domain       string `json:"domain,omitempty"`
	MaskingLevel int    `json:"maskingLevel,omitempty"`
	Path         string `json:"path,omitempty"`
	Port         uint16 `json:"port,omitempty"`
	TargetUrl    string `json:"targetUrl,omitempty"`
	Username     string `json:"username,omitempty"`
	StartTime    int64  `json:"startTime,omitempty"`
	EndTime      int64  `json:"endTime,omitempty"`
}

var TraceLogPattern = regexp.MustCompile(`time="?([^"]+)"?\s+level=info\s+msg="requesting record"\s+customIp="?([^"]+)"?\s+domain="?([^"]*)"?\s+maskingLevel=(\d?)\s+path="?([^"]*)"?\s+port=(\d+)\s+target="?([^"]*)"?\s+username="?([^"]*)"?`)

func CountDefaultFileLogs(startTime, endTime time.Time, condition *AccessLog) (count int, earliestTime, latestTime time.Time, err error) {
	loggerDir := config.GetString("logger.dir")
	if loggerDir == "" {
		loggerDir = "logs"
	}
	p, _ := filepath.Abs(loggerDir)

	lfPath := filepath.Join(p, "info.log")
	f, err := os.Open(lfPath)
	if err != nil {
		return 0, time.Time{}, time.Time{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	return CountLogsFromFile(f, startTime, endTime, condition)
}

func CountLogsFromFile(f *os.File, startTime, endTime time.Time, condition *AccessLog) (count int, earliestTime, latestTime time.Time, err error) {
	return CountLogsFromReader(f, startTime, endTime, condition)
}

func CountLogsFromGzFileName(gzFileName string, startTime, endTime time.Time, condition *AccessLog) (count int, earliestTime, latestTime time.Time, err error) {
	gzFileName = filepath.Base(gzFileName)
	var gz *os.File
	gz, err = os.Open(gzFileName)
	if err != nil {
		return
	}
	defer func(gz *os.File) {
		_ = gz.Close()
	}(gz)

	return CountLogsFromGzFile(gz, startTime, endTime, condition)
}
func CountLogsFromGzFile(gzFile *os.File, startTime, endTime time.Time, condition *AccessLog) (count int, earliestTime, latestTime time.Time, err error) {
	gz, err := gzip.NewReader(gzFile)
	if err != nil {
		return 0, time.Time{}, time.Time{}, err
	}
	defer func(gz *gzip.Reader) {
		err = gz.Close()
		if err != nil {
			logger.Error(err)
		}
	}(gz)

	return CountLogsFromReader(gz, startTime, endTime, condition)
}

func CountLogsFromReader(r io.Reader, startTime, endTime time.Time, condition *AccessLog) (count int, earliestTime, latestTime time.Time, err error) {
	count = 0
	// 记录最早的时间和最晚的时间
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := TraceLogPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			logTime, err := time.ParseInLocation("2006-01-02 15:04:05", matches[1], time.Local)
			if err != nil {
				continue
			}

			if earliestTime.IsZero() || logTime.Before(earliestTime) {
				earliestTime = logTime
			}

			if latestTime.IsZero() || logTime.After(latestTime) {
				latestTime = logTime
			}

			if logTime.After(startTime) && logTime.Before(endTime) {
				if condition != nil {
					if condition.CustomIp != "" && condition.CustomIp != matches[2] {
						continue
					}
					if condition.Domain != "" && condition.Domain != matches[3] {
						continue
					}
					if condition.MaskingLevel != 0 {
						if matches[4] == "" || condition.MaskingLevel != int(matches[4][0]-48) {
							continue
						}
					}
					if condition.Path != "" && condition.Path != matches[5] {
						continue
					}
					if condition.Port != 0 && matches[6] != strconv.FormatInt(int64(condition.Port), 10) {
						continue
					}
					if condition.TargetUrl != "" && condition.TargetUrl != matches[7] {
						continue
					}
					if condition.Username != "" && condition.Username != matches[8] {
						continue
					}
				}
				count++
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return 0, time.Time{}, time.Time{}, err
	}

	return
}

func ProxyTraceLogPath() string {
	loggerDir := config.GetString("logger.dir")
	if loggerDir == "" {
		loggerDir = "logs"
	}
	p, _ := filepath.Abs(loggerDir)
	p = filepath.Join(p, "proxy_trace")
	return p
}

func CurrentProxyTraceLogPath() string {
	return filepath.Join(ProxyTraceLogPath(), "info.log")
}

// NextProxyTraceLogGzPath 获取下个代理跟踪日志文件的路径
func NextProxyTraceLogGzPath(preFileName string) string {
	var latestTime time.Time

	if !strings.HasSuffix(preFileName, ".log") {
		t, err := time.Parse("2006-01-02T15-04-05.999.log.gz", preFileName[5:len(preFileName)-7])
		if err != nil {
			logger.Error(err)
			return ""
		}
		latestTime = t
	}

	dir := ProxyTraceLogPath()
	files, err := filepath.Glob(filepath.Join(dir, "*.log.gz"))
	if err != nil {
		logger.Error(err)
		return ""
	}
	if len(files) == 0 {
		return ""
	}
	// 找到最新的一个gz文件，文件命名规则：info-2024-07-01T08-58-56.164.log.gz
	var latestFile string
	for _, file := range files {
		fileName := filepath.Base(file)
		if !strings.HasPrefix(fileName, "info-") || !strings.HasSuffix(fileName, ".log.gz") {
			continue
		}
		t, err := time.Parse("2006-01-02T15-04-05.999.log.gz", fileName[5:len(fileName)-7])
		if err != nil {
			logger.Error(err)
			continue
		}
		if t.After(latestTime) {
			latestTime = t
			latestFile = file
		}
	}

	return latestFile
}

func CountProxyTraceLogs(startTime, endTime time.Time, condition *AccessLog) (count int, err error) {
	fileName := CurrentProxyTraceLogPath()
	var lf *os.File
	lf, err = os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer func(lf *os.File) {
		_ = lf.Close()
	}(lf)

	var earliestTime, latestTime time.Time

	count, earliestTime, latestTime, err = CountLogsFromFile(lf, startTime, endTime, condition)
	if err != nil {
		return
	}

	for !earliestTime.IsZero() && !latestTime.IsZero() && earliestTime.After(startTime) {
		fileName = NextProxyTraceLogGzPath(filepath.Base(fileName))
		if fileName == "" {
			break
		}
		gzFileName := filepath.Base(fileName)

		var c int
		c, earliestTime, latestTime, err = CountLogsFromGzFileName(gzFileName, startTime, endTime, condition)
		if err != nil {
			return
		}
		count += c
	}

	return

}
