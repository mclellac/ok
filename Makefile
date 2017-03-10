GOFMT=gofmt -w
DEPS=$(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
PACKAGES := $(shell go list ./...)
UNAME_S	 := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	RESET = $(shell echo -e "\033[0m")
	GREEN = $(shell echo -e "\033[32;01m")
	ERROR = $(shell echo -e "\033[31;01m")
	WARN  = $(shell echo -e "\033[33;01m")
endif
ifeq ($(UNAME_S),Darwin)
	RESET := $(shell echo "\033[0m")
	GREEN := $(shell echo "\033[32;01m")
	ERROR := $(shell echo "\033[31;01m")
	WARN  := $(shell echo "\033[33;01m")
endif

default: build

dep:
	@echo "$(GREEN)>>> Installing dependencies$(RESET)"
	@go get -u -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d

update:
	@echo "$(GREEN)>>> Updating all dependencies$(RESET)"
	@go get -d -u ./...
	@echo $(DEPS) | xargs -n1 go get -d -u

proto:
	@echo "$(GREEN)>>> Generating protocol buffers$(RESET)"
	@if ! which protoc > /dev/null; then \
		echo "$(WARN)Error: protoc not found$(GREEN)" >&2; \
		exit 1; \
	fi
	go get -u -v github.com/golang/protobuf/protoc-gen-go
	# use $$dir as the root for all proto files in the same directory
	for dir in $$(git ls-files '*.proto' | xargs -n1 dirname | uniq); do \
		protoc -I $$dir --go_out=plugins=grpc:$$dir $$dir/*.proto; \
	done

format:
	@echo "$(GREEN)>>> Formatting$(RESET)"
	$(foreach ENTRY,$(PACKAGES),$(GOFMT) $(GOPATH)/src/$(ENTRY);)

build:
	@echo "$(GREEN)>>> Building$(RESET)"
	go build -o ./ok    ./client
	go build -o ./postd ./servers/post

clean:
	@echo "$(GREEN)>>> Cleaning$(RESET)"
	go clean -i -r -x
	rm ./ok && rm ./postd

install:
	@echo "$(GREEN)>>> Installing$(RESET)"
	install ./postd $(GOPATH)/bin
	install ./ok    $(GOPATH)/bin

lint:
	@echo "$(GREEN)>>> Linting$(RESET)"
	$(GOPATH)/bin/golint ./client
	$(GOPATH)/bin/golint ./servers/post

vet:
	go vet ./client/
	go vet ./servers/post/

all: format dep proto vet build
