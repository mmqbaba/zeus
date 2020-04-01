package hook

import (
	"errors"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type Hook struct {
	levels    []logrus.Level
	formatter logrus.Formatter
	output    map[logrus.Level]*rotatelogs.RotateLogs
}

func NewRotateFileHook(levels []logrus.Level, formatter logrus.Formatter, output map[logrus.Level]*rotatelogs.RotateLogs) (*Hook, error) {
	hook := &Hook{
		levels,
		formatter,
		output,
	}
	return hook, nil
}

func (hook *Hook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *Hook) Fire(entry *logrus.Entry) (err error) {
	var b []byte
	if b, err = hook.formatter.Format(entry); err != nil {
		return err
	}
	o, ok := hook.output[entry.Level]
	if !ok || o == nil {
		err = errors.New("rotatelogs was nil.")
		return
	}
	if _, err = o.Write(b); err != nil {
		return
	}
	return
}
