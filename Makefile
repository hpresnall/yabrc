default: testandbuild

.PHONY: testandbuild
testandbuild: lint testcoverage build

.PHONY: testcoverage
testcoverage: test coverage

build:
	@./scripts/build-all.sh

# run once after git clone
# may also need to run if there are errors like "export data version x is greater than maximum supported version y"
setup: modules
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.15.0
	go install golang.org/x/tools/gopls@latest

# update dependent packages as needed
modules:
	go mod tidy

.PHONY: test
test:
	@go test -timeout 30s -race -covermode=atomic -coverprofile=coverage.out ./...

coverage:
	@go tool cover -func=coverage.out | grep "total:" | awk '{ print "Total Test Coverage = " $$3 }'
	@go tool cover -html=coverage.out -o=coverage.html

# helper task for development
dofmt:
	@go fmt ./...

lint:
	@$(GOPATH)/bin/golangci-lint run

clean:
	@go clean
	@go clean -cache -testcache
	@rm -rf out coverage.out coverage.html
