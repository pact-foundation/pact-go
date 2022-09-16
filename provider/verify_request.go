package provider

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pact-foundation/pact-go/v2/command"
	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
)

// Hook functions are used to tap into the lifecycle of a Consumer or Provider test
type Hook func() error

// VerifyRequest contains the verification params.
type VerifyRequest struct {
	// URL to hit during provider verification.
	// NOTE: if specified alongside PactURLs, PactFiles or PactDirs it will run the verification once for
	// each dynamic pact (Broker) discovered and user specified (URL) pact.
	// It may also be specified by the PACT_BROKER_BASE_URL environment variable
	ProviderBaseURL string

	// URL of the build to associate with the published verification results.
	BuildURL string

	// Consumer name to filter the pacts to be verified (can be repeated)
	FilterConsumers []string

	// Only validate interactions whose descriptions match this filter
	// It may also be specified by the PACT_DESCRIPTION environment variable
	FilterDescription string

	// Only validate interactions whose provider states match this filter
	// It may also be specified by the PACT_PROVIDER_STATE environment variable
	FilterState string

	// Only validate interactions that do not havve a provider state
	// It may also be specified by setting the environment variable PACT_PROVIDER_NO_STATE to "true"
	FilterNoState bool

	// HTTP paths to Pact files.
	// NOTE: if specified alongside BrokerURL, PactFiles or PactDirs it will run the verification once for
	// each dynamic pact (Broker) discovered and user specified (URL) pact.
	PactURLs []string

	// Local paths to Pact files.
	// NOTE: if specified alongside PactURLs, BrokerURL or PactDirs it will run the verification once for
	// each dynamic pact (Broker) discovered and user specified (URL) pact.
	PactFiles []string

	// Local path to a directory containing Pact files.
	// NOTE: if specified alongside PactURLs, PactFiles or BrokerURL it will run the verification once for
	// each dynamic pact (Broker) discovered and user specified (URL) pact.
	PactDirs []string

	// Pact Broker URL for broker-based verification
	BrokerURL string

	// Selectors are the way we specify which pacticipants and
	// versions we want to use when configuring verifications
	// See https://docs.pact.io/selectors for more
	ConsumerVersionSelectors []Selector

	// Retrieve the latest pacts with this consumer version tag
	Tags []string

	// Tags to apply to the provider application version
	ProviderTags []string

	// Branch to apply to the provider application version
	ProviderBranch string

	// ProviderStatesSetupURL is the endpoint to post current provider state
	// to on the Provider API.
	// Deprecated: For backward compatibility ProviderStatesSetupURL is
	// still supported. Use StateHandlers instead.
	ProviderStatesSetupURL string

	// Provider is the name of the Providing service.
	Provider string

	// Username when authenticating to a Pact Broker.
	// It may also be specified by the PACT_BROKER_USERNAME environment variable
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	// It may also be specified by the PACT_BROKER_PASSWORD environment variable
	BrokerPassword string

	// BrokerToken is required when authenticating using the Bearer token mechanism
	// It may also be specified by the PACT_BROKER_TOKEN environment variable
	BrokerToken string

	// FailIfNoPactsFound configures the framework to return an error
	// if no pacts were found when looking up from a broker
	FailIfNoPactsFound bool

	// PublishVerificationResults to the Pact Broker.
	PublishVerificationResults bool

	// ProviderVersion is the semantical version of the Provider API.
	ProviderVersion string

	// CustomProviderHeaders are headers to add during pact verification `requests`.
	// eg 'Authorization: Basic cGFjdDpwYWN0'.
	//
	// NOTE: Use this feature very carefully, as anything in here is not captured
	// in the contract (e.g. time-bound tokens)
	//
	// NOTE: This should be used very carefully and deliberately, as anything you do here
	// runs the risk of changing the contract and breaking the real system.
	// CustomProviderHeaders []string

	// StateHandlers contain a mapped list of message states to functions
	// that are used to setup a given provider state prior to the message
	// verification step.
	StateHandlers models.StateHandlers

	// BeforeEach allows you to configure your provider prior to the individual test execution
	// e.g. setup temporary tokens, prepare data
	BeforeEach Hook

	// AfterEach allows you to configure your provider prior to the test execution
	// e.g. reset the database state
	AfterEach Hook

	// RequestFilter is a piece of middleware that will intercept requests/responses
	// from the provider in order to modify it. This is useful in situations where
	// you need to override a value due to time sensitivity - such as a OAuth Bearer
	// token.
	// NOTE: This should be used very carefully and deliberately, as anything you do here
	// runs the risk of changing the contract and breaking the real system.
	RequestFilter proxy.Middleware

	// Custom TLS Configuration to use when making the requests to/from
	// the Provider API. Useful for setting custom certificates, MASSL etc.
	CustomTLSConfig *tls.Config

	// Allow pending pacts to be included in verification (see pact.io/pending)
	EnablePending bool

	// Pull in new WIP pacts from _any_ tag (see pact.io/wip)
	IncludeWIPPactsSince *time.Time

	handle *native.Verifier
}

