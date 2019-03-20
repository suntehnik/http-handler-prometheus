SHELL := /bin/bash
GO := GO15VENDOREXPERIMENT=1 GO111MODULE=on go
NAME := http-handler-prometheus
OS := $(shell uname)
MAIN_GO := main.go
ROOT_PACKAGE := $(GIT_PROVIDER)/$(ORG)/$(NAME)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)
PKGS := $(shell go list ./... | grep -v /vendor | grep -v generated | grep -v http-handler-prometheus)
PKGS := $(subst  :,_,$(PKGS))
BUILDFLAGS := ''
CGO_ENABLED = 1


all: build

check: fmt build test

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME)

test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -test.v ./...

full: $(PKGS)

install:
	GOBIN=${GOPATH}/bin $(GO) install -ldflags $(BUILDFLAGS) $(MAIN_GO)

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build release

linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME)

.PHONY: release clean

FGT := $(GOPATH)/bin/fgt
$(FGT):
	go get github.com/GeertJohan/fgt

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

$(PKGS): $(GOLINT) $(FGT)
	@echo "LINTING"
	@$(FGT) $(GOLINT) $(GOPATH)/src/$@/*.go
	@echo "VETTING"
	@go vet -v $@
	@echo "TESTING"
	@go test -v $@

.PHONY: lint http-handler-prometheus
lint: $(PKGS) $(GOLINT) http-handler-prometheus
	@cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
	    test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	done ; exit $$ret

http-handler-prometheus:
	@echo "LINTING"
	@$(FGT) $(GOLINT) ./...
	@echo "VETTING"
	@go vet -v ./...
	@echo "TESTING"
	@go test -v ./...

watch:
	reflex -r "\.go$" -R "vendor.*" make skaffold-run

skaffold-run: build
	skaffold run -p dev  -f delivery/skaffold.yaml --tail

run: build
	LD_LIBRARY_PATH=/usr/local/lib bin/$(NAME) --bind ":8080" \
	-hazel-hosts "138.201.215.226:5701" \
	-hazel-pass "dev-pass" -hazel-user "dev" \
	-kafka-hosts "test.tpc.re:9094" --kafka-topic "gdpr-test"


api-test:
	pushd apitests
	./gradlew test -Dtest=CatsRunner
	popd