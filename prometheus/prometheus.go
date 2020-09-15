package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"log"
)

// PromClient struct zeus prometheus client & pub Business prometheus client
type PromClient struct {
	innerClient *InnerClient
	pubClient   *PubClient
	pHost       string
}

// PubClient Business prometheus client
type PubClient struct {
	BusinessErrCount  *Prom
	BusinessInfoCount *Prom
	CacheHit          *Prom
	CacheMiss         *Prom
	LibClient         *Prom
	HTTPClient        *Prom
}

type InnerClient struct {
	RPCClient  *Prom
	HTTPClient *Prom
	HTTPServer *Prom
	RPCServer  *Prom
}

// Prom struct info
type Prom struct {
	timer   *prometheus.HistogramVec
	counter *prometheus.CounterVec
	state   *prometheus.GaugeVec
}

func InitClient(cfg *config.Prometheus) *PromClient {
	prom := &PromClient{
		innerClient: &InnerClient{
			RPCClient:  New().WithTimer("zeus_rpc_client", []string{"user", "method"}).WithState("go_rpc_client_state", []string{"method", "name"}).WithCounter("go_rpc_client_code", []string{"method", "code"}),
			HTTPClient: New().WithTimer("zeus_http_client", []string{"method"}).WithState("go_http_client_state", []string{"method", "name"}).WithCounter("go_http_client_code", []string{"method", "code"}),
			HTTPServer: New().WithTimer("zeus_http_server", []string{"user", "method"}).WithCounter("go_http_server_code", []string{"user", "method", "code"}),
			RPCServer:  New().WithTimer("zeus_rpc_server", []string{"user", "method"}).WithCounter("go_rpc_server_code", []string{"user", "method", "code"}),
		},
		pHost: cfg.PullHost,
	}
	prom.pubClient = newPrometheusClient(prom.innerClient)
	return prom
}

func newPrometheusClient(inner *InnerClient) *PubClient {
	promPubClient := &PubClient{
		LibClient:         New().WithTimer("go_lib_client", []string{"method", "sql"}).WithState("go_lib_client_state", []string{"method", "name"}).WithCounter("go_lib_client_code", []string{"method", "code"}),
		BusinessErrCount:  New().WithCounter("go_business_err_count", []string{"name"}).WithState("go_business_err_state", []string{"name"}),
		BusinessInfoCount: New().WithCounter("go_business_info_count", []string{"name"}).WithState("go_business_info_state", []string{"name"}),
		CacheHit:          New().WithCounter("go_cache_hit", []string{"name"}),
		CacheMiss:         New().WithCounter("go_cache_miss", []string{"name"}),
		HTTPClient:        inner.HTTPClient,
	}
	log.Printf("[prometheus.newPrometheusClient] success \n")
	return promPubClient
}

func (prom *PromClient) GetPubCli() *PubClient {
	return prom.pubClient
}

func (prom *PromClient) GetInnerCli() *InnerClient {
	return prom.innerClient
}

func (prom *PromClient) GetListenHost() string {
	return prom.pHost
}

// New creates a Prom instance.
func New() *Prom {
	return &Prom{}
}

// WithTimer with summary timer
func (p *Prom) WithTimer(name string, labels []string) *Prom {
	if p == nil || p.timer != nil {
		return p
	}
	p.timer = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name,
			Help: name,
		}, labels)
	prometheus.MustRegister(p.timer)
	return p
}

// WithCounter sets counter.
func (p *Prom) WithCounter(name string, labels []string) *Prom {
	if p == nil || p.counter != nil {
		return p
	}
	p.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: name,
		}, labels)
	prometheus.MustRegister(p.counter)
	return p
}

// WithState sets state.
func (p *Prom) WithState(name string, labels []string) *Prom {
	if p == nil || p.state != nil {
		return p
	}
	p.state = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: name,
		}, labels)
	prometheus.MustRegister(p.state)
	return p
}

// Timing log timing information (in milliseconds) without sampling
func (p *Prom) Timing(name string, time int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.timer != nil {
		p.timer.WithLabelValues(label...).Observe(float64(time))
	}
}

// Incr increments one stat counter without sampling
func (p *Prom) Incr(name string, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Inc()
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Inc()
	}
}

// Decr decrements one stat counter without sampling
func (p *Prom) Decr(name string, extra ...string) {
	if p.state != nil {
		label := append([]string{name}, extra...)
		p.state.WithLabelValues(label...).Dec()
	}
}

// State set state
func (p *Prom) State(name string, v int64, extra ...string) {
	if p.state != nil {
		label := append([]string{name}, extra...)
		p.state.WithLabelValues(label...).Set(float64(v))
	}
}

// Add add count    v must > 0
func (p *Prom) Add(name string, v int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Add(float64(v))
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Add(float64(v))
	}
}
