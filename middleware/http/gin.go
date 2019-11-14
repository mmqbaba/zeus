package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

var SuccessResponse SuccessResponseHandler = defaultSuccessResponse
var ErrorResponse ErrorResponseHandler = defaultErrorResponse

type SuccessResponseHandler func(c *gin.Context, rsp interface{})
type ErrorResponseHandler func(c *gin.Context, err error)

func NotFound(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ExtractLogger(c).Debugf("url not found url: %s\n", c.Request.URL)
		// c.JSON(http.StatusNotFound, "not found")
		c.String(http.StatusNotFound, "not found")
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

func defaultSuccessResponse(c *gin.Context, rsp interface{}) {
	logger := ExtractLogger(c)
	logger.Debug("defaultSuccessResponse")
	res := zeuserrors.New(zeuserrors.ECodeSuccessed, "", "")
	res.Data = rsp
	res.Write(c.Writer)
}

func defaultErrorResponse(c *gin.Context, err error) {
	logger := ExtractLogger(c)
	logger.Debug("defaultErrorResponse")
	zeusErr := assertError(err)
	if zeusErr == nil {
		zeusErr = zeuserrors.New(zeuserrors.ECodeSystem, "err was a nil error or was a nil *zeuserrors.Error", "assertError")
	}
	zeusErr.Write(c.Writer)
}

func assertError(e error) (err *zeuserrors.Error) {
	if e == nil {
		return
	}
	var zeusErr *zeuserrors.Error
	if errors.As(e, &zeusErr) {
		err = zeusErr
		return
	}
	err = zeuserrors.New(zeuserrors.ECodeSystemAPI, e.Error(), "assertError")
	return
}
