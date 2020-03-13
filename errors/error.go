package errors

import (
	"bytes"
	syserrors "errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

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

// Error .
type Error struct {
	ErrCode   ErrorCode   `json:"errcode"` // 错误码  五位数字
	ErrMsg    string      `json:"errmsg"`  // 错误信息
	Cause     string      `json:"cause,omitempty"`
	ServiceID string      `json:"serviceid,omitempty"` // 服务ID
	TracerID  string      `json:"tracerid,omitempty"`  // tracerID
	Data      interface{} `json:"data,omitempty"`
}

// New new error
func New(code ErrorCode, msg, cause string) *Error {
	errMsg := msg
	if utils.IsEmptyString(errMsg) {
		errMsg = ECodeMsg[code]
	}
	return &Error{
		ErrCode: code,
		ErrMsg:  errMsg,
		Cause:   cause,
	}
}

// Error for the error interface
func (e Error) Error() string {
    if strings.TrimSpace(e.Cause) == "" {
        return e.ErrMsg
    }
	return e.ErrMsg + " (cause: " + e.Cause + ")"
}

func (e Error) toJSONString() string {
	if e.ErrCode != ECodeSuccessed && len(strings.TrimSpace(e.TracerID)) > 0 {
		e.ErrMsg = "[" + strings.TrimSpace(e.TracerID) + "]" + e.ErrMsg
	}
	if p, ok := e.Data.(proto.Message); ok {
		bf := bytesBuffPool.Get().(*bytes.Buffer)
		defer bytesBuffPool.Put(bf)
		bf.Reset()
		jsonPBMarshaler.Marshal(bf, p)
		tmpl := `{"errcode":%d,"errmsg":"%s","cause":"%s","serviceid":"%s","tracerid":"%s","data":%s}`
		ret := fmt.Sprintf(tmpl, e.ErrCode, e.ErrMsg, e.Cause, e.ServiceID, e.TracerID, bf.Bytes())
		return ret
	}
	b, _ := utils.Marshal(e)
	return string(b)
}

// StatusCode ...
func (e Error) StatusCode() int {
	status, ok := ECodeStatus[e.ErrCode]
	if !ok {
		status = http.StatusBadRequest
	}
	return status
}

func (e Error) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode())
	_, err := w.Write([]byte(e.toJSONString() + "\n"))
	return err
}

// ErrorCode 错误码
type ErrorCode int

func (c ErrorCode) String() string {
	return strconv.Itoa(int(c)) + ":" + ECodeMsg[c]
}

// ParseErr 错误转义
func (c ErrorCode) ParseErr(msg string) *Error {
	return New(c, msg, "")
}

func (c ErrorCode) Equal(err error) bool {
	err1 := AssertError(err)
	if err1 == nil {
		return c == ECodeSuccessed
	} else if c == ECodeSystem {
		return false
	}
	return c == err1.ErrCode
}

// AssertError .
func AssertError(e error) (err *Error) {
	if e == nil {
		return
	}
	if utils.IsBlank(reflect.ValueOf(e)) {
		return
	}
	var zeusErr *Error
	if syserrors.As(e, &zeusErr) {
		err = zeusErr
		return
	}
	err = New(ECodeSystem, e.Error(), "AssertError")
	return
}
