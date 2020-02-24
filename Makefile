include make/config.mk

TEST?=./...

.DEFAULT_GOAL := ci

ci:: docker deps clean bin test pact goveralls

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
	rm -rf build output dist

deps:
	@echo "--- 🐿  Fetching build dependencies "
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover
	go get github.com/modocache/gover
	go get github.com/mitchellh/gox

goveralls:
	goveralls -service="travis-ci" -coverprofile=coverage.txt -repotoken $(COVERALLS_TOKEN)

# uname_output=$(uname); \
# case $uname_output in; \
# 	'Linux'); \
# 		linux_uname_output=$(uname -m); \
# 		case $linux_uname_output in; \
# 			'x86_64'); \
# 				os='linux-x86_64'; \
# 				;;; \
# 			'i686'); \
# 				os='linux-x86'; \
# 				;;; \
# 			*); \
# 				echo "Sorry, you'll need to install the pact-ruby-standalone manually."; \
# 				exit 1; \
# 				;;; \
# 		esac; \
# 		;;; \
# 	'Darwin'); \
# 		os='osx'; \
# 		;;; \
# 	*); \
# 	echo "Sorry, you'll need to install the pact-ruby-standalone manually."; \
# 	exit 1; \
# 		;;; \
# esac; \

install:
	@if [ ! -d pact/bin ]; then\
		echo "--- 🐿 Installing Pact CLI dependencies"; \
		os=linux-x86_64; \
		response=$(curl -s -v https://github.com/pact-foundation/pact-ruby-standalone/releases/latest 2>&1); \
		tag=$(echo "$response" | grep -o "Location: .*" | sed -e 's/[[:space:]]*$//' | grep -o "Location: .*" | grep -o '[^/]*$'); \
		version=${tag#v}; \
		curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/${tag}/pact-${version}-${os}.tar.gz; \
		tar xzf pact-${version}-${os}.tar.gz; \
		rm pact-${version}-${os}.tar.gz; \
  fi

pact: install docker
	@echo "--- 🔨 Running Pact examples"
	go test -tags=consumer -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample
	go test -tags=provider -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -race -coverprofile=profile.out -covermode=atomic $$d; \
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