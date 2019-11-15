package http

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
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

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Access(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := ng.GetContainer().GetLogger()
		ctx := context.Background()
		l := logger.WithFields(logrus.Fields{"tag": "gin"})
		////// zipkin begin
		cfg, err := ng.GetConfiger()
		if err != nil {
			l.Error(err)
			return
		}
		name := c.Request.URL.Path
		tracer := ng.GetContainer().GetTracer()
		if tracer == nil {
			l.Error("tracer is nil")
			return
		}
		spnctx, span, err := tracer.StartSpanFromContext(ctx, name)
		if err != nil {
			l.Error(err)
			return
		}

		header, _ := utils.Marshal(c.Request.Header)
		span.SetTag("http request.header", string(header))
		span.SetTag("http request.method", c.Request.Method)
		span.SetTag("http request.url", c.Request.URL.String())

		if c.Request.Body != nil {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			if err == nil {
				span.SetTag("http request.body", string(bodyBytes))
				// Restore the io.ReadCloser to its original state
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		if cfg.Get().Trace.OnlyLogErr {
			c.Writer = blw
		}

		// before request
		defer func() {
			if blw.body.Len() > 0 && blw.body.Bytes()[0] == '{' {
				baseRsp := struct {
					Errcode int32 `json:"errcode"`
				}{}
				if err1 := utils.Unmarshal(blw.body.Bytes(), &baseRsp); err1 == nil && baseRsp.Errcode == 0 {
					return
				}
			}
			span.Finish()
		}()
		l = l.WithFields(logrus.Fields{"tracerid": span.Context().(zipkintracer.SpanContext).TraceID.ToHex()})
		ctx = zeusctx.LoggerToContext(spnctx, l)
		////// zipkin finish

		c.Set("zeusctx", ctx)
		l.Debugln("access start", c.Request.URL.Path)
		c.Next()
		l.Debugln("access end", c.Request.URL.Path)
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
