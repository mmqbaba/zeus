package utils

import (
	"errors"
	"strings"
	"time"
)

var timeFormats = []string{
	"2006-01",
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006.01",
	"2006.01.02",
	"2006.01.02 15:04:05",
	"2006/01",
	"2006/01/02",
	"2006/01/02 15:04:05",
	"200601",
	"20060102",
	"20060102150405",
	"2006-01-02T15:04:05Z",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
}

const (
	DefaultLayout24h = "yyyy-MM-dd HH:mm:ss"
	DefaultLayout12h = "yyyy-MM-dd hh:mm:ss"
)

func ParseLocalTime(str string) (t time.Time, err error) {
	location := time.Now().Location()
	for _, format := range timeFormats {
		t, err = time.ParseInLocation(format, str, location)
		if err == nil {
			return
		}
	}
	err = errors.New("Can't parse string as time: " + str)
	return
}

func FormatTime(t time.Time, layout string) string {
	if IsEmptyString(layout) {
		return ""
	}
	patterns := []string{
		// 年
		"yyyy", "2006", // 完整表示的年份
		"yy", "06", // 2位数字表示的年份
		// 月
		"MM", "01", // 数字表示的月份，有前导零
		// 日
		"dd", "02", // 月份中的第几天，有前导零的2位数字
		// 小时
		"hh", "03", // 12 小时格式，有前导零
		"HH", "15", // 24 小时格式，有前导零
		// 分钟
		"mm", "04", // 有前导零的分钟数
		// 秒
		"ss", "05", // 秒数，有前导零
	}
	replacer := strings.NewReplacer(patterns...)
	layout = replacer.Replace(layout)
	return t.Format(layout)
}

func UnixSecToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}
