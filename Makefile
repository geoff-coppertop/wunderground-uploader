# https://kodfabrik.com/journal/a-good-makefile-for-go
# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html
include .env

OWNER?=geoff-coppertop
PROJECTNAME?=$(shell basename "$(PWD)")
REGISTRY?=ghcr.io
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go related variables.
GOBASE=$(shell pwd | sed 's/ /\\ /g')
GOPATH="$(GOBASE)/vendor:$(GOBASE)"
GOBIN=$(GOBASE)/bin
GOFILES=$(shell find . -name '*.go' -not -name '*_test.go')

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## clean: Clean build files. Runs `go clean` internally.
clean: go-clean

## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## docker-build: Builds the docker image.
docker-build:
	@echo "  >  Building docker image..."
	@docker build -t $(REGISTRY)/$(OWNER)/$(PROJECTNAME):latest .

## docker-push: Pushes the docker image.
docker-push: docker-build
	@echo $(CR_PAT) | docker login $(REGISTRY) -u $(OWNER) --password-stdin
	@docker push $(REGISTRY)/$(OWNER)/$(PROJECTNAME):latest
	@docker logout $(REGISTRY)

## exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

go-compile: go-clean go-build

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
	@rm -rf bin

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
