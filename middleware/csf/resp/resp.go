package resp

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"

	// jsonpb "google.golang.org/protobuf/encoding/protojson"
	proto "github.com/golang/protobuf/proto"
	// proto "google.golang.org/protobuf/proto"
	"github.com/mmqbaba/zeus/errors"
	zeusmwhttp "github.com/mmqbaba/zeus/middleware/http"
	"github.com/mmqbaba/zeus/utils"
)

var pluginName = "CSF"

func init() {
	log.Printf("SetResponsePlugin: %s\n", pluginName)
	zeusmwhttp.SetResponsePlugin(pluginName, successResponse, errorResponse)
}

type response struct {
	errors.Error `json:"-"`
	Code         int         `json:"code"`
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	Msg          string      `json:"msg,omitempty"`
}

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

func (e response) toJSONString() string {
	if p, ok := e.Data.(proto.Message); ok {
		bf := bytesBuffPool.Get().(*bytes.Buffer)
		defer bytesBuffPool.Put(bf)
		bf.Reset()
		jsonPBMarshaler.Marshal(bf, p)
		tmpl := `{"code":%d,"msg":"%s","success":%v,"data":%s}`
		ret := fmt.Sprintf(tmpl, e.Code, e.Msg, e.Success, bf.Bytes())
		return ret
	}
	b, err := utils.Marshal(e)
	if err != nil {
		fmt.Printf("csf Response.toJSONString, utils.Marshal(e) err: %s\n", err)
	}
	return string(b)
}

func (e response) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode())
	_, err := w.Write([]byte(e.toJSONString() + "\n"))
	return err
}

func new(err *errors.Error) *response {
	code := int(err.ErrCode)
	if err.ErrCode == 0 {
		code = 200
	}
	return &response{
		Error: *err,
		Code:  code,
		Msg:   err.ErrMsg,
	}
}

func successResponse(c *gin.Context, rsp interface{}) {
	logger := zeusmwhttp.ExtractLogger(c)
	logger.Debug("CSF successResponse")

	res := new(errors.ECodeSuccessed.ParseErr(""))
	res.Success = true
	res.Data = rsp
	if err := res.Write(c.Writer); err != nil {
		logger.Errorf("CSF successResponse res.Write err: %s\n", err)
	}
}

func errorResponse(c *gin.Context, err error) {
	logger := zeusmwhttp.ExtractLogger(c)
	logger.Debug("CSF errorResponse")
	zeusErr := errors.AssertError(err)
	if zeusErr == nil {
		zeusErr = errors.New(errors.ECodeSystem, "err was a nil error or was a nil *zeuserrors.Error", "errors.AssertError")
	}
	res := new(zeusErr)
	res.Success = false
	c.Set(zeusmwhttp.ZEUS_HTTP_ERR, err)

	if err := res.Write(c.Writer); err != nil {
		logger.Errorf("CSF errorResponse zeusErr.Write err: %s\n", err)
	}
}
