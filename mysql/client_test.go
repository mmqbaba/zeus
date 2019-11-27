package mysqlclient

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

func TestDBSession_QuerySql(t *testing.T) {
	cfg := map[string]conf.MysqlDB{
		"e_seal": conf.MysqlDB{
			DataSourceName: "root:123456@tcp(localhost:3306)/e_seal",
			MaxIdleConns:   30,
			MaxOpenConns:   1024,
		},
	}
	ReloadConfig(cfg)
	session, _ := GetSession(context.Background(), "e_seal")
	type args struct {
		ctx     context.Context
		sqlStmt string
		args    []interface{}
	}
	tests := []struct {
		name       string
		session    *DBSession
		args       args
		wantResult []map[string]interface{}
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			"select",
			session,
			args{
				context.Background(),
				"select order_id from tb_sign_order",
				nil,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.session.QuerySql(tt.args.ctx, tt.args.sqlStmt, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBSession.QuerySql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(gotResult, tt.wantResult) {
			// 	t.Errorf("DBSession.QuerySql() = %v, want %v", gotResult, tt.wantResult)
			// }
			t.Logf("%v", gotResult)
		})
	}
}
