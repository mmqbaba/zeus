package mysqlclient

import (
	"github.com/jinzhu/gorm"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeusprometheus "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
	"log"
	"time"
)

// Handler ...
type Handler func(*gorm.Scope)

// Interceptor ...
type Interceptor func(*conf.Mysql, *zeusprometheus.Prom) func(next Handler) Handler

func metricInterceptor(conf *conf.Mysql, prometheus *zeusprometheus.Prom) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(scope *gorm.Scope) {
			beg := time.Now()
			next(scope)
			cost := time.Since(beg)

			// error metric
			if scope.HasError() {
				prometheus.Incr(TypeGorm, conf.DataSourceName+"."+scope.TableName(), conf.Host, "ERR")
				// todo sql语句，需要转换成脱密状态才能记录到日志
				if scope.DB().Error != gorm.ErrRecordNotFound {
					log.Printf("mysql err (%+v) , table_(%s)", scope.DB().Error, conf.DataSourceName+"."+scope.TableName())
				} else {
					log.Printf("record not found (%+v) , table_(%s)", scope.DB().Error, conf.DataSourceName+"."+scope.TableName())
				}
			} else {
				prometheus.Incr(TypeGorm, conf.DataSourceName+"."+scope.TableName(), conf.Host, "OK")
			}
			prometheus.Timing(TypeGorm, int64(cost/time.Millisecond), conf.DataSourceName+"."+scope.TableName(), logSQL(scope.SQL, scope.SQLVars, true))
		}
	}
}

// containArgs 是否打印包含参数的sql语句
func logSQL(sql string, args []interface{}, containArgs bool) string {
	if containArgs {
		return bindSQL(sql, args)
	}
	return sql
}
