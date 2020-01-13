package zhttp

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

const ZEUS_CTX = "zeusctx"
const ZEUS_HTTP_TAG_RAW_RSP = "zeus_http_tag_raw_rsp"
const ZEUS_HTTP_REWRITE_ERR = "zeus_http_rewrite_err"
const ZEUS_HTTP_REWRITE_RESPONSE = "zeus_http_rewrite_response"
const ZEUS_HTTP_ERR = "zeus_http_err"
const ZEUS_HTTP_DISABLE_PB_VALIDATE = "zeus_http_disable_pb_validate"
const ZEUS_HTTP_USE_GINBIND_VALIDATE_FOR_PB = "zeus_http_use_ginbind_validate_for_pb"
const ZEUS_HTTP_WRAP_HANDLER_CTX = "zeus_http_wrap_handler_ctx"

var zeusEngine engine.Engine
var bytesBuffPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
var jsonPBMarshaler = &jsonpb.Marshaler{
	EnumsAsInts:  true,
	EmitDefaults: true,
	OrigName:     true,
}
var jsonPBUmarshaler = &jsonpb.Unmarshaler{
	AllowUnknownFields: true,
}

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
	zeusEngine = ng
	return func(c *gin.Context) {
		logger := ng.GetContainer().GetLogger()
		ctx := c.Request.Context()
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
		bf := bytesBuffPool.Get().(*bytes.Buffer)
		defer bytesBuffPool.Put(bf)
		bf.Reset()
		blw := &bodyLogWriter{body: bf, ResponseWriter: c.Writer}
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
		////// zipkin finish
		l = l.WithFields(logrus.Fields{"tracerid": span.Context().(zipkintracer.SpanContext).TraceID.ToHex()})
		ctx = zeusctx.LoggerToContext(spnctx, l)
		ctx = zeusctx.GMClientToContext(ctx, ng.GetContainer().GetGoMicroClient())
		if ng.GetContainer().GetRedisCli() != nil {
			ctx = zeusctx.RedisToContext(ctx, ng.GetContainer().GetRedisCli().GetCli())
		}
		if ng.GetContainer().GetMongo() != nil {
			ctx = zeusctx.MongoToContext(ctx, ng.GetContainer().GetMongo())
		}

		c.Set(ZEUS_CTX, ctx)
		l.Debugln("access start", c.Request.URL.Path)
		c.Next()
		l.Debugln("access end", c.Request.URL.Path)
	}
}

func ExtractLogger(c *gin.Context) *logrus.Entry {
	ctx := c.Request.Context()
	if cc, ok := c.Value(ZEUS_CTX).(context.Context); ok && cc != nil {
		ctx = cc
	}
	return zeusctx.ExtractLogger(ctx)
}

func ExtractTracerID(c *gin.Context) string {
	ctx := c.Request.Context()
	if cc, ok := c.Value(ZEUS_CTX).(context.Context); ok && cc != nil {
		ctx = cc
	}
	span := opentracing.SpanFromContext(ctx)
	return span.Context().(zipkintracer.SpanContext).TraceID.ToHex()
}

func defaultSuccessResponse(c *gin.Context, rsp interface{}) {
	logger := ExtractLogger(c)
	logger.Debug("defaultSuccessResponse")
	if c.GetBool(ZEUS_HTTP_TAG_RAW_RSP) {
		if p, ok := rsp.(proto.Message); ok {
			bf := bytesBuffPool.Get().(*bytes.Buffer)
			defer bytesBuffPool.Put(bf)
			bf.Reset()
			jsonPBMarshaler.Marshal(bf, p)
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write(bf.Bytes())
			return
		}
		c.JSON(http.StatusOK, rsp)
		return
	}
	res := zeuserrors.New(zeuserrors.ECodeSuccessed, "", "")
	res.TracerID = ExtractTracerID(c)
	res.ServiceID = zeusEngine.GetContainer().GetServiceID()
	res.Data = rsp
	f, exists := c.Get(ZEUS_HTTP_REWRITE_RESPONSE)
	if exists && f != nil {
		c.Set(ZEUS_HTTP_REWRITE_RESPONSE, nil)
		ff, ok := f.(reWriteResponseFn)
		if ok {
			ff(c, rsp)
			return
		}
	}
	res.Write(c.Writer)
}

func defaultErrorResponse(c *gin.Context, err error) {
	logger := ExtractLogger(c)
	logger.Debug("defaultErrorResponse")
	zeusErr := assertError(err)
	if zeusErr == nil {
		zeusErr = zeuserrors.New(zeuserrors.ECodeSystem, "err was a nil error or was a nil *zeuserrors.Error", "assertError")
	}
	if utils.IsEmptyString(zeusErr.TracerID) {
		zeusErr.TracerID = ExtractTracerID(c)
	}
	if utils.IsEmptyString(zeusErr.ServiceID) {
		zeusErr.ServiceID = zeusEngine.GetContainer().GetServiceID()
	}
	c.Set(ZEUS_HTTP_ERR, err)
	f, exists := c.Get(ZEUS_HTTP_REWRITE_ERR)
	if exists && f != nil {
		c.Set(ZEUS_HTTP_REWRITE_ERR, nil)
		ff, ok := f.(reWriteErrFn)
		if ok {
			ff(c, err)
			return
		}
	}
	zeusErr.Write(c.Writer)
}

