DIST := dist
EXECUTABLE := douyin
GOFMT ?= gofmt "-s"
GO ?= go
	VERSION ?= v0.2.0

TARGETS ?= linux windows js
ARCHS ?= amd64 386 arm arm64 mips64 mips64le mips mipsle ppc64 ppc64le riscv64 s390x wasm
PACKAGES ?= $(shell $(GO) list ./...)
SOURCES ?= $(shell find . -name "*.go" -type f)
TAGS ?=

LDFLAGS ?= -X 'main.Version=$(VERSION)' -X 'main.DroneBuildNumber=$(DRONE_BUILD_NUMBER)' -X 'main.DroneTag=$(DRONE_TAG)'

#ifneq ($(shell uname), Darwin)
#	EXTLDFLAGS = -extldflags "-static" $(null)
##	EXTLDFLAGS = -extldflags $(null)
#else
#	EXTLDFLAGS =
#endif
EXTLDFLAGS = -extldflags "-static" $(null)

all: build

c: vet misspell-check
#c: vet lint misspell-check sec
 
fmt:
	$(GOFMT) -w $(SOURCES)

vet:
	$(GO) vet $(PACKAGES)

lint:
	@hash revive > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	revive -config .revive.toml ./... || exit 1

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error $(SOURCES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w $(SOURCES)

sec:
	@hash gosec > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/securego/gosec/v2/cmd/gosec; \
	fi
	gosec -exclude=G401,G404,G302,G304,G307,G501 ./...

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(SOURCES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

verify: vet misspell-check fmt-check

test: fmt-check
	@$(GO) test -v -cover -coverprofile coverage.txt ./... && echo "\n==>\033[32m Ok\033[m\n" || exit 1

build:
	$(GO) build -v -tags "$(TAGS)" -ldflags "$(EXTLDFLAGS) -s -w $(LDFLAGS)"  -o api ./apps/api/cmd/api.go

build-cron:
	$(GO) build -v -tags "$(TAGS)" -ldflags "$(EXTLDFLAGS) -s -w $(LDFLAGS)"  -o cron ./apps/cron/cmd/cron.go

release: release-dirs release-build release-copy release-check

release-dirs:
	rm -rf $(DIST); mkdir -p $(DIST)/binaries $(DIST)/release

release-build:
	@which gox > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mitchellh/gox; \
	fi
	gox -os="$(TARGETS)" -arch="$(ARCHS)" -tags="$(TAGS)" -ldflags="-s -w $(LDFLAGS)" -output="$(DIST)/binaries/$(EXECUTABLE)-$(VERSION)-{{.OS}}-{{.Arch}}"

release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)

release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)
