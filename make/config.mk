export PATH := $(PWD)/pact/bin:$(PATH)
export PATH
export PACT_BROKER_BASE_URL?=http://127.0.0.1

ifndef PACT_BROKER_TOKEN
	export PACT_BROKER_USERNAME?=pact_workshop
	export PACT_BROKER_PASSWORD?=pact_workshop
endif