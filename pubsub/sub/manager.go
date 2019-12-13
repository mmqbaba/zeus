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
	List map[string]*subServer
}

func NewManager(list map[string]*config.Broker) (m *Manager, err error) {
	tmp := &Manager{
		List: make(map[string]*subServer),
	}
	for k, c := range list {
		var ss *subServer
		ss, err = newS(c, nil)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		tmp.List[k] = ss
	}
	m = tmp
	return
}

func (m *Manager) Subscribe(ctx context.Context, handlers map[string]map[string]interface{}) (err error) {
	for k, ss := range m.List {
		err = ss.subscribe(ctx, handlers[k])
		if err != nil {
			return
		}
	}
	return
}

func (m *Manager) Run(ctx context.Context) {
	for _, ss := range m.List {
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
