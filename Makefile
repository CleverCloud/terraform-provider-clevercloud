TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=CleverCloud
NAME=clevercloud
BINARY=terraform-provider-${NAME}
OS_ARCH=linux_amd64
TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)

ifndef VERSION
	ifeq ($(COMMIT), $(TAG_COMMIT))
		VERSION := $(TAG)
	else
		VERSION := $(TAG)-$(COMMIT)
	endif
endif

default: install

build:
	go build -buildvcs -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build \
		-buildvcs\
		-o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build \
		-buildvcs \
		-o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

lint:
	golangci-lint run

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


.PHONY: docs
docs:
	tfplugindocs generate
# https://github.com/hashicorp/terraform-plugin-docs/pull/446
	sed -i '/subcategory*/d' ./docs/index.md
	find ./docs/resources/* -type f -exec sed -i '/subcategory*/d' {} \;
	tfplugindocs validate
