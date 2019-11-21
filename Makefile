

.PHONY:  ALL tools

ALL:


tools: gen_zeus


gen_zeus:
	GOOS=linux go build -o tools/bin/ ./tools/gen-zeus
	GOOS=windows go build -o tools/bin/ ./tools/gen-zeus

errdef:
	gen-zeus -onlyzeuserr -errdef errors/errdef.proto -dest .