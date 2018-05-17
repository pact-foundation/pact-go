TEST?=./...

.DEFAULT_GOAL := ci

ci:: clean bin tools test pact goveralls

bin:
	@sh -c "$(CURDIR)/scripts/build.sh"

clean:
	@sh -c "$(CURDIR)/scripts/clean.sh"

dev:
	@TF_DEV=1 sh -c "$(CURDIR)/scripts/dev.sh"

test:
	"$(CURDIR)/scripts/test.sh"

release:
	"$(CURDIR)/scripts/release.sh"

pact:
	"$(CURDIR)/scripts/pact.sh"

testrace:
	go test -race $(TEST) $(TESTARGS)

tools:
	"$(CURDIR)/scripts/install-cli-tools.sh"

goveralls:
	"$(CURDIR)/scripts/goveralls.sh"

updatedeps:
	go get -d -v -p 2 ./...

.PHONY: bin default dev test pact updatedeps clean release
