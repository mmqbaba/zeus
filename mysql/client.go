package mysqlclient

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"log"
	"sync"
	"time"
)

const driverName = "mysql"

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

func InitClient(sqlconf *conf.Mysql) *Client {
	rds := new(Client)
	rds.client = newMysqlClient(sqlconf)
	return rds
}

func (dbs *Client) GetCli() *gorm.DB {
	return dbs.client
}

func newMysqlClient(cfg *conf.Mysql) *gorm.DB {
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
	//全局禁用表名复数
	_db.SingularTable(true) //如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	_db.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	_db.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	_db.DB().SetConnMaxLifetime(cfg.ConnMaxLifetime)

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
