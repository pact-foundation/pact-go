include make/config.mk

TEST?=./...

.DEFAULT_GOAL := ci

ci:: clean bin test pact goveralls

install:
	@if [ ! -d pact/bin ]; then\
		echo "--- Installing Pact CLI dependencies";\
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash;\
    fi

bin:
	@sh -c "$(CURDIR)/scripts/build.sh"

clean:
	@sh -c "$(CURDIR)/scripts/clean.sh"

test:
	@echo "--- ✅ Running tests"
	go test -count=1 $(TEST)

release:
	"$(CURDIR)/scripts/release.sh"

pact: install
	@echo "--- 🔨Running Pact examples "
	go test  -tags=consumer -count=1 -v github.com/pact-foundation/pact-go/examples/./... -run TestExample
	go test  -tags=provider -count=1 -v github.com/pact-foundation/pact-go/examples/./... -run TestExample

testrace:
	go test -race $(TEST) $(TESTARGS)

goveralls:
	"$(CURDIR)/scripts/goveralls.sh"

updatedeps:
	go get -d -v -p 2 ./...

.PHONY: install bin default dev test pact updatedeps clean release
