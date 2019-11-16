package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
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

func Method(rootRouter *gin.RouterGroup, subGroup map[string]*gin.RouterGroup, r *Route) {
	if utils.IsEmptyString(r.Group) {
		switch r.Method {
		case http.MethodPost:
			rootRouter.POST(r.Path, r.Mix()...)
		case http.MethodPut:
			rootRouter.PUT(r.Path, r.Mix()...)
		case http.MethodGet:
			rootRouter.GET(r.Path, r.Mix()...)
		case http.MethodDelete:
			rootRouter.DELETE(r.Path, r.Mix()...)
		}
	} else {
		if g, ok := subGroup[r.Group]; !ok || g == nil {
			log.Printf("subGroup was nil, groupname: %s\n", r.Group)
		} else {
			switch r.Method {
			case http.MethodPost:
				g.POST(r.Path, r.Mix()...)
			case http.MethodPut:
				g.PUT(r.Path, r.Mix()...)
			case http.MethodGet:
				g.GET(r.Path, r.Mix()...)
			case http.MethodDelete:
				g.DELETE(r.Path, r.Mix()...)
			}
		}
	}
}
