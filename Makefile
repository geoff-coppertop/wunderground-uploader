# https://kodfabrik.com/journal/a-good-makefile-for-go
# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html
# https://github.com/vincentbernat/hellogopher
# https://betterprogramming.pub/my-ultimate-makefile-for-golang-projects-fcc8ca20c9bb
include .env

# Go related variables.
GOBASE=$(shell pwd | sed 's/ /\\ /g')
GOPATH="$(GOBASE)/vendor:$(GOBASE)"
GOBIN=$(GOBASE)/bin
GOFILES := $(shell find . -path ./vendor -prune -o -name "*.go" -not -name '*_test.go' -print)

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/wunderground-uploader-stderr.txt

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## clean: Clean build files.
.PHONY: clean
clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
	@rm -rf bin

.PHONY: deps
deps:
	@echo "  >  Getting binary dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go mod download

compile: clean deps
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -mod=mod -o $(GOBIN)/wunderground-uploader $(GOFILES)

## build: Compile the binary.
build:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## docker-build: Builds the docker image.
docker-build:
	@echo "  >  Building docker image..."
	@docker build -t ghcr.io/geoff-coppertop/wunderground-uploader:latest .

## docker-push: Pushes the docker image.
docker-push: docker-build
	@echo $(CR_PAT) | docker login ghcr.io -u geoff-coppertop --password-stdin
	@docker push ghcr.io/geoff-coppertop/wunderground-uploader:latest
	@docker logout ghcr.io

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in wunderground-uploader:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
