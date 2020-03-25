.ONESHELL:
.DEFAULT_GOAL := run

APPNAME := go-rri-client
PKG := github.com/DENICeG/$(APPNAME)
TEST_PKG_LIST = internal/env pkg/rri

GITCOMMITHASH := $(shell git log --max-count=1 --pretty="format:%h" HEAD)
GITCOMMIT := -X main.gitCommit=$(GITCOMMITHASH)

BUILDTIMEVALUE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
BUILDTIME := -X main.buildTime=$(BUILDTIMEVALUE)

LDFLAGS := '-extldflags "-static" -d -s -w $(GITCOMMIT) $(BUILDTIME)'
LDFLAGS_WINDOWS := '-extldflags "-static" -s -w $(GITCOMMIT) $(BUILDTIME)'

clean:
	@echo "Cleaning up"
	rm -rf bin

dep:
	@go get -v -d ./...
	@go get github.com/stretchr/testify/assert

build-linux: dep
	@echo Building for linux
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(APPNAME) -a -ldflags $(LDFLAGS) .

build-windows: dep
	@echo Building for windows
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows go build -o bin/$(APPNAME).exe -a -ldflags $(LDFLAGS_WINDOWS) .

unit-test:
	@rm -f coverage.out
	@rm -f checktest.out
	@echo 0 > checktest.out
	@echo "mode: set " >> coverage.out

	@for i in $(TEST_PKG_LIST); do \
		CGO_ENABLED=1 go test -p 1 -coverprofile cover.out $(PKG)/$$i || (echo 1 > checktest.out) ;\
		grep -h -v "^mode:" cover.out >> coverage.out; \
		rm -f cover.out; \
	done
	@go tool cover -func=coverage.out
	@exit $$(cat checktest.out)

run:
	go run ./main.go commandline.go
