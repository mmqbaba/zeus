package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

const defaultClient = "default"

type lconfig struct {
	Name            string
	Hosts           string
	Username        string
	Password        string
	MaxPoolSize     uint64
	MaxConnIdleTime time.Duration
}

type Client struct {
	Conf *lconfig
	C    *mongo.Client
}

type ClientMgr struct {
	clients map[string]*Client
	rw      sync.RWMutex
}

func New(mconf *config.MongoDB) (c *Client, err error) {
	lconf := &lconfig{
		Hosts:           mconf.Host,
		Username:        mconf.User,
		Password:        mconf.Pwd,
		MaxPoolSize:     mconf.MaxPoolSize,
		MaxConnIdleTime: time.Duration(mconf.MaxConnIdleTime) * time.Second,
	}
	if lconf.MaxPoolSize == 0 {
		lconf.MaxPoolSize = 50
	}
	if lconf.MaxConnIdleTime == 0 {
		lconf.MaxConnIdleTime = 30 * time.Second
	}
	return newLocal(lconf)
}

func newLocal(conf *lconfig) (c *Client, err error) {
	tmp := new(Client)
	tmp.Conf = conf
	hostsStr := conf.Hosts
	dbUser := conf.Username
	dbPwd := conf.Password

	// 连接格式：mongodb://[username:password@]host1[:port1][,host2[:port2],…[,hostN[:portN]]][/[database][?options]]
	uri := fmt.Sprintf("mongodb://%s", hostsStr)
	if dbUser != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s/%s", dbUser, dbPwd, hostsStr, "admin")
	}
	log.Println("mongo.NewClient mongo uri: ", uri)
	var cl *mongo.Client
	cl, err = mongo.NewClient(
		options.Client().ApplyURI(uri),
		options.Client().SetMaxConnIdleTime(conf.MaxConnIdleTime),
		options.Client().SetMaxPoolSize(conf.MaxPoolSize),
	)
	if err != nil {
		log.Printf("mongo new client failed: %s\n", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err = cl.Connect(ctx); err != nil {
		log.Printf("mongo connect failed: %s\n", err.Error())
		return
	}
	tmp.C = cl
	c = tmp
	return
}

// DB 获取db
// name 数据库名
func (c *Client) DB(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	return c.C.Database(name, opts...)
}

func (c *Client) Release() (err error) {
	if err = c.C.Disconnect(context.Background()); err != nil {
		log.Println("mongo release err: ", err)
	}
	return
}

func newMgr() *ClientMgr {
	cs := make(map[string]*Client)
	return &ClientMgr{
		clients: cs,
	}
}

func (mgr *ClientMgr) Add(key string, c *Client) error {
	mgr.rw.Lock()
	defer mgr.rw.Unlock()
	mgr.clients[key] = c
	return nil
}

func (mgr *ClientMgr) Get(name string) *Client {
	mgr.rw.RLock()
	defer mgr.rw.RUnlock()
	if c, ok := mgr.clients[name]; ok {
		return c
	}
	return nil
}

var DefaultMgoMgr *ClientMgr

var onceDefaultInit sync.Once

func InitDefalut(conf *config.MongoDB) *ClientMgr {
	onceDefaultInit.Do(func() {
		// init mongo mgr
		DefaultMgoMgr = newMgr()
		if mgoCli, err := New(conf); err != nil {
			panic(err)
		} else {
			DefaultMgoMgr.Add(defaultClient, mgoCli)
		}
	})
	log.Println("init default mongo client")
	return DefaultMgoMgr
}

func ReloadDefault(conf *config.MongoDB) *ClientMgr {
	if DefaultMgoMgr == nil {
		log.Println("DefaultMgoMgr未初始化")
		return nil
	}

	DefaultMgoMgr.rw.Lock()
	defer DefaultMgoMgr.rw.Unlock()

	// 释放客户端
	if dc, ok := DefaultMgoMgr.clients[defaultClient]; ok {
		if err := dc.C.Disconnect(context.Background()); err != nil {
			log.Println("mongo client ReloadDefault Disconnect err: ", err)
			return DefaultMgoMgr
		}
	}

	log.Println("mongo client ReloadDefault appConf.MongoDB: ", conf)
	// init mongo mgr
	tmp := newMgr()
	if mgoCli, err := New(conf); err != nil {
		panic(err)
	} else {
		tmp.Add(defaultClient, mgoCli)
	}
	DefaultMgoMgr.clients = tmp.clients
	return DefaultMgoMgr
}

func DefaultClient() (*Client, error) {
	if DefaultMgoMgr == nil {
		return nil, errors.New("DefaultMgoMgr未初始化")
	}
	return DefaultMgoMgr.Get(defaultClient), nil
}

func DefaultClientRelease() {
	if DefaultMgoMgr == nil {
		log.Println("DefaultMgoMgr未初始化")
		return
	}
	if err := DefaultMgoMgr.Get(defaultClient).C.Disconnect(context.Background()); err != nil {
		log.Println("mongo DefaultClientRelease err: ", err)
	}
	log.Println("release default mongo client")
}
