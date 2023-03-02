.PHONY: setup build


CMD_GO_FILES := $(shell find cmd/ -type f -name '*.go')
PKG_GO_FILES := $(shell find pkg/ -type f -name '*.go')

bin/sblocks: $(CMD_GO_FILES) $(PKG_GO_FILES)
	@mkdir -p bin
	go mod tidy
	go build -o $@ main.go

build: bin/sblocks