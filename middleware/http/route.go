package zhttp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mmqbaba/zeus/utils"
)

type Route struct {
	RLink       RouteLink
	Method      string
	Path        string
	Group       string
	Middlewares gin.HandlersChain
	Handle      gin.HandlerFunc
}

func (r *Route) AddMW(h ...gin.HandlerFunc) {
	r.Middlewares = append(r.Middlewares, h...)
}

func (r *Route) Mix() gin.HandlersChain {
	return append(r.Middlewares, r.Handle)
}

type RouteLink string

func (rl RouteLink) AddMW(routes map[RouteLink]*Route, h ...gin.HandlerFunc) {
	if r, ok := routes[rl]; ok && r != nil {
		r.AddMW(h...)
	}
}

func (rl RouteLink) SetGroup(routes map[RouteLink]*Route, group string) {
	if r, ok := routes[rl]; ok && r != nil {
		r.Group = group
	}
}

func Method(groups map[string]*gin.RouterGroup, r *Route) {
	group := r.Group
	if utils.IsEmptyString(group) {
		group = "default"
	}
	g, ok := groups[group]
	if !ok || g == nil {
		panic(fmt.Errorf("the grouprouter was nil or not in groups, groupname: %s", group))
	}
	switch r.Method {
	case http.MethodPost:
		g.POST(r.Path, r.Mix()...)
	case http.MethodPut:
		g.PUT(r.Path, r.Mix()...)
	case http.MethodGet:
		g.GET(r.Path, r.Mix()...)
	case http.MethodDelete:
		g.DELETE(r.Path, r.Mix()...)
	default:
		panic(fmt.Errorf("unsupport the method, methodname: %s", r.Method))
	}
}

type CustomRouteFn func(routes map[RouteLink]*Route)
