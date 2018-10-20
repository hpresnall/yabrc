GOPACKAGES=$(shell go list ./... | grep -v /vendor)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"

# must run 'make deps' at least once before the default will succeed!`
default: testandbuild

.PHONY: testandbuild
testandbuild: lintall testcoverage build

.PHONY: testcoverage
testcoverage: test coverage

.PHONY: lintall
lintall: fmt vet lint

build:
	@./scripts/build-all.sh

setup:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep ensure
	go get -u github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null 2>&1

# update dependant packages as needed
deps:
	dep ensure 

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

fmt:
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Please run gofmt -lw on your code.' && exit 1; fi

vet:
	@$(GOPATH)/bin/gometalinter --disable-all --enable=vet --enable=vetshadow --enable=ineffassign --enable=goconst --tests --vendor -e docs.go ./...

lint:
	@$(GOPATH)/bin/golint -set_exit_status=true $(GOPACKAGES)

clean:
	rm -f coverage.out coverage_percent.out coverage.html
	rm -rf out vendor
