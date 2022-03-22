GOPACKAGES=$(shell go list ./...)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"

default: testandbuild

.PHONY: testandbuild
testandbuild: lint testcoverage build

.PHONY: testcoverage
testcoverage: test coverage

build:
	@./scripts/build-all.sh

setup: modules
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# update dependent packages as needed
modules:
	go mod tidy

test:
	@echo 'mode: atomic' > coverage.out
	@$(foreach pkg,$(GOPACKAGES),\
		go test -timeout 30s -race -covermode=atomic -coverprofile=coverage.tmp $(pkg);\
		tail -n +2 coverage.tmp >> coverage.out;)
	@rm -f coverage.tmp

coverage:
	@go tool cover -func=coverage.out | grep "total:" | awk '{ print "Total Test Coverage = " $$3 }'
	@go tool cover -html=coverage.out -o=coverage.html

# helper task for development
dofmt:
	gofmt -w -l ${GOFILES}

lint:
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Please run make dofmt.' && exit 1; fi
	@$(GOPATH)/bin/golangci-lint run

clean:
	rm -f coverage.out coverage.html
