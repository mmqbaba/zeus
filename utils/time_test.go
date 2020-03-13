package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestParseLocalTime(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		wantT   time.Time
		wantErr bool
	}{
		{
			"TestParseLocalTime0",
			args{str: "20200313140534"},
			time.Now(),
			false,
		},
		{
			"TestParseLocalTime1",
			args{str: "2020-03-13 14:05:23"},
			time.Now(),
			false,
		},
		{
			"TestParseLocalTime2",
			args{str: "2020/03/13 14:05:23"},
			time.Now(),
			false,
		},
		{
			"TestParseLocalTime3",
			args{str: "2020-03-13T14:05:23Z"},
			time.Now(),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, err := ParseLocalTime(tt.args.str)
			if err != nil {
				t.Errorf(err.Error())
			}
			fmt.Println(gotT)
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("ParseLocalTime() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(gotT, tt.wantT) {
			//	t.Errorf("ParseLocalTime() gotT = %v, want %v", gotT, tt.wantT)
			//}
		})
	}
}

func TestFormatTime(t *testing.T) {
	type args struct {
		t      time.Time
		layout string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"TestFormatTime0",
			args{
				t:      time.Now(),
				layout: "yyyyMMdd HHmmss",
			},
			"",
		},
		{
			"TestFormatTime1",
			args{
				t:      time.Now(),
				layout: "yyyy-MM-dd HH:mm:ss",
			},
			"",
		},
		{
			"TestFormatTime2",
			args{
				t:      time.Now(),
				layout: "yyyy/MM/dd hh:mm:ss",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.args.t, tt.args.layout)
			fmt.Println(got)
			//if got := FormatTime(tt.args.t, tt.args.layout); got != tt.want {
			//    t.Errorf("FormatTime() = %v, want %v", got, tt.want)
			//}
		})
	}
}
