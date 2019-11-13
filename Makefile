

.PHONY:  ALL tools

ALL:


tools: gen_zeus


gen_zeus:
	go build -o tools/bin/ ./tools/gen-zeus