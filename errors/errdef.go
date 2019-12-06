// Code generated by zeus-gen. DO NOT EDIT.
package errors

import (
	"net/http"
)

// 每个子项目特有的错误码定义，避免使用 0 ~ 19999，与公共库冲突
const (
	ECodeSuccessed                ErrorCode = 0
	ECodeSystem                   ErrorCode = 10001
	ECodeSystemAPI                ErrorCode = 10002
	ECodeSignature                ErrorCode = 10003
	ECodeBadRequest               ErrorCode = 10004
	ECodeBadUserID                ErrorCode = 10005
	ECodeBadToken                 ErrorCode = 10006
	ECodeBadAuthWay               ErrorCode = 10007
	ECodeNotRealname              ErrorCode = 10008
	ECodeNeedPerAuth              ErrorCode = 10009
	ECodeNeedOrgAuth              ErrorCode = 10010
	ECodeInvalidParams            ErrorCode = 10011
	ECodeNoPermission             ErrorCode = 10012
	ECodeEBusAPI                  ErrorCode = 10013
	ECodeRedisErr                 ErrorCode = 10014
	ECodeMongoErr                 ErrorCode = 10015
	ECodeMysqlErr                 ErrorCode = 10016
	ECodeObsErr                   ErrorCode = 10017
	ECodeNoRecord                 ErrorCode = 10018
	ECodeNoFile                   ErrorCode = 10019
	ECodeGrpcError                ErrorCode = 10020
	ECodeUrlNotFound              ErrorCode = 10021
	ECodeUUIDErr                  ErrorCode = 10022
	ECodeMetadataErr              ErrorCode = 10023
	ECodeTypeCovert               ErrorCode = 10024
	ECodePbMarshal                ErrorCode = 10025
	ECodeBusOrSp                  ErrorCode = 10026
	ECodeAuthErr                  ErrorCode = 10027
	ECodeInvalidUserType          ErrorCode = 10028
	ECodeGetSpTokenFailed         ErrorCode = 10029
	ECodeJsonMarshal              ErrorCode = 10030
	ECodeJsonUnmarshal            ErrorCode = 10031
	ECodeTifClientRequest         ErrorCode = 10032
	ECodeCreateSpAccountFailed    ErrorCode = 10033
	ECodeTifClientGetCorpsFailed  ErrorCode = 10034
	ECodeXmlMarshal               ErrorCode = 10035
	ECodeXmlUnmarshal             ErrorCode = 10036
	ECodeLockRandToken            ErrorCode = 10037
	ECodeLockUnlockFailed         ErrorCode = 10038
	ECodeLockNotObtained          ErrorCode = 10039
	ECodeLockRefreshFailed        ErrorCode = 10040
	ECodeLockDurationExceeded     ErrorCode = 10041
	ECodeNilZeusErr               ErrorCode = 10042
	ECodeInternalFuctionCalledErr ErrorCode = 10043
	ECodePublishMsgFailed         ErrorCode = 10044
	ECodeMysqlModelParseFailed    ErrorCode = 10045
	ECodeAppCfg                   ErrorCode = 10046
	ECodeInternal                 ErrorCode = 10047
	ECodeNotFound                 ErrorCode = 10048
	ECodeUnauthorized             ErrorCode = 10049
	ECodeProxyFailed              ErrorCode = 10050
	ECodePbUnmarshal              ErrorCode = 10051
	ECodeJSONPBMarshal            ErrorCode = 10052
	ECodeJSONPBUnmarshal          ErrorCode = 10053
)

