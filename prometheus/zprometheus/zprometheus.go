package zprometheus

import (
	prom "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
)

type Prometheus interface {
	GetPubCli() *prom.PubClient
	GetInnerCli() *prom.InnerClient
}
