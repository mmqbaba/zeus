package lock

import (
	"context"
	"testing"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	redisclient "gitlab.dg.com/BackEnd/deliver/tif/zeus/redis"
)

func TestObtain(t *testing.T) {
	rdc := redisclient.InitClient(&config.Redis{
		Host: "127.0.0.1:6379",
	})
	type args struct {
		ctx    context.Context
		client RedisClient
		key    string
		opts   *Options
	}
	tests := []struct {
		name    string
		args    args
		want    *Locker
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"obtain_lock",
			args{
				context.Background(),
				rdc.GetCli(),
				"unittest:lock:",
				nil,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Obtain(tt.args.ctx, tt.args.client, tt.args.key, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Obtain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("Got = %#v", got)
		})
	}
}
