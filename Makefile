# https://kodfabrik.com/journal/a-good-makefile-for-go
# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html
# https://github.com/vincentbernat/hellogopher
# https://betterprogramming.pub/my-ultimate-makefile-for-golang-projects-fcc8ca20c9bb
include .env

# Go related variables.
GOBASE=$(shell pwd | sed 's/ /\\ /g')
GOPATH="$(GOBASE)/vendor:$(GOBASE)"
GOBIN=$(GOBASE)/bin

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

PLATFORM?=linux/arm64

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/wunderground-uploader-stderr.txt

## clean: Clean build files.
clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean -mod=mod
	@rm -rf bin
	@rm -rf out

deps:
	@echo "  >  Getting binary dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go mod download

## test: Generate and run all unit tests
test: clean deps
	@echo "  >  Running tests..."
	@mkdir -p out
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go test -v -coverprofile=./out/coverage.out -mod=mod ./...

## coverage: Show unit test coverage report
coverage: test
	@echo "  >  Parsing coverage..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go tool cover -html=./out/coverage.out

compile: clean deps test
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -mod=mod -o $(GOBIN)/wunderground-uploader $(GOBASE)/cmd/wunderground-uploader/main.go

## build: Compile the binary.
build:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## docker-build: Builds the docker image, defaults to linux/amd64 platform can be specified by platform=<platform>.
docker-build:
	@echo "  >  Building docker image..."
	@echo $(CR_PAT) | docker login ghcr.io -u geoff-coppertop --password-stdin
	@docker buildx build \
		--platform $(PLATFORM) \
		-t ghcr.io/geoff-coppertop/wunderground-uploader:latest \
		--push .
	@docker logout ghcr.io

## docker-build-all: Builds all docker images.
docker-build-all:
	@-$(MAKE) -s docker-build PLATFORM=linux/arm64,linux/amd64

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in wunderground-uploader:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
