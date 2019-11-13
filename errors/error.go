package errors

import (
	"encoding/json"
	syserrors "errors"
	"net/http"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/enum"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

// Error ...
type Error struct {
	ErrCode  enum.ErrorCode `json:"errcode"` //错误码  五位数字
	ErrMsg   string         `json:"errmsg"`  //错误信息
	Cause    string         `json:"cause,omitempty"`
	ServerID string         `json:"serverid"` //服务ID
	TracerID string         `json:"tracerid"` //tracerID
	Data     interface{}    `json:"data,omitempty"`
}

// New new error
func New(code enum.ErrorCode, msg, cause string) *Error {
	errMsg := msg
	if utils.IsEmptyString(errMsg) {
		errMsg = enum.ECodeMsg[code]
	}
	return &Error{
		ErrCode: code,
		ErrMsg:  enum.ECodeMsg[code],
		Cause:   cause,
	}
}

// Error for the error interface
func (e Error) Error() string {
	return e.ErrMsg + " (cause: " + e.Cause + ")"
}

func (e Error) toJSONString() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// StatusCode ...
func (e Error) StatusCode() int {
	status, ok := enum.ECodeStatus[e.ErrCode]
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

func AssertError(e error) (err *Error) {
	var zeusErr *Error
	if syserrors.As(e, &zeusErr) {
		err = zeusErr
		return
	}
	err = New(enum.ECodeSystem, "", "")
	return
}
