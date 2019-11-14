package generator

func GenerateErrdef(PD *Generator, rootdir string) (err error) {
	err = genErrdef(PD, rootdir)
	if err != nil {
		return
	}

	err = genErrdefEnum(PD, rootdir)
	if err != nil {
		return
	}
	return
}

func genErrdef(PD *Generator, rootdir string) (err error) {
	header := ``
	context := `package errdef

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

var (
	ErrOK = errors.New(errors.ECodeSuccessed, "OK", "")
)

`
	fn := GetTargetFileName(PD, "errdef", rootdir)
	return writeContext(fn, header, context, false)
}

func genErrdefEnum(PD *Generator, rootdir string) error {
	header := ``
	context := `package errdef

import (
	"net/http"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

// 每个子项目特有的错误码定义，避免使用 0 ~ 19999，与公共库冲突
const (
	ECodeSampleServiceOK errors.ErrorCode = iota + 20000
	ECodeSampleServiceErr
)

func init() {
	// ECodeMsg and ECodeStatus
	errors.ECodeMsg[ECodeSampleServiceOK] = "ECodeSampleServiceOK"
	errors.ECodeStatus[ECodeSampleServiceOK] = http.StatusOK

	errors.ECodeMsg[ECodeSampleServiceErr] = "ECodeSampleServiceErr"
	errors.ECodeStatus[ECodeSampleServiceErr] = http.StatusInternalServerError
}

`
	fn := GetTargetFileName(PD, "errdef.enum", rootdir)
	return writeContext(fn, header, context, false)
}
