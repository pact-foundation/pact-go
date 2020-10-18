include make/config.mk

TEST?=./...

.DEFAULT_GOAL := ci

ci:: docker deps snyk clean bin test pact goveralls

docker:
	@echo "--- 🛠 Starting docker"
	docker-compose up -d

bin:
	gox -os="darwin" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="windows" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="linux" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="linux" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	@echo "==> Results:"
	ls -hl build/

clean:
	rm -rf build output dist examples/v3/pacts

deps: snyk-install
	@echo "--- 🐿  Fetching build dependencies "
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover
	go get github.com/modocache/gover
	go get github.com/mitchellh/gox

goveralls:
	goveralls -service="travis-ci" -coverprofile=coverage.txt -repotoken $(COVERALLS_TOKEN)

install:
	@if [ ! -d pact/bin ]; then\
		@echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
  fi

installv3:
	@if [ ! -d ./libs/libpact_mock_server_ffi.dylib ]; then\
		@echo "--- 🐿 Installing Rust lib"; \
		make rust; \
  fi

pact: install docker
	@echo "--- 🔨 Running Pact examples"
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample
	go test -v -tags=provider -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample

pactv3: clean
	@echo "--- 🔨 Running Pact examples"
	mkdir -p ./examples/v3/pacts
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/examples/v3/...

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -race -coverprofile=profile.out -covermode=atomic $$d; \
			if [ $$? != 0 ]; then \
				exit 1; \
			fi; \
			if [ -f profile.out ]; then \
					cat profile.out | tail -n +2 >> coverage.txt; \
					rm profile.out; \
			fi; \
	done; \
	go tool cover -func coverage.txt


testrace:
	go test -race $(TEST) $(TESTARGS)

updatedeps:
	go get -d -v -p 2 ./...

snyk-install:
 ifeq (, $(shell which snyk))
	npm i -g snyk
 endif

snyk:
	@if [ "$$TRAVIS_PULL_REQUEST" != "false" ]; then\
		snyk test; \
	fi

rust:
	cd ~/development/public/pact-reference/rust; \
	cargo build; \
	cp ~/development/public/pact-reference/rust/target/debug/libpact_mock_server_ffi.dylib ./libs/libpact_mock_server_ffi.dylib

.PHONY: install bin default dev test pact updatedeps clean release