// ECodeMsg error message
var ECodeMsg = map[ErrorCode]string{
	ECodeSuccessed:                "ok",
	ECodeSystem:                   "system error",
	ECodeSystemAPI:                "system api error",
	ECodeSignature:                "signature invalid",
	ECodeBadRequest:               "bad request",
	ECodeBadUserID:                "invalid userid",
	ECodeBadToken:                 "invalid access_token",
	ECodeBadAuthWay:               "invalid auth_way",
	ECodeNotRealname:              "user not realname",
	ECodeNeedPerAuth:              "person need auth",
	ECodeNeedOrgAuth:              "org need auth",
	ECodeInvalidParams:            "invalid params",
	ECodeNoPermission:             "no permission",
	ECodeEBusAPI:                  "request ebus api failed",
	ECodeRedisErr:                 "redis error",
	ECodeMongoErr:                 "mongo error",
	ECodeMysqlErr:                 "mysql error",
	ECodeObsErr:                   "obs error",
	ECodeNoRecord:                 "no record",
	ECodeNoFile:                   "no file",
	ECodeGrpcError:                "grpc invoke error",
	ECodeUrlNotFound:              "url not found",
	ECodeUUIDErr:                  "get uuid error",
	ECodeMetadataErr:              "get metedata error",
	ECodeTypeCovert:               "type convert error",
	ECodePbMarshal:                "marshal protobuf error",
	ECodeBusOrSp:                  "busid or spid error",
	ECodeAuthErr:                  "authentication error",
	ECodeInvalidUserType:          "invalid usertype",
	ECodeGetSpTokenFailed:         "get sp token failed",
	ECodeJsonMarshal:              "json marshal error",
	ECodeJsonUnmarshal:            "json unmarshal error",
	ECodeTifClientRequest:         "tifclient request error",
	ECodeCreateSpAccountFailed:    "create sp account failed",
	ECodeTifClientGetCorpsFailed:  "tifclient request get corps failed",
	ECodeXmlMarshal:               "xml marshal error",
	ECodeXmlUnmarshal:             "xml unmarshal error",
	ECodeLockRandToken:            "gen lock random token error",
	ECodeLockUnlockFailed:         "lock unlock failed",
	ECodeLockNotObtained:          "lock not obtained",
	ECodeLockRefreshFailed:        "lock refresh failed",
	ECodeLockDurationExceeded:     "lock duration exceeded",
	ECodeNilZeusErr:               "zeuserr was nil",
	ECodeInternalFuctionCalledErr: "internal fuction call error",
	ECodePublishMsgFailed:         "publish message failed",
	ECodeMysqlModelParseFailed:    "model parse failed",
	ECodeAppCfg:                   "appcfg error",
	ECodeInternal:                 "服务器内部错误",
	ECodeNotFound:                 "未能成功匹配路由",
	ECodeUnauthorized:             "未认证的请求",
	ECodeProxyFailed:              "代理服务错误",
	ECodePbUnmarshal:              "unmarshal protobuf error",
	ECodeJSONPBMarshal:            "marshal jsonpb error",
	ECodeJSONPBUnmarshal:          "unmarshal jsonpb error",
}

// ECodeStatus http status code
var ECodeStatus = map[ErrorCode]int{
	ECodeSuccessed:                http.StatusOK,
	ECodeSystem:                   http.StatusOK,
	ECodeSystemAPI:                http.StatusOK,
	ECodeSignature:                http.StatusOK,
	ECodeBadRequest:               http.StatusOK,
	ECodeBadUserID:                http.StatusOK,
	ECodeBadToken:                 http.StatusOK,
	ECodeBadAuthWay:               http.StatusOK,
	ECodeNotRealname:              http.StatusOK,
	ECodeNeedPerAuth:              http.StatusOK,
	ECodeNeedOrgAuth:              http.StatusOK,
	ECodeInvalidParams:            http.StatusOK,
	ECodeNoPermission:             http.StatusOK,
	ECodeEBusAPI:                  http.StatusOK,
	ECodeRedisErr:                 http.StatusOK,
	ECodeMongoErr:                 http.StatusOK,
	ECodeMysqlErr:                 http.StatusOK,
	ECodeObsErr:                   http.StatusOK,
	ECodeNoRecord:                 http.StatusOK,
	ECodeNoFile:                   http.StatusOK,
	ECodeGrpcError:                http.StatusOK,
	ECodeUrlNotFound:              http.StatusOK,
	ECodeUUIDErr:                  http.StatusOK,
	ECodeMetadataErr:              http.StatusOK,
	ECodeTypeCovert:               http.StatusOK,
	ECodePbMarshal:                http.StatusOK,
	ECodeBusOrSp:                  http.StatusOK,
	ECodeAuthErr:                  http.StatusOK,
	ECodeInvalidUserType:          http.StatusOK,
	ECodeGetSpTokenFailed:         http.StatusOK,
	ECodeJsonMarshal:              http.StatusOK,
	ECodeJsonUnmarshal:            http.StatusOK,
	ECodeTifClientRequest:         http.StatusOK,
	ECodeCreateSpAccountFailed:    http.StatusOK,
	ECodeTifClientGetCorpsFailed:  http.StatusOK,
	ECodeXmlMarshal:               http.StatusOK,
	ECodeXmlUnmarshal:             http.StatusOK,
	ECodeLockRandToken:            http.StatusOK,
	ECodeLockUnlockFailed:         http.StatusOK,
	ECodeLockNotObtained:          http.StatusOK,
	ECodeLockRefreshFailed:        http.StatusOK,
	ECodeLockDurationExceeded:     http.StatusOK,
	ECodeNilZeusErr:               http.StatusOK,
	ECodeInternalFuctionCalledErr: http.StatusOK,
	ECodePublishMsgFailed:         http.StatusOK,
	ECodeMysqlModelParseFailed:    http.StatusOK,
	ECodeAppCfg:                   http.StatusOK,
	ECodeInternal:                 http.StatusInternalServerError,
	ECodeNotFound:                 http.StatusNotFound,
	ECodeUnauthorized:             http.StatusUnauthorized,
	ECodeProxyFailed:              http.StatusBadRequest,
	ECodePbUnmarshal:              http.StatusOK,
	ECodeJSONPBMarshal:            http.StatusOK,
	ECodeJSONPBUnmarshal:          http.StatusOK,
}
