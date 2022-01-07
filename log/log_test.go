package log

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/mmqbaba/zeus/config"
)

func BenchmarkLogInfo(b *testing.B) {
	b.Run("report", func(b *testing.B) {
		conf := &config.LogConf{
			Log:                 "file",
			Level:               "info",
			LogDir:              "./",
			Format:              "json",
			DisableReportCaller: false,
		}
		lb, _ := New(conf)
		l := lb.Logger.WithFields(logrus.Fields{
			"tag":    "benchmark",
			"logrus": true,
			"mark1":  "mark1",
			"mark2":  "mark2",
			"mark3":  "mark3",
			"mark4":  "mark4",
			"mark5":  "mark5",
		})
		for i := 0; i < b.N; i++ {
			l.Info("BenchmarktLogInfo")
		}
	})
	b.Run("not report", func(b *testing.B) {
		conf := &config.LogConf{
			Log:                 "file",
			Level:               "info",
			LogDir:              "./",
			Format:              "json",
			DisableReportCaller: true,
		}
		lb, _ := New(conf)
		l := lb.Logger.WithFields(logrus.Fields{
			"tag":    "benchmark",
			"logrus": true,
			"mark1":  "mark1",
			"mark2":  "mark2",
			"mark3":  "mark3",
			"mark4":  "mark4",
			"mark5":  "mark5",
		})
		for i := 0; i < b.N; i++ {
			l.Info("BenchmarktLogInfo")
		}
	})
}
