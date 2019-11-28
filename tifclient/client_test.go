package tifclient

import (
	"context"
	"testing"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

func TestTifRequest(t *testing.T) {
	InitClient(&config.AppConf{})
	type args struct {
		ctx      context.Context
		method   string
		url      string
		postData string
		info     *IdentificationInfo
	}
	tests := []struct {
		name        string
		args        args
		wantRspBody []byte
		wantStatus  int
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			"hello",
			args{
				context.Background(),
				"POST",
				"",
				"",
				nil,
			},
			nil,
			200,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// gotRspBody, gotStatus, err := TifRequest(tt.args.ctx, tt.args.method, tt.args.url, tt.args.postData, tt.args.info)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("TifRequest() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(gotRspBody, tt.wantRspBody) {
			// 	t.Errorf("TifRequest() gotRspBody = %v, want %v", gotRspBody, tt.wantRspBody)
			// }
			// if gotStatus != tt.wantStatus {
			// 	t.Errorf("TifRequest() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			// }
		})
	}
}
