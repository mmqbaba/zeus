package zsub

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

type Manager struct {
	subs map[string]*subServer
}

type ManagerConfig struct {
	Conf map[string]*SubConfig
}

type SubConfig struct {
	BrokerConf *config.Broker
	Handlers   map[string]interface{}
}

func NewManager(mc *ManagerConfig) (m *Manager, err error) {
	tmp := &Manager{
		subs: make(map[string]*subServer),
	}
	for k, c := range mc.Conf {
		var ss *subServer
		ss, err = newS(c.BrokerConf, nil)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		ss.handlers = c.Handlers
		tmp.subs[k] = ss
	}
	m = tmp
	return
}

func (m *Manager) Subscribe(ctx context.Context) (err error) {
	for _, ss := range m.subs {
		err = ss.subscribe(ctx, ss.handlers)
		if err != nil {
			return
		}
	}
	return
}

func (m *Manager) Run(ctx context.Context) {
	for _, ss := range m.subs {
		tmp := ss
		utils.AsyncFuncSafe(ctx, func(args ...interface{}) {
			err := tmp.run(ctx)
			if err != nil {
				log.Println(err)
				return
			}
		})
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Println("=========================[Run subserver manager] waiting syscall=====================")
	log.Printf("Subserver Manager Received signal %s\n", <-ch)
}
