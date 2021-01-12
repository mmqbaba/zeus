package log

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/log/hook"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

type LogBuilder struct {
	Logger *logrus.Logger

	conf      config.LogConf
	formatter logrus.Formatter
	m         sync.Map
}

// New logbuilder
func New(cfg *config.LogConf) (l *LogBuilder, err error) {
	logB := &LogBuilder{
		conf: *cfg,
	}
	logger, err := newLogger(cfg)
	if err != nil {
		return
	}
	logB.Logger = logger
	if err = logB.setFormatter(); err != nil {
		return
	}
	if err = logB.setOutput(); err != nil {
		return
	}
	l = logB
	log.Println("[new logger] success.")
	return
}

func newLogger(cfg *config.LogConf) (l *logrus.Logger, err error) {
	logger := logrus.New()
	ll := cfg.Level
	if utils.IsEmptyString(ll) {
		ll = "info"
	}
	level, err := logrus.ParseLevel(ll)
	if err != nil {
		return
	}
	logger.SetLevel(level)
	logger.SetReportCaller(!cfg.DisableReportCaller)
	l = logger
	return
}

func (l *LogBuilder) setFormatter() (err error) {
	format := l.conf.Format
	if utils.IsEmptyString(format) {
		format = "text"
	}
	switch format {
	case "text":
		f := &logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			FieldMap:        logrus.FieldMap{logrus.FieldKeyMsg: "message"},
			DisableColors:   true,
		}
		l.Logger.SetFormatter(f)
		l.formatter = f
	case "json":
		f := &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			FieldMap:        logrus.FieldMap{logrus.FieldKeyMsg: "message"},
		}
		l.Logger.SetFormatter(f)
		l.formatter = f
	default:
		err = fmt.Errorf("unsupport log format: %s", format)
		return
	}
	return
}

func (l *LogBuilder) setOutput() (err error) {
	output := l.conf.Log
	if utils.IsEmptyString(output) {
		output = "console"
	}
	switch output {
	case "console":
		l.Logger.SetOutput(os.Stdout) // 标准输出（线程安全）
		l.Logger.SetNoLock()
	case "file":
		// use hook
		rotateFileOutput := make(map[logrus.Level]*rotatelogs.RotateLogs)
		for _, ll := range logrus.AllLevels {
			var o *rotatelogs.RotateLogs
			if o, err = newRotateFileOutput(&l.conf, ll); err != nil {
				return
			}
			rotateFileOutput[ll] = o
		}
		var h logrus.Hook
		if h, err = hook.NewRotateFileHook(logrus.AllLevels, l.formatter, rotateFileOutput); err != nil {
			return
		}
		l.Logger.AddHook(h)
		l.Logger.SetOutput(ioutil.Discard)
		l.Logger.SetNoLock()
	default:
		err = fmt.Errorf("unsupport log output: %s", output)
		return
	}
	return
}

func newRotateFileOutput(logConf *config.LogConf, level logrus.Level) (rl *rotatelogs.RotateLogs, err error) {
	filename := filepath.Join(logConf.LogDir, "%Y%m%d"+"."+strings.ToLower(level.String())+"_"+"%H"+".log")
	duration := time.Hour
	switch logConf.RotationTime {
	case "hour":
		duration = time.Hour
	case "day":
		duration = 24 * time.Hour
		filename = filepath.Join(logConf.LogDir, "%Y%m%d"+"."+strings.ToLower(level.String())+".log")
	}
	var maxage time.Duration = -1
	if logConf.MaxAge >= 0 {
		maxage = time.Duration(logConf.MaxAge) * time.Second
	}
	logf, err := rotatelogs.New(
		filename,
		rotatelogs.WithRotationTime(duration),
		rotatelogs.WithMaxAge(maxage), //默认每7天清除下日志文件，需要设置为rotatelogs.WithMaxAge(-1)才不会清除日志
	)
	if err != nil {
		fmt.Printf("failed to create rotatelogs: %s", err)
	}
	rl = logf
	return
}
