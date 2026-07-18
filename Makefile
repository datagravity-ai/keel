JOBDATE		?= $(shell date -u +%Y-%m-%dT%H%M%SZ)
GIT_REVISION	= $(shell git rev-parse --short HEAD)
VERSION		?= $(shell git describe --tags --abbrev=0)

LDFLAGS		+= -linkmode external -extldflags -static
LDFLAGS		+= -X github.com/datagravity-ai/keel/version.Version=$(VERSION)
LDFLAGS		+= -X github.com/datagravity-ai/keel/version.Revision=$(GIT_REVISION)
LDFLAGS		+= -X github.com/datagravity-ai/keel/version.BuildDate=$(JOBDATE)

.PHONY: release

fetch-certs:
	curl --remote-name --time-cond cacert.pem https://curl.haxx.se/ca/cacert.pem
	cp cacert.pem ca-certificates.crt

test:
	go install github.com/mfridman/tparse@latest
	go test -json -v `go list ./... | egrep -v /tests` -cover | tparse -all -smallscreen

build:
	@echo "++ Building keel"
	GOOS=linux cd cmd/keel && go build -a -tags netgo -ldflags "$(LDFLAGS) -w -s" -o keel .

install:
	@echo "++ Installing keel"
	GOOS=linux go install -ldflags "$(LDFLAGS)" github.com/datagravity-ai/keel/cmd/keel	

image:
	docker build -t keelhq/keel:alpha -f Dockerfile .

alpha: image
	@echo "++ Pushing keel alpha"
	docker push keelhq/keel:alpha

e2e: install
	cd tests && go test

run:
	go install github.com/datagravity-ai/keel/cmd/keel
	keel --no-incluster --ui-dir ui/dist

lint-ui:
	cd ui && yarn 
	yarn run lint --no-fix && yarn run build

run-ui:
	cd ui && yarn run serve

build-ui:
	docker build -t keelhq/keel:ui -f Dockerfile .
	docker push keelhq/keel:ui

run-debug: install
	DEBUG=true keel --no-incluster
