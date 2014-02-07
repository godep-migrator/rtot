RTOT_PACKAGE := github.com/modcloth-labs/rtot
TARGETS := \
  $(RTOT_PACKAGE) \
  $(RTOT_PACKAGE)/rtot-server

VERSION_VAR := $(RTOT_PACKAGE).VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)

REV_VAR := $(RTOT_PACKAGE).RevisionString
REPO_REV := $(shell git rev-parse --sq HEAD)

GO ?= go
GODEP ?= godep
GO_TAG_ARGS ?= -tags full
TAGS_VAR := $(RTOT_PACKAGE).BuildTags
GOBUILD_LDFLAGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION) -X $(REV_VAR) $(REPO_REV) -X $(TAGS_VAR) '$(GO_TAG_ARGS)' "

RTOT_HTTPADDR ?= :8457

all: clean test save

test: build fmtpolice testdeps coverage.html
	./mtbb -v

coverage.html: coverage.out
	$(GO) tool cover -html=$^ -o $@

coverage.out:
	$(GO) test -covermode=count -coverprofile=$@ $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(RTOT_PACKAGE)
	$(GO) tool cover -func=$@

testdeps:
	$(GO) test -i $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)

build: deps
	$(GO) install -x $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)

deps: mtbb
	if [ ! -e $${GOPATH%%:*}/src/$(RTOT_PACKAGE) ] ; then \
		mkdir -p $${GOPATH%%:*}/src/github.com/modcloth-labs ; \
		ln -sv $(PWD) $${GOPATH%%:*}/src/$(RTOT_PACKAGE) ; \
	fi
	$(GO) get -x $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)
	$(GODEP) restore

mtbb:
	curl -s -o mtbb https://raw.github.com/modcloth-labs/mtbb/master/lib/mtbb.rb
	chmod +x mtbb

clean:
	rm -vf coverage.html coverage.out
	$(GO) clean -x $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*rtot*' -exec rm -v {} \; ; \
	fi

save:
	$(GODEP) save -copy=false $(RTOT_PACKAGE)

fmtpolice:
	set -e; for f in $(shell git ls-files '*.go'); do gofmt $$f | diff -u $$f - ; done

serve:
	exec $${GOPATH%%:*}/bin/rtot-server -a=$(RTOT_HTTPADDR)

.PHONY: all build clean deps serve test fmtpolice
