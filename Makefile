GO ?= go

SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

PYLON = github.com/kayteh/pylon
HASH := $(shell test -d .git && git rev-parse --short HEAD || cat revision~)
BUILD_DATE := $(shell date +%FT%T%z)

# CI Helpers
DOCKER_COMPOSE = docker-compose -f docker/compose-ci.yml


LDFLAGS = -ldflags "-X ${PYLON}/meta.Ref=${HASH} -X ${PYLON}/meta.BuildDate=${BUILD_DATE}"

.PHONY: test test-ci install clean

default: build
build: pylond

pylond: ${SOURCES}
	${GO} build ${LDFLAGS} cmd/pylond/pylond.go

pylon0: ${SOURCES}
	${GO} build ${LDFLAGS} cmd/pylon0/pylon0.go

test:
	${GO} test ${LDFLAGS} -v $(shell ${GO} list ./... | grep -v /vendor/)

test-ci: export KAFKA_ADDR=$(shell ${DOCKER_COMPOSE} port kafka 9092)
test-ci: export REDIS_ADDR=$(shell ${DOCKER_COMPOSE} port redis 6379)
test-ci: test 

install:
	${GO} install ${LDFLAGS} -v cmd/pylond

clean:
	rm pylond pylon0