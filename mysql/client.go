package mysqlclient

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeusprometheus "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	prometheus   *zeusprometheus.Prom
	interceptors []Interceptor
)

const (
	driverName   = "mysql"
	createOption = "create"
	updateOption = "update"
	delOption    = "delete"
	findOption   = "find"
	TypeGorm     = "gorm"
)

type DataSource struct {
	Host            string
	User            string
	Pwd             string
	DataSourceName  string
	CharSet         string
	ParseTime       bool
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

type Client struct {
	client *gorm.DB
	rw     sync.RWMutex
}

func (dbs *Client) Reload(cfg *conf.Mysql) {
	dbs.rw.Lock()
	defer dbs.rw.Unlock()
	if err := dbs.client.Close(); err != nil {
		log.Printf("redis close failed: %s\n", err.Error())
		return
	}
	log.Printf("[redis.Reload] redisclient reload with new conf: %+v\n", cfg)
	dbs.client = newMysqlClient(cfg)
}

func InitClientWithProm(sqlconf *conf.Mysql, promClient *zeusprometheus.Prom) *Client {
	prometheus = promClient
	mysql := new(Client)
	mysql.client = newMysqlClient(sqlconf)
	return mysql
}

func InitClient(sqlconf *conf.Mysql) *Client {
	mysql := new(Client)
	mysql.client = newMysqlClient(sqlconf)
	return mysql
}

func (dbs *Client) GetCli() *gorm.DB {
	return dbs.client
}

func (dbs *Client) ZCreate(value interface{}) *gorm.DB {
	sqlStartTime := time.Now()
	_db := dbs.client.Create(value)
	sql := strings.Join([]string{createOption, value.(string)}, ":")
	prometheus.Timing(sql, int64(time.Since(sqlStartTime)/time.Millisecond), strconv.Itoa(int(_db.RowsAffected)))
	prometheus.Incr(sql, _db.Error.Error())
	prometheus.StateIncr(sql, createOption)
	return _db
}

func (dbs *Client) ZUpdate(attrs ...interface{}) *gorm.DB {
	sqlStartTime := time.Now()
	_db := dbs.client.Update(attrs)
	sql := updateOption
	for _, attr := range attrs {
		sql = strings.Join([]string{updateOption, attr.(string)}, ":")
	}
	prometheus.Timing(sql, int64(time.Since(sqlStartTime)/time.Millisecond), strconv.Itoa(int(_db.RowsAffected)))
	prometheus.Incr(sql, _db.Error.Error())
	prometheus.StateIncr(sql, updateOption)
	return _db
}

func (dbs *Client) ZDelete(value interface{}, where ...interface{}) *gorm.DB {
	sqlStartTime := time.Now()
	_db := dbs.client.Delete(value, where)
	sql := strings.Join([]string{delOption, value.(string)}, ":")
	for _, w := range where {
		sql = strings.Join([]string{sql, w.(string)}, ":")
	}
	prometheus.Timing(sql, int64(time.Since(sqlStartTime)/time.Millisecond), strconv.Itoa(int(_db.RowsAffected)))
	prometheus.Incr(sql, _db.Error.Error())
	prometheus.StateIncr(sql, delOption)
	return _db
}

func (dbs *Client) ZFind(out interface{}, where ...interface{}) *gorm.DB {
	sqlStartTime := time.Now()
	_db := dbs.client.Find(out, where)
	sql := strings.Join([]string{findOption, out.(string)}, ":")
	for _, w := range where {
		sql = strings.Join([]string{sql, w.(string)}, ":")
	}
	prometheus.Timing(sql, int64(time.Since(sqlStartTime)/time.Millisecond), strconv.Itoa(int(_db.RowsAffected)))
	prometheus.Incr(sql, _db.Error.Error())
	prometheus.StateIncr(sql, findOption)
	return _db
}

func newMysqlClient(cfg *conf.Mysql) *gorm.DB {
	if prometheus != nil {
		interceptors = make([]Interceptor, 0)
		interceptors = append(interceptors, metricInterceptor)
	}
	_db, err := open(cfg)
	if err != nil {
		prometheus.Incr(TypeGorm, cfg.DataSourceName+".ping", cfg.Host, "open err")
		fmt.Printf("")
		log.Printf("mysql open err (%+v) , table_(%s)", err.Error(), cfg.DataSourceName+"."+".ping")
		return _db
	}

	// 监听修复BadConnections
	go func() {
		for {
			if err := _db.DB().Ping(); err != nil {
				println("mysql gorm ping err(%+v) ", err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return _db
}

// open ... with interceptors
func open(cfg *conf.Mysql) (*gorm.DB, error) {
	url := "%v:%v@(%v)/%v?charset=%v&parseTime=%v&loc=Local"
	//user:password@/dbname?charset=utf8&parseTime=True&loc=Local
	host := cfg.Host
	userName := cfg.User
	passWord := cfg.Pwd
	dbName := cfg.DataSourceName
	charSet := cfg.CharSet
	parseTime := cfg.ParseTime
	url = fmt.Sprintf(url, userName, passWord, host, dbName, charSet, parseTime)
	_db, err := gorm.Open(driverName, url)
	if err != nil {
		println("mysql gorm init err(%+v) ", err)
		panic(fmt.Sprintf("mysql gorm init  failed:%s", err.Error()))
	}

	_db.LogMode(cfg.Debug)

	//全局禁用表名复数
	_db.SingularTable(true) //如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	_db.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	_db.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	_db.DB().SetConnMaxLifetime(cfg.ConnMaxLifetime)

	replace := func(processor func() *gorm.CallbackProcessor, callbackName string, interceptors ...Interceptor) {
		old := processor().Get(callbackName)
		var handler = old
		for _, inte := range interceptors {
			handler = inte(cfg, prometheus)(handler)
		}
		processor().Replace(callbackName, handler)
	}

	replace(
		_db.Callback().Delete,
		"gorm:delete",
		interceptors...,
	)
	replace(
		_db.Callback().Update,
		"gorm:update",
		interceptors...,
	)
	replace(
		_db.Callback().Create,
		"gorm:create",
		interceptors...,
	)
	replace(
		_db.Callback().Query,
		"gorm:query",
		interceptors...,
	)
	replace(
		_db.Callback().RowQuery,
		"gorm:row_query",
		interceptors...,
	)
	return _db, err
}
