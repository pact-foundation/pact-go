include make/config.mk

TEST?=./...
.DEFAULT_GOAL := ci
DOCKER_HOST_HTTP?="http://host.docker.internal"
PACT_CLI="docker run --rm -v ${PWD}:${PWD} -e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) -e PACT_BROKER_USERNAME -e PACT_BROKER_PASSWORD pactfoundation/pact-cli"

ci:: docker deps clean bin test pact

# Run the ci target from a developer machine with the environment variables
# set as if it was on Travis CI.
# Use this for quick feedback when playing around with your workflows.
fake_ci:
	@CI=true \
	APP_SHA=`git rev-parse --short HEAD`+`date +%s` \
	APP_BRANCH=`git rev-parse --abbrev-ref HEAD` \
	make ci

# same as above, but just for pact
fake_pact:
	@CI=true \
	APP_SHA=`git rev-parse --short HEAD`+`date +%s` \
	APP_BRANCH=`git rev-parse --abbrev-ref HEAD` \
	make pact

docker:
	@echo "--- 🛠 Starting docker"
	docker-compose up -d

bin:
	go build -o build/pact-go

clean:
	mkdir -p ./examples/pacts
	rm -rf build output dist examples/pacts

deps: download_plugins
	@echo "--- 🐿  Fetching build dependencies "
	cd /tmp; \
	go install github.com/mitchellh/gox@latest; \
	cd -

download_plugins:
	@echo "--- 🐿  Installing plugins"; \
	./scripts/install-cli.sh
	~/.pact/bin/pact-plugin-cli -y install https://github.com/pactflow/pact-protobuf-plugin/releases/tag/v-0.3.13
	~/.pact/bin/pact-plugin-cli -y install https://github.com/pact-foundation/pact-plugins/releases/tag/csv-plugin-0.0.1
	~/.pact/bin/pact-plugin-cli -y install https://github.com/mefellows/pact-matt-plugin/releases/tag/v0.0.9
	~/.pact/bin/pact-plugin-cli -y install https://github.com/austek/pact-avro-plugin/releases/tag/v0.0.3

cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
	fi

install: bin
	echo "--- 🐿 Installing Pact FFI dependencies"
	./build/pact-go	 -l DEBUG install --libDir /tmp

pact: clean install docker
	@echo "--- 🔨 Running Pact examples"
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/...
	make publish
	go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...

publish:
	@echo "-- 📃 Publishing pacts"
	@"${PACT_CLI}" publish ${PWD}/examples/pacts --consumer-app-version ${APP_SHA} --tag ${APP_BRANCH} --tag prod

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -v -race -coverprofile=profile.out -covermode=atomic $$d; \
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

.PHONY: install bin default dev test pact updatedeps clean release

PROTOC ?= $(shell which protoc)

.PHONY: protos
protos:
	@echo "--- 🛠 Compiling Protobufs"
	cd ./examples/grpc/routeguide &&  $(PROTOC) --go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. ./route_guide.proto

.PHONY: grpc-test
grpc-test:
	rm -rf ./examples/pacts
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/grpc
	go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/grpc
