package etcd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
)

var afterIdx int64

type ng struct {
	entry      *config.Entry
	configer   config.Configer
	client     *etcd.Client
	container  *plugin.Container
	context    context.Context
	cancelFunc context.CancelFunc
	options    *Options
}

type Options struct {
	context context.Context
}

type Option func(o *Options)

func New(entry *config.Entry, container *plugin.Container, opts ...Option) (engine.Engine, error) {
	n := &ng{
		entry:     entry,
		container: container,
		options:   &Options{},
	}
	for _, o := range opts {
		o(n.options)
	}
	if err := n.reconnect(); err != nil {
		log.Println(err)
		return nil, err
	}
	return n, nil
}

func (n *ng) reconnect() error {
	var client *etcd.Client
	cfg := n.getEtcdClientConfig()
	var err error
	if client, err = etcd.New(cfg); err != nil {
		return err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	n.context = ctx
	n.cancelFunc = cancelFunc

	if n.client != nil {
		// 关闭
		n.client.Close()
	}

	n.client = client
	return nil
}

func (n *ng) getEtcdClientConfig() etcd.Config {
	c := etcd.Config{
		Endpoints: n.entry.EndPoints,
	}
	if !utils.IsEmptyString(n.entry.UserName) && !utils.IsEmptyString(n.entry.Password) {
		c.Username = n.entry.UserName
		c.Password = n.entry.Password
	}
	return c
}

// loadConfig 加载初始化配置，失败则程序退出
func (n *ng) loadConfig() (err error) {
	log.Printf("[zeus] [engine.loadConfig] Begin: 加载配置，configpath: %s\n", n.entry.ConfigPath)
	if utils.IsEmptyString(n.entry.ConfigPath) {
		msg := "[zeus] [engine.loadConfig] 配置路径不能为空"
		log.Println(msg)
		err = errors.New(msg)
		return
	}
	tt := 30 * time.Second
	c, ccf := context.WithTimeout(n.context, tt)
	defer ccf()
	response, err := n.client.Get(c, n.entry.ConfigPath)
	ccf()
	if err != nil {
		log.Println(err)
		return
	}

	if len(response.Kvs) == 0 || utils.IsEmptyString(string(response.Kvs[0].Value)) {
		msg := "[zeus] [engine.loadConfig] " + n.entry.ConfigPath + " " + "配置信息为空"
		log.Println(msg)
		err = errors.New(msg)
		return
	}
	afterIdx = response.Header.Revision + 1
	content := response.Kvs[0].Value
	err = n.refreshConfig(content)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[zeus] [engine.loadConfig] End: 加载配置成功，configpath: %s\n", n.entry.ConfigPath)
	return
}

// refreshConfig 刷新配置，失败则保留原来配置，不影响当前的运行
func (n *ng) refreshConfig(content []byte) (err error) {
	log.Printf("[zeus] [engine.refreshConfig] configpath: %s，configcontent: %s\n", n.entry.ConfigPath, string(content))
	configFormat := n.entry.ConfigFormat
	if utils.IsEmptyString(configFormat) {
		configFormat = "json"
	}
	var configer config.Configer
	switch configFormat {
	case "json":
		jsoner := &config.Jsoner{}
		err = jsoner.Init(content)
		if err != nil {
			msg := "[zeus] [engine.refreshConfig] jsoner 加载配置失败，configpath: " + n.entry.ConfigPath
			log.Println(msg)
			return
		}
		configer = jsoner
	case "toml":
		msg := "[zeus] [engine.refreshConfig] toml:不支持的配置格式，configpath: " + n.entry.ConfigPath
		log.Println(msg)
		err = errors.New(msg)
		return
	default:
		msg := "[zeus] [engine.refreshConfig] " + configFormat + ":不支持的配置格式，configpath: " + n.entry.ConfigPath
		log.Println(msg)
		err = errors.New(msg)
		return
	}

	if configer != nil {
		log.Printf("[zeus] [engine.refreshConfig] 刷新配置成功，configpath: %s\n", n.entry.ConfigPath)
		n.configer = configer
		return
	}
	log.Printf("[zeus] [engine.refreshConfig] 刷新配置失败，保留原来配置，configpath: %s\n", n.entry.ConfigPath)
	return
}

func (n *ng) Init(changes chan interface{}, watcherCancelC chan struct{}, watcherErrorC chan struct{}) (err error) {
	// 初始化配置config
	// 1. 读取配置
	// 2. 初始化容器组件
    // 3. 监听配置变化

	// 读取配置
	if err = n.loadConfig(); err != nil {
		return
	}

	// 初始化容器组件
	n.container.Init(n.configer.Get())

	// 监听配置变化
	go 	func(){
		defer close(changes)
		if err := n.subscribe(changes, afterIdx, watcherCancelC); err != nil {
			log.Println("[zeus] [n.subscribe] err:", err)
			watcherErrorC <- struct{}{}
			return
		}
	}()

	return nil
}
  
func (n *ng) GetConfiger() (config.Configer, error) {
	return n.configer, nil
}

func (n *ng) GetContainer() *plugin.Container {
	return n.container
}

// subscribe 监听
func (n *ng) subscribe(changes chan interface{}, afterIdx int64, cancelC chan struct{}) error {
	watcher := etcd.NewWatcher(n.client)
	defer watcher.Close()
	log.Printf("[zeus] [engine.Subscribe] Begin watching etcd configpath: %s，afterIdx: %d\n", n.entry.ConfigPath, afterIdx)
	rch := watcher.Watch(n.context, n.entry.ConfigPath, etcd.WithRev(afterIdx))
	for wresp := range rch {
		if wresp.Canceled {
			log.Println("[zeus] [engine.Subscribe] Stop watching: graceful shutdown")
			return nil
		}
		if err := wresp.Err(); err != nil {
			log.Printf("[zeus] [engine.Subscribe] Stop watching: error: %v\n", err)
			return err
		}
		for _, ev := range wresp.Events {
			change, err := n.parseChange(ev)
			if err != nil {
				log.Printf("[zeus] [engine.Subscribe] ignore '%s', error: %s\n", eventToString(ev), err)
				continue
			}
			if change != nil {
				log.Printf("[zeus] [engine.Subscribe] configPath: %s change\n", n.entry.ConfigPath)
				select {
				case changes <- change:
				case <-cancelC:
					log.Printf("[zeus] [engine.Subscribe] cancel watch config: %s\n", n.entry.ConfigPath)
					return nil
				default: // 防止忘记消费changes导致一直阻塞
					log.Printf("[zeus] [engine.Subscribe] channel is blocked, can not push change into changes")
				}
			}
		}
	}
	return nil
}

func eventToString(e *etcd.Event) string {
	return fmt.Sprintf("%s: %v -> %v", e.Type, e.PrevKv, e.Kv)
}

// MatcherFn 匹配事件操作
type MatcherFn func(*etcd.Event) (interface{}, error)

func (n *ng) parseChange(e *etcd.Event) (interface{}, error) {
	matchers := []MatcherFn{
		n.parseConfigChange,
	}
	for _, matcher := range matchers {
		m, err := matcher(e)
		if m != nil || err != nil {
			return m, err
		}
	}
	return nil, nil
}

func (n *ng) parseConfigChange(e *etcd.Event) (interface{}, error) {
	switch e.Type {
	case etcd.EventTypePut:
		err := n.refreshConfig(e.Kv.Value)
		if err != nil {
			return e, err
		}
		// reload all plugin
		n.container.Reload(n.configer.Get())
		return n.configer, nil
	case etcd.EventTypeDelete:
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported action on the: %v %v", e.Kv.Key, e.Type)
}
