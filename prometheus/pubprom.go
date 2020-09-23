package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

// Prom struct info
type PromPub struct {
	timer   map[string]*prometheus.HistogramVec
	counter map[string]*prometheus.CounterVec
	state   map[string]*prometheus.GaugeVec
}

// PubClient Business prometheus client
type PubClient struct {
	BusinessErrCount  *PromPub
	BusinessInfoCount *PromPub
	CacheHit          *PromPub
	CacheMiss         *PromPub
	DbClient          *Prom
	CacheClient       *Prom
	HTTPClient        *Prom
}

func newPrometheusClient() *PubClient {
	promPubClient := &PubClient{
		DbClient:          newInner().withTimer("zeus_db_client_duration", []string{"sql", "affected_row"}).withState("zeus_db_client_state", []string{"sql", "msg"}).withCounter("zeus_db_client_counter", []string{"sql", "options"}),
		CacheClient:       newInner().withTimer("zeus_cache_client_duration", []string{"options", "key"}).withState("zeus_cache_state", []string{"options", "key", "msg"}).withCounter("go_lib_client_code", []string{"options", "key"}),
		BusinessErrCount:  New(),
		BusinessInfoCount: New(),
		CacheHit:          New(),
		CacheMiss:         New(),
		HTTPClient:        newInner().withTimer("zeus_http_client_duration", []string{"trace_id", "url"}).withState("zeus_http_client_state", []string{"url"}).withCounter("zeus_http_client_code", []string{"trace_id", "url", "err_code", "state_code"}),
	}
	log.Printf("[prometheus.newPrometheusClient] success \n")
	return promPubClient
}

// New creates a Prom pub instance.
func New() *PromPub {
	return &PromPub{
		timer:   make(map[string]*prometheus.HistogramVec),
		counter: make(map[string]*prometheus.CounterVec),
		state:   make(map[string]*prometheus.GaugeVec),
	}
}

// PromPub WithPubTimer with summary timer
func (pb *PromPub) WithPubTimer(name string, labels []string) *Prom {
	if _, ok := pb.timer[name]; ok {
		return &Prom{timer: pb.timer[name]}
	} else {
		p := &Prom{
			timer: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: name,
					Help: name,
				}, labels)}
		prometheus.MustRegister(p.timer)
		pb.timer[name] = p.timer
		return p
	}
}

//PromPub WithPubCounter sets counter.
func (pb *PromPub) WithPubCounter(name string, labels []string) *Prom {
	if _, ok := pb.counter[name]; ok {
		return &Prom{counter: pb.counter[name]}
	} else {
		p := &Prom{
			counter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: name,
					Help: name,
				}, labels)}
		prometheus.MustRegister(p.counter)
		pb.counter[name] = p.counter
		return p
	}
}

//PromPub WithPubState sets counter.
func (pb *PromPub) WithPubState(name string, labels []string) *Prom {
	if _, ok := pb.state[name]; ok {
		return &Prom{state: pb.state[name]}
	} else {
		p := &Prom{
			state: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: name,
					Help: name,
				}, labels)}
		prometheus.MustRegister(p.state)
		pb.state[name] = p.state
		return p
	}
}
