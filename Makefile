BINARY = rover
HST_OUT=rover_output
VET_REPORT = vet.report
TEST_REPORT = tests.xml
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
VERSION?=?
VCS=github.com
VCS_USERNAME=brianshumate
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags -unsafeptr
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_DIR=${GOPATH}/src/${VCS}/${VCS_USERNAME}/${BINARY}
CURRENT_DIR=$(pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})
BUILD_TAGS?=rover
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"
EXTERNAL_TOOLS=\
	github.com/mitchellh/gox

all: link clean test vet errcheck darwin freebsd linux

clean:
	-rm -f ${TEST_REPORT}
	-rm -f ${VET_REPORT}
	-rm -rf ${HST_OUT}
	-rm -rf darwin-${GOARCH}
	-rm -rf freebsd-${GOARCH}
	-rm -rf linux-${GOARCH}
	-rm -rf bin
	-rm -rf pkg

errcheck:
	if ! hash errcheck 2>/dev/null; then go get -u github.com/kisielk/errcheck; fi ; \
	errcheck ./...

dev:	clean test vet errcheck dev-build

dev-build:
	mkdir -p pkg/$(GOOS)_$(GOARCH)/ bin/
	go install -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)'
	cp $(GOPATH)/bin/rover bin/
	cp $(GOPATH)/bin/rover pkg/$(GOOS)_$(GOARCH)

darwin:
	cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o pkg/darwin-${GOARCH}/${BINARY} . ; \
	cd - >/dev/null

freebsd:
	cd ${BUILD_DIR}; \
	GOOS=freebsd GOARCH=${GOARCH} go build ${LDFLAGS} -o pkg/freebsd-${GOARCH}/${BINARY} . ; \
	cd - >/dev/null

linux:
	cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o pkg/linux-${GOARCH}/${BINARY} . ; \
	cd - >/dev/null

link:
	BUILD_DIR=${BUILD_DIR}; \
	BUILD_DIR_LINK=${BUILD_DIR_LINK}; \
	CURRENT_DIR=${CURRENT_DIR}; \
	if [ "$${BUILD_DIR_LINK}" != "$${CURRENT_DIR}" ]; then \
	    echo "Fixing symlinks for build"; \
	    rm -f $${BUILD_DIR}; \
	    ln -s $${CURRENT_DIR} $${BUILD_DIR}; \
	fi

test:
	if ! hash go2xunit 2>/dev/null; then go get github.com/tebeka/go2xunit; fi
	cd ${BUILD_DIR}; \
	godep go test -v ./... 2>&1 | go2xunit -output ${TEST_REPORT} ; \
	cd - >/dev/null

vet:
	-cd ${BUILD_DIR}; \
	godep go vet ./... > ${VET_REPORT} 2>&1 ; \
	cd - >/dev/null

windows:
	cd ${BUILD_DIR}; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null


fmt:
	cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

.PHONY: link darwin freebsd linux test vet errcheck fmt clean
