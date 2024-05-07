include make/config.mk

TEST?=./...
.DEFAULT_GOAL := ci
DOCKER_HOST_HTTP?="http://host.docker.internal"
ifeq ($(OS),Windows_NT)
	EXT=.exe
endif

ci:: docker deps clean bin test pact
ci_unit:: deps clean bin test
ci_examples:: docker pact
ci_hosted_examples:: run_ci_hosted_examples

run_ci_hosted_examples:
	DOCKER_HOST_HTTP=$(PACT_BROKER_BASE_URL) make pact
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
PLUGIN_PACT_PROTOBUF_VERSION=0.3.14
PLUGIN_PACT_CSV_VERSION=0.0.5
PLUGIN_PACT_MATT_VERSION=0.1.1
PLUGIN_PACT_AVRO_VERSION=0.0.3

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
		if [ -z $$SKIP_PLUGIN_AVRO ]; then\
			if [ ! -f ~/.pact/plugins/avro-$(PLUGIN_PACT_AVRO_VERSION)/bin/pact-avro-plugin ]; then \
				~/.pact/bin/pact-plugin-cli -y install https://github.com/austek/pact-avro-plugin/releases/tag/v$(PLUGIN_PACT_AVRO_VERSION); \
			else \
				echo "--- 🐿  Pact avro-$(PLUGIN_PACT_AVRO_VERSION) already installed"; \
			fi; \
		fi; \
	fi


cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
	fi

install: bin
	echo "--- 🐿 Installing Pact FFI dependencies"
	./build/pact-go -l DEBUG install --libDir /tmp

pact: clean install
	@echo "--- 🔨 Running Pact examples"
	go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/...
	make publish
	go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...

publish:
	@echo "-- 📃 Publishing pacts"
	@${PACT_BROKER_COMMAND} publish ${PWD}/examples/pacts --consumer-app-version ${APP_SHA} --tag ${APP_BRANCH} --tag prod --branch ${APP_BRANCH}

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

RACE?='-race'

ifdef SKIP_RACE
	RACE=
endif

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -v $(RACE) -coverprofile=profile.out -covermode=atomic $$d; \
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

docker_build:
	docker build -f Dockerfile --build-arg VERSION=1.21 -t pactfoundation/pact-go-test .
docker_run_test:
	docker run \
		-e LOG_LEVEL=info \
		-e APP_SHA=foo \
		--rm \
		-it \
		pactfoundation/pact-go-test \
		/bin/sh -c "make test"
docker_run_examples:
	docker run \
		-e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) \
		-e PACT_BROKER_TOKEN \
		-e PACT_BROKER_USERNAME \
		-e PACT_BROKER_PASSWORD \
		-e LOG_LEVEL=info \
		-e APP_SHA=foo \
		--rm \
		-it \
		pactfoundation/pact-go-test \
		/bin/sh -c "make download_plugins && make install-pact-ruby-standalone && PACT_TOOL=standalone make pact"

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

## =====================
## Multi-platform detection and support
## Pact CLI install/uninstall tasks
## =====================
SHELL := /bin/bash
PACT_TOOL?=docker
PACT_CLI_DOCKER_VERSION?=latest
PACT_CLI_VERSION?=latest
PACT_CLI_STANDALONE_VERSION?=2.4.1
PACT_CLI_DOCKER_RUN_COMMAND?=docker run --rm -v /${PWD}:/${PWD} -w ${PWD} -e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) -e PACT_BROKER_TOKEN -e PACT_BROKER_USERNAME -e PACT_BROKER_PASSWORD pactfoundation/pact-cli:${PACT_CLI_DOCKER_VERSION}
PACT_BROKER_COMMAND=pact-broker
PACTFLOW_CLI_COMMAND=pactflow

ifeq '$(findstring ;,$(PATH))' ';'
	detected_OS := Windows
