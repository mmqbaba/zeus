package generator

import (
	"strings"
)

func GenerateDockerfile(PD *Generator, rootdir string) (err error) {
	err = genDockerfile(PD, rootdir)
	if err != nil {
		return err
	}
	return
}

func genDockerfile(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `FROM alpine

COPY ./{PKG}_server /

WORKDIR /

RUN chmod +x {PKG}_server

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

CMD ./{PKG}_server --port 9090 --apiPort 8081 --apiInterface 127.0.0.1 --confEntryPath /zeus/{PKG}

`
	context := strings.ReplaceAll(tmpContext, "{PKG}", strings.ToLower(PD.PackageName))
	fn := GetTargetFileName(PD, "dockerfile", rootdir)
	return writeContext(fn, header, context, false)
}
