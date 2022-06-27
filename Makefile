include make/config.mk

TEST?=./...
.DEFAULT_GOAL := ci
DOCKER_HOST_HTTP?="http://host.docker.internal"
PACT_CLI="docker run --rm -v ${PWD}:${PWD} -e PACT_BROKER_BASE_URL=$(DOCKER_HOST_HTTP) -e PACT_BROKER_USERNAME -e PACT_BROKER_PASSWORD pactfoundation/pact-cli"

ci:: docker deps clean bin test pact 

docker:
	@echo "--- 🛠 Starting docker"
	docker-compose up -d

bin:
	gox -os="darwin" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="darwin" -arch="arm64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="windows" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="linux" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	gox -os="linux" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
	@echo "==> Results:"
	ls -hl build/

clean:
	rm -rf build output dist

deps:
	@echo "--- 🐿  Fetching build dependencies "
	cd /tmp; \
	go install github.com/mitchellh/gox@latest; \
	cd -

install:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash -x; \
  	fi

publish_pacts: 
	@echo "\n========== STAGE: publish pacts ==========\n"
	@"${PACT_CLI}" publish ${PWD}/examples/pacts --consumer-app-version ${GIT_COMMIT} --tag ${GIT_BRANCH} --tag dev --tag prod

pact_local:
	GIT_COMMIT=`git rev-parse --short HEAD`+`date +%s` \
	GIT_BRANCH=`git rev-parse --abbrev-ref HEAD` \
	make pact

pact: install docker
	@echo "--- 🔨 Running Pact examples"
	go test -tags=consumer -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample
	make publish_pacts
	go test -tags=provider -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps install
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -race -coverprofile=profile.cov -covermode=atomic $$d; \
			if [ $$? != 0 ]; then \
				exit 1; \
			fi; \
			if [ -f profile.cov ]; then \
					cat profile.cov | tail -n +2 >> coverage.txt; \
					rm profile.cov; \
			fi; \
	done; \
	go tool cover -func coverage.txt

testrace:
	go test -race $(TEST) $(TESTARGS)

updatedeps:
	go get -d -v -p 2 ./...

.PHONY: install bin default dev test pact updatedeps clean release
