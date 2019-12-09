package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var url string
var mc *Core

type Core struct {
	zapcore.LevelEnabler
	enc zapcore.Encoder
	out map[zapcore.Level]zapcore.WriteSyncer
}

func (c *Core) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	addFields(clone.enc, fields)
	return clone
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func (c *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *Core) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	_, err = c.out[ent.Level].Write(buf.Bytes())
	buf.Free()
	if err != nil {
		return err
	}
	if ent.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output. Ignore Sync
		// errors, pending a clean solution to issue #370.
		c.out[ent.Level].Sync()
	}
	return nil
}

func (c *Core) Sync() (err error) {
	for _, o := range c.out {
		if err = o.Sync(); err != nil {
			return
		}
	}
	return
}

func (c *Core) clone() *Core {
	return &Core{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		out:          c.out,
	}
}

func newRotateFileOutput(logConf *config.LogConf, level zapcore.Level) (rl *rotatelogs.RotateLogs, err error) {
	filename := filepath.Join(logConf.LogDir, "%Y%m%d"+"."+strings.ToLower(level.String())+"_"+"%H"+".log")
	duration := time.Hour
	switch logConf.RotationTime {
	case "hour":
		duration = time.Hour
	case "day":
		duration = 24 * time.Hour
		filename = filepath.Join(logConf.LogDir, "%Y%m%d"+"."+strings.ToLower(level.String())+".log")
	}
	logf, err := rotatelogs.New(
		filename,
		rotatelogs.WithRotationTime(duration),
		rotatelogs.WithMaxAge(-1), //默认每7天清除下日志文件，需要设置为rotatelogs.WithMaxAge(-1)才不会清除日志
	)
	if err != nil {
		fmt.Printf("failed to create rotatelogs: %s", err)
	}
	rl = logf
	return
}

func init() {
	url = "http://www.xxx.xx.com"
	out := make(map[zapcore.Level]zapcore.WriteSyncer)
	out[zapcore.DebugLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.InfoLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.WarnLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.ErrorLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.DPanicLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.PanicLevel] = zapcore.AddSync(os.Stdout)
	out[zapcore.FatalLevel] = zapcore.AddSync(os.Stdout)
	// lconf := &config.LogConf{
	// 	LogDir:       "./",
	// 	RotationTime: "hour",
	// }
	// debugOut, _ := newRotateFileOutput(lconf, zapcore.DebugLevel)
	// out[zapcore.DebugLevel] = zapcore.AddSync(debugOut)
	// infoOut, _ := newRotateFileOutput(lconf, zapcore.InfoLevel)
	// out[zapcore.InfoLevel] = zapcore.AddSync(infoOut)
	// warnOut, _ := newRotateFileOutput(lconf, zapcore.WarnLevel)
	// out[zapcore.WarnLevel] = zapcore.AddSync(warnOut)
	// errorOut, _ := newRotateFileOutput(lconf, zapcore.ErrorLevel)
	// out[zapcore.ErrorLevel] = zapcore.AddSync(errorOut)
	// dpanicOut, _ := newRotateFileOutput(lconf, zapcore.DPanicLevel)
	// out[zapcore.DPanicLevel] = zapcore.AddSync(dpanicOut)
	// panicOut, _ := newRotateFileOutput(lconf, zapcore.PanicLevel)
	// out[zapcore.PanicLevel] = zapcore.AddSync(panicOut)
	// fatalOut, _ := newRotateFileOutput(lconf, zapcore.FatalLevel)
	// out[zapcore.FatalLevel] = zapcore.AddSync(fatalOut)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	mc = &Core{
		LevelEnabler: zapcore.DebugLevel,
		enc:          zapcore.NewJSONEncoder(encoderConfig),
		out:          out,
	}
	// opt := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
	// 	return zapcore.NewSampler(mc, time.Second, 100, 100) // 设置采样
	// 	// return mc
	// })
	// logger, _ := zap.NewProduction(opt)
	// logger = zap.New(zapcore.NewSampler(mc, time.Second, 100, 50), zap.AddCaller()) // 设置采样，每个周期：记录前100条日志，前100条日志之后每隔50条日志才做一次记录
	logger = zap.New(mc, zap.AddCaller())
	sugar = logger.Sugar()
}

func print() {
	// defer logger.Sync() // flushes buffer, if any
	sugar.Infow("failed to fetch URL",
		// Structured context as loosely typed key-value pairs.
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
}
