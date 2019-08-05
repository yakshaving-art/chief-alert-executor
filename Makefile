# Makefile for gitlab group manager
# vim: set ft=make ts=8 noet
# Copyright Yakshaving.art
# Licence MIT

# Variables
# UNAME		:= $(shell uname -s)

COMMIT_ID := `git log -1 --format=%H`
COMMIT_DATE := `git log -1 --format=%aI`
VERSION := $${CI_COMMIT_TAG:-SNAPSHOT-$(COMMIT_ID)}

GOOS ?= linux
GOARCH ?= amd64

# this is godly
# https://news.ycombinator.com/item?id=11939200
.PHONY: help
help:	### this screen. Keep it first target to be default
ifeq ($(UNAME), Linux)
	@grep -P '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
else
	@# this is not tested, but prepared in advance for you, Mac drivers
	@awk -F ':.*###' '$$0 ~ FS {printf "%15s%s\n", $$1 ":", $$2}' \
		$(MAKEFILE_LIST) | grep -v '@awk' | sort
endif

# Targets
#
.PHONY: debug
debug:	### debug Makefile itself
	@echo $(UNAME)

.PHONY: check
check:	### sanity checks
	@find . -type f \( -name \*.yml -o -name \*yaml \) \! -path './vendor/*' \
		| xargs -r yq '.' # >/dev/null

.PHONY: lint
lint:	check
lint:	### run all the lints
	gometalinter

.PHONY: test
test:	### run all the unit tests
# test: lint
	@go test -v -coverprofile=coverage.out $$(go list ./... | grep -v '/vendor/') \
		&& go tool cover -func=coverage.out

.PHONY: build
build: ### build the binary applying the correct version from git
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X \
		gitlab.com/yakshaving.art/chief-alert-executor/version.Version=$(VERSION) -X \
		gitlab.com/yakshaving.art/chief-alert-executor/version.Commit=$(COMMIT_ID) -X \
		gitlab.com/yakshaving.art/chief-alert-executor/version.Date=$(COMMIT_DATE)" \
		-o chief-alert-executor-$(GOARCH)