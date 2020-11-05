package pprof

import (
    "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
    "net/http"
	"net/http/pprof"
	"sync"

	"github.com/pkg/errors"
)

var (
	_perfOnce sync.Once
)


// 启动监听pprof
func startPerf( pprofc config.PProf) {
	_perfOnce.Do(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

		go func() {
			if pprofc.HostURI == "" {
				panic(errors.Errorf("pprof: http perf must be set tcp://$host:port ", pprofc.HostURI))
			}
			if err := http.ListenAndServe(pprofc.HostURI, mux); err != nil {
				panic(errors.Errorf("pprof: listen %s: error(%v)",pprofc.HostURI, err))
			}
		}()
	})
}
