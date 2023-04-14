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
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3

internal/executorpb/%.pb.go: internal/executorpb/%.proto
	protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative $^

internal/executorpb/%_grpc.pb.go: internal/executorpb/%.proto
	protoc --experimental_allow_proto3_optional --go-grpc_out=. --go-grpc_opt=paths=source_relative $^

bin/sblocks: $(CMD_GO_FILES) $(PKG_GO_FILES)
	@mkdir -p bin
	go mod tidy
	go build -o $@ main.go

build: internal/executorpb/definition.pb.go internal/executorpb/definition_grpc.pb.go bin/sblocks

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