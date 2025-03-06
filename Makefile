include make/config.mk

TEST?=./...
.DEFAULT_GOAL := ci
DOCKER_HOST_HTTP?="http://host.docker.internal"
PACT_CLI="docker run --rm -v ${PWD}:${PWD} -e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) -e PACT_BROKER_USERNAME -e PACT_BROKER_PASSWORD pactfoundation/pact-cli"
PLUGIN_PACT_PROTOBUF_VERSION=0.5.4
PLUGIN_PACT_CSV_VERSION=0.0.6
PLUGIN_PACT_MATT_VERSION=0.1.1
PLUGIN_PACT_AVRO_VERSION=0.0.6

GO_VERSION?=1.23
IMAGE_VARIANT?=debian
ci:: docker deps clean bin test pact
PACT_DOWNLOAD_DIR=/tmp
ifeq ($(OS),Windows_NT)
	PACT_DOWNLOAD_DIR=$$TMP
endif
SKIP_RACE?=false
RACE?=-race
ifeq ($(SKIP_RACE),true)
	RACE=
endif
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
	docker compose up -d

docker_build:
	docker build -f Dockerfile.$(IMAGE_VARIANT) --build-arg GO_VERSION=${GO_VERSION} -t pactfoundation/pact-go-test-$(IMAGE_VARIANT) .

docker_test: docker_build
	docker run \
		-e LOG_LEVEL=INFO \
		-e SKIP_PROVIDER_TESTS=$(SKIP_PROVIDER_TESTS) \
		-e SKIP_RACE=$(SKIP_RACE) \
		--rm \
		-it \
		pactfoundation/pact-go-test-$(IMAGE_VARIANT) \
		/bin/sh -c "make test"
docker_pact: docker_build
	docker run \
		-e LOG_LEVEL=INFO \
		-e SKIP_PROVIDER_TESTS=$(SKIP_PROVIDER_TESTS) \
		-e SKIP_RACE=$(SKIP_RACE) \
		--rm \
		pactfoundation/pact-go-test-$(IMAGE_VARIANT) \
		/bin/sh -c "make pact_local"
docker_test_all: docker_build
	docker run \
		-e LOG_LEVEL=INFO \
		-e SKIP_PROVIDER_TESTS=$(SKIP_PROVIDER_TESTS) \
		-e SKIP_RACE=$(SKIP_RACE) \
		--rm \
		pactfoundation/pact-go-test-$(IMAGE_VARIANT) \
		/bin/sh -c "make test && make pact_local"

bin:
	go build -o build/pact-go

clean:
	mkdir -p ./examples/pacts
	rm -rf build output dist examples/pacts

deps: download_plugins

download_plugins:
	@echo "--- 🐿  Installing plugins"; \
	if [ -z $$SKIP_PLUGINS ]; then\
		if [ ! -f ~/.pact/bin/pact-plugin-cli ]; then \
			./scripts/install-cli.sh; \
		else \
			echo "--- 🐿  Pact CLI already installed"; \
		fi; \
		if [ ! -f ~/.pact/plugins/protobuf-$(PLUGIN_PACT_PROTOBUF_VERSION)/pact-protobuf-plugin ]; then \
			~/.pact/bin/pact-plugin-cli -y install https://github.com/pactflow/pact-protobuf-plugin/releases/tag/v-$(PLUGIN_PACT_PROTOBUF_VERSION); \
		else \
			echo "--- 🐿  Pact protobuf-$(PLUGIN_PACT_PROTOBUF_VERSION) already installed"; \
		fi; \
		if [ ! -f ~/.pact/plugins/csv-$(PLUGIN_PACT_CSV_VERSION)/pact-csv-plugin ]; then \
			~/.pact/bin/pact-plugin-cli -y install https://github.com/pact-foundation/pact-plugins/releases/tag/csv-plugin-$(PLUGIN_PACT_CSV_VERSION); \
		else \
			echo "--- 🐿  Pact csv-$(PLUGIN_PACT_CSV_VERSION) already installed"; \
		fi; \
		if [ ! -f ~/.pact/plugins/matt-$(PLUGIN_PACT_MATT_VERSION)/matt ]; then \
			~/.pact/bin/pact-plugin-cli -y install https://github.com/mefellows/pact-matt-plugin/releases/tag/v$(PLUGIN_PACT_MATT_VERSION); \
		else \
			echo "--- 🐿  Pact matt-$(PLUGIN_PACT_MATT_VERSION) already installed"; \
		fi; \
		if [ ! -f ~/.pact/plugins/avro-$(PLUGIN_PACT_AVRO_VERSION)/bin/pact-avro-plugin ]; then \
			~/.pact/bin/pact-plugin-cli -y install https://github.com/austek/pact-avro-plugin/releases/tag/v$(PLUGIN_PACT_AVRO_VERSION); \
		else \
			echo "--- 🐿  Pact avro-$(PLUGIN_PACT_AVRO_VERSION) already installed"; \
		fi; \
	fi

cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
	fi

install: bin
	echo "--- 🐿 Installing Pact FFI dependencies"
	./build/pact-go -l DEBUG install --libDir $(PACT_DOWNLOAD_DIR)

pact: clean install docker
	@echo "--- 🔨 Running Pact examples"
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/...
	make publish
	go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...
pact_local: clean download_plugins install 
	@echo "--- 🔨 Running Pact examples"
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/...
	if [ "$(SKIP_PROVIDER_TESTS)" != "true" ]; then \
		SKIP_PUBLISH=true go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...; \
	fi

publish:
	@echo "-- 📃 Publishing pacts"
	@"${PACT_CLI}" publish ${PWD}/examples/pacts --consumer-app-version ${APP_SHA} --tag ${APP_BRANCH} --tag prod

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

ifeq ($(SKIP_PROVIDER_TESTS),true)
	PROVIDER_TEST_TAGS=
else
	PROVIDER_TEST_TAGS=-tags=provider
endif

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -v $(RACE) -coverprofile=profile.out $(PROVIDER_TEST_TAGS) -covermode=atomic $$d; \
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
	go test $(RACE) $(TEST) $(TESTARGS)

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
