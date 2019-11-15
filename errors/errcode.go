package errors

import (
	"net/http"
	"strconv"
)

// ErrorCode 错误码
type ErrorCode int

func (c ErrorCode) String() string {
	return strconv.Itoa(int(c)) + ":" + ECodeMsg[c]
}

// ParseErr 错误转义
func (c ErrorCode) ParseErr(msg string) *Error {
	return New(c, msg, "")
}

// 公共库错误码使用数字1打头，为五位数字
const (
	// ECodeSuccessed 成功
	ECodeSuccessed ErrorCode = 0
	// ECodeSystem 系统错误
	ECodeSystem ErrorCode = 10001
	// ECodeSystemAPI 系统api层错误
	ECodeSystemAPI ErrorCode = 10002
	ECodeSignature ErrorCode = 10003
	// ECodeEBusAPI 请求网关ebus接口错误
	ECodeEBusAPI       ErrorCode = 10004
	ECodeBadRequest    ErrorCode = 10005
	ECodeInternal      ErrorCode = 10006
	ECodeNotFound      ErrorCode = 10007
	ECodeUnauthorized  ErrorCode = 10008
	ECodeNoPermission  ErrorCode = 10009
	ECodeInvalidParams ErrorCode = 10010
	ECodeProxyFailed   ErrorCode = 10011
)

// ECodeMsg error message
var ECodeMsg = map[ErrorCode]string{
	ECodeSuccessed:     "ok",
	ECodeSystem:        "系统错误",
	ECodeSystemAPI:     "系统api层错误",
	ECodeSignature:     "签名错误",
	ECodeEBusAPI:       "请求网关ebus接口错误",
	ECodeBadRequest:    "bad request",
	ECodeInternal:      "服务器内部错误",
	ECodeNotFound:      "未能成功匹配路由",
	ECodeUnauthorized:  "未认证的请求",
	ECodeNoPermission:  "没有权限",
	ECodeInvalidParams: "请求参数错误",
	ECodeProxyFailed:   "代理服务错误",
}

// ECodeStatus http status code
var ECodeStatus = map[ErrorCode]int{
	ECodeSuccessed:     http.StatusOK,
	ECodeSystem:        http.StatusOK,
	ECodeSystemAPI:     http.StatusOK,
	ECodeSignature:     http.StatusOK,
	ECodeEBusAPI:       http.StatusOK,
	ECodeBadRequest:    http.StatusBadRequest,
	ECodeInternal:      http.StatusInternalServerError,
	ECodeNotFound:      http.StatusNotFound,
	ECodeUnauthorized:  http.StatusUnauthorized,
	ECodeNoPermission:  http.StatusForbidden,
	ECodeInvalidParams: http.StatusOK,
	ECodeProxyFailed:   http.StatusBadRequest,
}