func assertError(e error) (err *zeuserrors.Error) {
	if e == nil {
		return
	}
	if utils.IsBlank(reflect.ValueOf(e)) {
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

type validator interface {
	Validate() error
}

func GenerateGinHandle(handleFunc interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		h := reflect.ValueOf(handleFunc)
		reqT := h.Type().In(1).Elem()
		rspT := h.Type().In(2).Elem()

		reqV := reflect.New(reqT)
		rspV := reflect.New(rspT)

		req := reqV.Interface()
		// 针对proto.Message进行反序列化和校验
		if pb, ok := req.(proto.Message); ok {
			if c.GetBool(ZEUS_HTTP_USE_GINBIND_VALIDATE_FOR_PB) {
				if err := c.ShouldBind(req); err != nil {
					ExtractLogger(c).Debug(err)
					ErrorResponse(c, zeuserrors.ECodeInvalidParams.ParseErr(err.Error()))
					return
				}
			} else {
				if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
					// val := make(map[string]interface{})
					// if err := c.ShouldBind(&val); err != nil {
					// 	ExtractLogger(c).Debug(err)
					// 	ErrorResponse(c, zeuserrors.ECodeJsonUnmarshal.ParseErr(err.Error()))
					// 	return
					// }
					// b, err := utils.Marshal(val)
					// if err != nil {
					// 	ExtractLogger(c).Debug(err)
					// 	ErrorResponse(c, zeuserrors.ECodeJsonMarshal.ParseErr(err.Error()))
					// 	return
					// }
					// bf := bytesBuffPool.Get().(*bytes.Buffer)
					// defer bytesBuffPool.Put(bf)
					// bf.Reset()
					// bf.Write(b)
					err := jsonPBUmarshaler.Unmarshal(c.Request.Body, pb)
					if err != nil {
						ExtractLogger(c).Debug(err)
						ErrorResponse(c, zeuserrors.ECodeJSONPBUnmarshal.ParseErr(err.Error()))
						return
					}
				} else if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
					if err := c.ShouldBindQuery(req); err != nil {
						ExtractLogger(c).Debug(err)
						ErrorResponse(c, zeuserrors.ECodeInvalidParams.ParseErr(err.Error()))
						return
					}
				} else {
					if err := c.ShouldBind(req); err != nil {
						ExtractLogger(c).Debug(err)
						ErrorResponse(c, zeuserrors.ECodeInvalidParams.ParseErr(err.Error()))
						return
					}
				}
				if !c.GetBool(ZEUS_HTTP_DISABLE_PB_VALIDATE) {
					if v, ok := req.(validator); v != nil && ok {
						if err := v.Validate(); err != nil {
							ExtractLogger(c).Debug(err)
							ErrorResponse(c, zeuserrors.ECodeInvalidParams.ParseErr(err.Error()))
							return
						}
					}
				}
			}
		} else {
			// 非proto.Message
			if err := c.ShouldBind(req); err != nil {
				ExtractLogger(c).Debug(err)
				ErrorResponse(c, zeuserrors.ECodeInvalidParams.ParseErr(err.Error()))
				return
			}
		}
		ctx := c.Request.Context()
		if cc, ok := c.Value(ZEUS_CTX).(context.Context); ok && cc != nil {
			ctx = cc
		}
		if wrapHandlerCtx, exists := c.Get(ZEUS_HTTP_WRAP_HANDLER_CTX); wrapHandlerCtx != nil && exists {
			if f, ok := wrapHandlerCtx.(wrapHandlerCtxFn); ok {
				ctx = f(c, ctx)
			}
		}
		ctxV := reflect.ValueOf(ctx)
		ret := h.Call([]reflect.Value{ctxV, reqV, rspV})
		if !ret[0].IsNil() {
			err, ok := ret[0].Interface().(error)
			if ok {
				ErrorResponse(c, err)
				return
			}
			ErrorResponse(c, zeuserrors.ECodeInternal.ParseErr("UNKNOW ERROR"))
			return
		}
		SuccessResponse(c, rspV.Interface())
	}
}

// TagRawRsp 标记返回值原样返回
func TagRawRsp(raw bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_TAG_RAW_RSP, raw)
		c.Next()
	}
}

// DisablePBValidate 对实现validate接口的pb.message禁用pb数据校验
func DisablePBValidate(b bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_DISABLE_PB_VALIDATE, b)
		c.Next()
	}
}

// UseGinBindValidateForPB 使用gin bind对pb进行数据绑定和校验
func UseGinBindValidateForPB(b bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_USE_GINBIND_VALIDATE_FOR_PB, b)
		c.Next()
	}
}

type reWriteErrFn func(c *gin.Context, err error)

// SetReWriteErrFn 自定义错误处理
func SetReWriteErrFn(f reWriteErrFn) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_REWRITE_ERR, f)
		c.Next()
	}
}

type reWriteResponseFn func(c *gin.Context, rsp interface{})

// SetReWriteResponseFn 自定义返回处理
func SetReWriteResponseFn(f reWriteResponseFn) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_REWRITE_RESPONSE, f)
		c.Next()
	}
}

type wrapHandlerCtxFn func(c *gin.Context, ctx context.Context) context.Context

// WrapHandlerCtx 包装handler ctx
func WrapHandlerCtx(f wrapHandlerCtxFn) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(ZEUS_HTTP_WRAP_HANDLER_CTX, f)
		c.Next()
	}
}
