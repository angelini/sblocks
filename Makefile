PLATFORM := linux-amd64
ETCD_VERSION := 3.5.7

CMD_GO_FILES := $(shell find cmd/ -type f -name '*.go')
PKG_GO_FILES := $(shell find pkg/ -type f -name '*.go')

.PHONY: install build 
.PHONY: db reset-db
.PHONY: client-list client-create client-delete

bin/etcd:
	@mkdir -p bin
	curl -fsSL -o bin/etcd.tar.gz https://github.com/etcd-io/etcd/releases/download/v$(ETCD_VERSION)/etcd-v$(ETCD_VERSION)-$(PLATFORM).tar.gz
	tar -xzf bin/etcd.tar.gz -C bin
	mv bin/etcd-v$(ETCD_VERSION)-$(PLATFORM)/etcd bin/etcd
	mv bin/etcd-v$(ETCD_VERSION)-$(PLATFORM)/etcdctl bin/etcdctl
	mv bin/etcd-v$(ETCD_VERSION)-$(PLATFORM)/etcdutl bin/etcdutl
	rm bin/etcd.tar.gz
	rm -rf bin/etcd-v$(ETCD_VERSION)-$(PLATFORM)/

install: bin/etcd

bin/sblocks: $(CMD_GO_FILES) $(PKG_GO_FILES)
	@mkdir -p bin
	go mod tidy
	go build -o $@ main.go

build: bin/sblocks

db:
	@mkdir -p tmp
	bin/etcd --name 'sblocks' --data-dir tmp/sblocks.etcd

reset-db:
	rm tmp/sblocks.etcd

ENV_NAME := "example"

client-list:
	go run main.go list -e $(ENV_NAME)

client-create:
	go run main.go create -e $(ENV_NAME) -s 3

client-delete:
	go run main.go delete