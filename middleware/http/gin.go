package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

func NotFound(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ExtractLogger(c).Debugf("url not found url: %s\n", c.Request.URL)
		c.JSON(http.StatusNotFound, "not found")
	}
}

func Access(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := ng.GetContainer().GetLogger()
		ctx := context.Background()
		l := logger.WithFields(logrus.Fields{"tag": "gin"})
		ctx = zeusctx.LoggerToContext(ctx, l)
		c.Set("zeusctx", ctx)
		l.Debug("access start", c.Request.URL.Path)
		c.Next()
		l.Debug("access end", c.Request.URL.Path)
	}
}

func ExtractLogger(c *gin.Context) *logrus.Entry {
	ctx := context.Background()
	if cc, ok := c.Value("zeusctx").(context.Context); ok && cc != nil {
		ctx = cc
	}
	return zeusctx.ExtractLogger(ctx)
}