// Validate checks that the minimum fields are provided.
func (v *VerifyRequest) validate() error {

	var scheme, hostname, path string
	var port int

	if v.ProviderBaseURL != "" {
		url, err := url.Parse(v.ProviderBaseURL)
		if err != nil {
			return err
		}
		hostname = url.Hostname()

		if url.Port() != "" {
			port, _ = strconv.Atoi(url.Port())
		} else if url.Scheme == "http" {
			port = 80
		} else if url.Scheme == "https" {
			port = 443
		}

		if url.Path != "" {
			path = url.Path
		}
	} else {
		return fmt.Errorf("ProviderBaseURL is mandatory")
	}

	v.handle.SetProviderInfo(v.Provider, scheme, hostname, uint16(port), path)

	filterDescription := valueOrFromEnvironment(v.FilterDescription, "PACT_DESCRIPTION")
	filterState := valueOrFromEnvironment(v.FilterState, "PACT_PROVIDER_STATE")
	filterNoState := valueOrFromEnvironment(fmt.Sprintf("%t", v.FilterNoState), "PACT_PROVIDER_NO_STATE") == "true"

	if filterDescription != "" || filterState != "" || os.Getenv("PACT_PROVIDER_NO_STATE") != "" {
		v.handle.SetFilterInfo(filterDescription, filterState, filterNoState)
	}

	if v.ProviderStatesSetupURL != "" {
		v.handle.SetProviderState(v.ProviderStatesSetupURL, true, true)
	}

	// TODO: support these
	// SetVerificationOptions: 4,

	if v.PublishVerificationResults && v.ProviderVersion != "" {
		v.handle.SetPublishOptions(v.ProviderVersion, v.BuildURL, v.ProviderTags, v.ProviderBranch)
	}

	if len(v.FilterConsumers) > 0 {
		v.handle.SetConsumerFilters(v.FilterConsumers)
	}

	// TODO:
	// AddCustomHeader: 7,

	for _, url := range v.PactURLs {
		v.handle.AddURLSource(url, valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME"), valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD"), valueOrFromEnvironment(v.BrokerToken, "PACT_BROKER_TOKEN"))
	}

	for _, file := range v.PactFiles {
		v.handle.AddFileSource(file)
	}

	for _, dir := range v.PactDirs {
		v.handle.AddDirectorySource(dir)
	}

	if len(v.PactURLs) == 0 && len(v.PactFiles) == 0 && len(v.PactDirs) == 0 && v.BrokerURL == "" {
		return fmt.Errorf("one of 'PactURLs', 'PactFiles', 'PactDIRs' or 'BrokerURL' must be specified")
	}

	selectors := make([]string, len(v.ConsumerVersionSelectors))

	if len(v.ConsumerVersionSelectors) != 0 {
		for i, selector := range v.ConsumerVersionSelectors {
			body, err := json.Marshal(selector)
			if err != nil {
				return fmt.Errorf("invalid consumer version selector specified: %v", err)
			}

			selectors[i] = string(body)
		}
	}

	if v.BrokerURL != "" && (v.ProviderVersion == "" || v.Provider == "") {
		return errors.New("'ProviderVersion', and 'Provider' must be supplied if 'BrokerURL' given")
	}

	if v.BrokerURL != "" && ((valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME") == "" && valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD") != "") || (valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME") != "" && valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD") == "")) {
		return errors.New("both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given")
	}

	includeWIPPactsSince := ""
	if v.IncludeWIPPactsSince != nil {
		includeWIPPactsSince = v.IncludeWIPPactsSince.Format(time.RFC3339)
	}

	if v.BrokerURL != "" && v.Provider != "" {
		v.handle.BrokerSourceWithSelectors(v.BrokerURL, valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME"), valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD"), valueOrFromEnvironment(v.BrokerToken, "PACT_BROKER_TOKEN"), v.EnablePending, includeWIPPactsSince, v.ProviderTags, v.ProviderBranch, selectors, v.Tags)
	}

	return nil
}

func valueOrFromEnvironment(value string, envKey string) string {
	if value != "" {
		return value
	}

	return os.Getenv(envKey)
}

type outputWriter interface {
	Log(args ...interface{})
}

func (v *VerifyRequest) Verify(writer outputWriter) error {
	v.handle = native.NewVerifier("pact-go", command.Version)

	err := v.validate()
	if err != nil {
		return err
	}

	address := getAddress(v.ProviderBaseURL)
	port := getPort(v.ProviderBaseURL)

	// TODO: parameterise client stuff here
	WaitForPort(port, "tcp", address, 10*time.Second,
		fmt.Sprintf(`Timed out waiting for Provider API to start on port %d - are you sure it's running?`, port))

	res := v.handle.Execute()
	defer v.handle.Shutdown()

	return res
}

// Get a port given a URL
func getPort(rawURL string) int {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		splitHost := strings.Split(parsedURL.Host, ":")
		if len(splitHost) == 2 {
			port, err := strconv.Atoi(splitHost[1])
			if err == nil {
				return port
			}
		}
		if parsedURL.Scheme == "https" {
			return 443
		}
		return 80
	}

	return -1
}

// Get the address given a URL
func getAddress(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	splitHost := strings.Split(parsedURL.Host, ":")
	return splitHost[0]
}
