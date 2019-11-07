package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"sync"
	"time"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine/etcd"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin/container"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

const (
	retryPeriod       = 5 * time.Second
	changesBufferSize = 10
)

var confEntry *config.Entry
var confEntryPath string
var engineProvidors map[string]engine.NewEngineFn

func init() {
	log.SetPrefix("[zeus] ")
	os := runtime.GOOS
	if os == "linux" || os == "darwin" {
		confEntryPath = "/etc/tif/zeus.json"
	} else {
		confEntryPath = "C:\\tif\\zeus.json"
	}

	engineProvidors = map[string]engine.NewEngineFn{
		"etcd": newEtcdEngine,
		"file": newFileEngine,
	}
}

func newEtcdEngine() (engine.Engine, error) {
	return etcd.New(confEntry, container.GetContainer())
}

func newFileEngine() (engine.Engine, error) {
	return nil, nil
}

func Run(cnt *plugin.Container) (err error) {
	opt, err := ParseCommandLine()
	if err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}
	s := NewService(opt, container.GetContainer())

	s.initConfEntry()

	if fn, ok := engineProvidors[confEntry.EngineType]; ok && fn != nil {
		if s.ng, err = fn(); err != nil {
			log.Printf("[zeus] [service.Run] err: %s\n", err)
		}
	} else {
		err = fmt.Errorf("no newEnginFn providor for engineType: %s", confEntry.EngineType)
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}

	// 启动engine，监听状态变化，处理更新，各种组件
	if err = s.load(); err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}

	// TODO: 启动服务，服务发现注册，http/grpc
	if err = s.startServer(); err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}

	s.watcherWg.Wait()
	return
}

type Service struct {
	options        Options
	container      *plugin.Container
	ng             engine.Engine
	errorC         chan struct{}
	watcherCancelC chan struct{}
	watcherErrorC  chan struct{}
	watcherWg      sync.WaitGroup
}

func NewService(options Options, container *plugin.Container) *Service {
	s := &Service{
		options:        options,
		container:      container,
		errorC:         make(chan struct{}),
		watcherCancelC: make(chan struct{}),
		watcherErrorC:  make(chan struct{}, 1),
	}
	if !utils.IsEmptyString(options.ConfEntryPath) {
		confEntryPath = options.ConfEntryPath
	}
	return s
}

func (s *Service) initConfEntry() {
	b, err := ioutil.ReadFile(confEntryPath)
	if err != nil {
		panic(err)
	}
	confEntry = &config.Entry{}
	if err := json.Unmarshal(b, confEntry); err != nil {
		panic(err)
	}
	log.Printf("[zeus] [service.initConfEntry] confEntry: %+v", confEntry)
	return
}

// load
func (s *Service) load() (err error) {
	changesC := make(chan interface{}, changesBufferSize)
	defer close(s.errorC)
	defer close(s.watcherErrorC)

	// 启动
	if err = s.ng.Init(changesC, s.watcherCancelC, s.watcherErrorC); err != nil {
		log.Println("[zeus] [service.load] s.ng.Init err: ", err)
		return
	}

	// 处理事件数据
	s.watcherWg.Add(1)
	go func() {
		defer s.watcherWg.Done()
		for change := range changesC {
			if err := s.processChange(change); err != nil {
				log.Printf("[zeus] failed to processChange, change=%#v, err=%s\n", change, err)
			}
		}
		log.Println("[zeus] change processor shutdown")
	}()

	return
}

func (s *Service) processChange(ev interface{}) (err error) {
	switch ev.(type) {
	case config.Configer:
		log.Printf("[zeus] config change\n")
	default:
		log.Printf("[zeus] unsupported event change\n")
	}
	return
}

func (s *Service) startServer() (err error) {
	// discovery/registry
	// http/grpc
	log.Println("[zeus] start server ...")
	return
}
