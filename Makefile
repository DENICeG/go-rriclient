.ONESHELL:
.DEFAULT_GOAL := run

APPNAME := go-rriclient
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
	rm -f checktest.out
	rm -f coverage.out
	rm -f cover.out

dep:
	@go get ./...

build-all: dep build-linux build-windows

build-linux:
	@echo Building for linux
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(APPNAME) -a -ldflags $(LDFLAGS) .

build-windows:
	@echo Building for windows
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows go build -o bin/$(APPNAME).exe -a -ldflags $(LDFLAGS_WINDOWS) .

build-all-container:
	@set -e
	@export UID=$$(id -u)
	@if [ -z "$$UID" ]; then
	@	echo "could not detect UID"
	@	exit 1
	@fi
	@export GID=$$(id -g)
	@if [ -z "$$GID" ]; then
	@	echo "could not detect GID"
	@	exit 1
	@fi
	docker run -it --rm --name rri-client-builder --network host --env UID=$$UID --env GID=$$GID --mount type=bind,source="$(CURDIR)",target=/data --workdir /data golang:1.15 sh -c '
		mkdir -p ./bin
		chown -R $$UID:$$GID ./bin
		make build-all
		chown -R $$UID:$$GID ./bin'

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
	@go run . --insecure

install-vulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest

run-vulncheck:
	govulncheck ./...