default: testandbuild

.PHONY: testandbuild
testandbuild: lint testcoverage build

.PHONY: testcoverage
testcoverage: test coverage

build:
	@./scripts/build-all.sh

# run once after git clone
setup: modules
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# update dependent packages as needed
modules:
	go mod tidy

test:
	@go test -timeout 30s -race -covermode=atomic -coverprofile=coverage.out ./...

coverage:
	@go tool cover -func=coverage.out | grep "total:" | awk '{ print "Total Test Coverage = " $$3 }'
	@go tool cover -html=coverage.out -o=coverage.html

# helper task for development
dofmt:
	go fmt ./...

lint:
	@$(GOPATH)/bin/golangci-lint run

clean:
	rm -f out coverage.out coverage.html
