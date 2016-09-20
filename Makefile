NO_COLOR := $(shell echo "\033[0m")
OK_COLOR := $(shell echo "\033[32;01m")
ERROR_COLOR := $(shell echo "\033[31;01m")
WARN_COLOR := $(shell echo "\033[33;01m")
GOFMT=gofmt -w
DEPS=$(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
PACKAGES := $(shell go list ./...)

default: build

dep:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d

update:
	@echo "$(OK_COLOR)==> Updating all dependencies$(NO_COLOR)"
	@go get -d -u ./...
	@echo $(DEPS) | xargs -n1 go get -d -u

proto:
	@echo "$(OK_COLOR)==> Generating protocol buffers$(NO_COLOR)"
	@if ! which protoc > /dev/null; then \
		echo "$(WARN_COLOR)error: protoc not installed$(OK_COLOR)" >&2; \
		exit 1; \
	fi
	go get -u -v github.com/golang/protobuf/protoc-gen-go
	for dir in $$(git ls-files '*.proto' | xargs -n1 dirname | uniq); do \
		# use $$dir as the root for all proto files in the same directory
		protoc -I $$dir --go_out=plugins=grpc:$$dir $$dir/*.proto; \
	done

format:
	@echo "$(OK_COLOR)==> Formatting$(NO_COLOR)"
	$(foreach ENTRY,$(PACKAGES),$(GOFMT) $(GOPATH)/src/$(ENTRY);)

build:
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	go build -o ./ok ./client
	go build -o ./postd ./servers/post

clean:
	go clean -i -r -x
	rm ./ok && rm ./postd

install:
	@echo "$(OK_COLOR)==> Installing$(NO_COLOR)"
	install ./postd $(GOPATH)/bin
	install ./ok    $(GOPATH)/bin

lint:
	@echo "$(OK_COLOR)==> Linting$(NO_COLOR)"
	$(GOPATH)/bin/golint .

vet:
	go vet ./client/
	go vet ./servers/post/

all: format lint test
