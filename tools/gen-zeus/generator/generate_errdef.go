package generator

func GenerateErrdef(PD *Generator, rootdir string) (err error) {
	header := ``
	context := `package errdef

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/enum"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

var (
	ErrOK = errors.New(enum.ECodeSuccessed, "OK", "")
)

`
	fn := GetTargetFileName(PD, "errdef", rootdir)
	return writeContext(fn, header, context, false)
}
