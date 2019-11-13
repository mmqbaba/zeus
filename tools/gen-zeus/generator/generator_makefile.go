package generator

import (
	"fmt"
	"strings"
)

func GenerateMakefile(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `

.PHONY:  ALL build init autoinit

service=%s

ALL: build



build: autoinit
	go build -o $(service)_server ./cmd/app


init:
	sh build-proto.sh

autoinit:
	[ -d proto/gw ] || sh build-proto.sh

`
	context := fmt.Sprintf(tmpContext, strings.ToLower(PD.PackageName))
	fn := GetTargetFileName(PD, "makefile", rootdir)
	return writeContext(fn, header, context, false)
}
