package file

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin/zcontainer"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

// ng fileengine的实现，目前并不完善，不要应用到生产，可简单用在开发测试
type ng struct {
	entry      *config.Entry
	configer   config.Configer
	container  zcontainer.Container
	context    context.Context
	cancelFunc context.CancelFunc
	options    *Options
}

type Options struct {
	context context.Context
}

type Option func(o *Options)

func New(entry *config.Entry, container zcontainer.Container, opts ...Option) (engine.Engine, error) {
	n := &ng{
		entry:     entry,
		container: container,
		options:   &Options{},
	}
	for _, o := range opts {
		o(n.options)
	}

	return n, nil
}

func (n *ng) Init() (err error) {
	// 读取配置
	d, err := ioutil.ReadFile(n.entry.ConfigPath)
	if err != nil {
		return
	}

	if err = n.refreshConfig(d); err != nil {
		return
	}

	return nil
}

func (n *ng) Subscribe(changes chan interface{}, cancelC chan struct{}) error {
	for {
		time.Sleep(10 * time.Second)
		// TODO: 检测文件修改

		d, err := ioutil.ReadFile(n.entry.ConfigPath)
		if err != nil {
			log.Printf("[zeus] [engine.Subscribe] error: %s\n", err)
			continue
		}
		if err = n.refreshConfig(d); err != nil {
			log.Printf("[zeus] [engine.Subscribe] ignore '%s', error: %s\n", string(d), err)
			continue
		}
		if n.configer != nil {
			log.Printf("[zeus] [engine.Subscribe] configPath: %s change\n", n.entry.ConfigPath)
			select {
			case changes <- n.configer:
			case <-cancelC:
				log.Printf("[zeus] [engine.Subscribe] cancel watch config: %s\n", n.entry.ConfigPath)
				return nil
			default: // 防止忘记消费changes导致一直阻塞
				log.Printf("[zeus] [engine.Subscribe] channel is blocked, can not push change into changes")
			}
		}
	}
	return nil
}

func (n *ng) GetConfiger() (config.Configer, error) {
	return n.configer, nil
}

func (n *ng) GetContainer() zcontainer.Container {
	return n.container
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
