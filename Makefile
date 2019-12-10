

.PHONY:  ALL tools

build_date = $(shell date '+%Y-%m-%d %H:%M:%S')
version = $(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')
goversion = $(shell go version)
ldflags = -ldflags "-X 'main.Version=$(version)' -X 'main.BuildDate=$(build_date)' -X 'main.GoVersion=$(goversion)'"

ALL:


tools: gen_zeus


gen_zeus:
	GOOS=linux go build $(ldflags) -o tools/bin/ ./tools/gen-zeus
	GOOS=windows go build $(ldflags) -o tools/bin/ ./tools/gen-zeus

errdef:
	gen-zeus -onlyzeuserr -errdef errors/errdef.proto -dest .
