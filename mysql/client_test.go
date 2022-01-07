package mysqlclient

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	conf "github.com/mmqbaba/zeus/config"
)

func TestDBSession_QuerySql(t *testing.T) {
	cfg := map[string]conf.MysqlDB{
		"world": conf.MysqlDB{
			DataSourceName: "root:root@tcp(localhost:3306)/world",
			MaxIdleConns:   30,
			MaxOpenConns:   1024,
		},
	}
	ReloadConfig(cfg)
	session, _ := GetSession(context.Background(), "world")
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
				"select * from country where Code='CHN'",
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

	tests1 := []struct {
		name       string
		session    *DBSession
		args       args
		wantResult []map[string]interface{}
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			"insert",
			session,
			args{
				context.Background(),
				"INSERT INTO `world`.`city`(`Name`, `CountryCode`, `District`) VALUES ('xxxxxxxxx', 'ABW', 'xxxxxxxxxxxx')",
				nil,
			},
			nil,
			false,
		},
		{
			"update",
			session,
			args{
				context.Background(),
				"UPDATE `world`.`city` SET `District` = 'xxxxxxxsss' WHERE `Name` = 'xxxxxxxxx'",
				nil,
			},
			nil,
			false,
		},
		{
			"delete",
			session,
			args{
				context.Background(),
				"DELETE FROM `world`.`city` WHERE `Name` = 'xxxxxxxxx'",
				nil,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests1 {
		t.Run(tt.name, func(t *testing.T) {
			v1, v2, err := tt.session.ExecSqlWithIncrementIdReturn(tt.args.ctx, tt.args.sqlStmt, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBSession.QuerySql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%v,%v", v1, v2)
		})
	}
}