else
	detected_OS := $(shell uname -sm 2>/dev/null || echo Unknown)
	detected_OS := $(patsubst CYGWIN%,Cygwin,$(detected_OS))
	detected_OS := $(patsubst MSYS%,MSYS,$(detected_OS))
	detected_OS := $(patsubst MINGW%,MSYS,$(detected_OS))
endif

ifeq ($(PACT_TOOL),ruby_standalone)
# add path to standalone, and add bat if windows
	ifneq ($(filter $(detected_OS),Windows MSYS),)
		PACT_BROKER_COMMAND:="./pact/bin/${PACT_BROKER_COMMAND}.bat"
		PACTFLOW_CLI_COMMAND:="./pact/bin/${PACTFLOW_CLI_COMMAND}.bat"
	else
		PACT_BROKER_COMMAND:="./pact/bin/${PACT_BROKER_COMMAND}"
		PACTFLOW_CLI_COMMAND:="./pact/bin/${PACTFLOW_CLI_COMMAND}"
	endif
endif

ifeq ($(PACT_TOOL),docker)
# add docker run command path
	PACT_BROKER_COMMAND:=${PACT_CLI_DOCKER_RUN_COMMAND} ${PACT_BROKER_COMMAND}
	PACTFLOW_CLI_COMMAND:=${PACT_CLI_DOCKER_RUN_COMMAND} ${PACTFLOW_CLI_COMMAND}
endif


install-pact-ruby-cli:
	case "${PACT_CLI_VERSION}" in \
	latest) gem install pact_broker-client;; \
	"") gem install pact_broker-client;; \
		*) gem install pact_broker-client -v ${PACT_CLI_VERSION} ;; \
	esac

uninstall-pact-ruby-cli:
	gem uninstall -aIx pact_broker-client

install-pact-ruby-standalone:
	case "${detected_OS}" in \
	Windows|MSYS) curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_CLI_STANDALONE_VERSION}/pact-${PACT_CLI_STANDALONE_VERSION}-windows-x86_64.zip && \
		unzip pact-${PACT_CLI_STANDALONE_VERSION}-windows-x86_64.zip && \
		rm pact-${PACT_CLI_STANDALONE_VERSION}-windows-x86_64.zip && \
		./pact/bin/pact-broker.bat help;; \
	"Darwin arm64") curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_CLI_STANDALONE_VERSION}/pact-${PACT_CLI_STANDALONE_VERSION}-osx-arm64.tar.gz && \
		tar xzf pact-${PACT_CLI_STANDALONE_VERSION}-osx-arm64.tar.gz && \
		rm pact-${PACT_CLI_STANDALONE_VERSION}-osx-arm64.tar.gz && \
		./pact/bin/pact-broker help;; \
	"Darwin x86_64") curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_CLI_STANDALONE_VERSION}/pact-${PACT_CLI_STANDALONE_VERSION}-osx-x86_64.tar.gz && \
		tar xzf pact-${PACT_CLI_STANDALONE_VERSION}-osx-x86_64.tar.gz && \
		rm pact-${PACT_CLI_STANDALONE_VERSION}-osx-x86_64.tar.gz && \
		./pact/bin/pact-broker help;; \
	"Linux aarch64") curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_CLI_STANDALONE_VERSION}/pact-${PACT_CLI_STANDALONE_VERSION}-linux-arm64.tar.gz && \
		tar xzf pact-${PACT_CLI_STANDALONE_VERSION}-linux-arm64.tar.gz && \
		rm pact-${PACT_CLI_STANDALONE_VERSION}-linux-arm64.tar.gz && \
		./pact/bin/pact-broker help;; \
	"Linux x86_64") curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v${PACT_CLI_STANDALONE_VERSION}/pact-${PACT_CLI_STANDALONE_VERSION}-linux-x86_64.tar.gz && \
		tar xzf pact-${PACT_CLI_STANDALONE_VERSION}-linux-x86_64.tar.gz && \
		rm pact-${PACT_CLI_STANDALONE_VERSION}-linux-x86_64.tar.gz && \
		./pact/bin/pact-broker help;; \
	esac