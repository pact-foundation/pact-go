package provider

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	native "github.com/pact-foundation/pact-go/v2/internal/native/verifier"
	logging "github.com/pact-foundation/pact-go/v2/log"
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
	ProviderBaseURL string

	// URL of the build to associate with the published verification results.
	BuildURL string

	// Consumer name to filter the pacts to be verified (can be repeated)
	FilterConsumers []string

	// Only validate interactions whose descriptions match this filter
	FilterDescriptions []string

	// Only validate interactions whose provider states match this filter
	FilterStates []string

	// Only validate interactions that do not havve a provider state
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
	ConsumerVersionSelectors []ConsumerVersionSelector

	// Retrieve the latest pacts with this consumer version tag
	Tags []string

	// Tags to apply to the provider application version
	ProviderTags []string

	// ProviderStatesSetupURL is the endpoint to post current provider state
	// to on the Provider API.
	// Deprecated: For backward compatibility ProviderStatesSetupURL is
	// still supported. Use StateHandlers instead.
	ProviderStatesSetupURL string

	// Provider is the name of the Providing service.
	Provider string

	// Username when authenticating to a Pact Broker.
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	BrokerPassword string

	// BrokerToken is required when authenticating using the Bearer token mechanism
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

	args []string
}

// Validate checks that the minimum fields are provided.
func (v *VerifyRequest) validate() error {
	v.args = []string{}

	for _, url := range v.PactURLs {
		v.args = append(v.args, "--url", url)
	}

	for _, url := range v.PactFiles {
		v.args = append(v.args, "--file", url)
	}

	for _, dir := range v.PactDirs {
		v.args = append(v.args, "--dir", dir)
	}

	if len(v.PactURLs) == 0 && len(v.PactFiles) == 0 && len(v.PactDirs) == 0 && v.BrokerURL == "" {
		return fmt.Errorf("one of 'PactURLs', 'PactFiles', 'PactDIRs' or 'BrokerURL' must be specified")
	}

	if len(v.ConsumerVersionSelectors) != 0 {
		for _, selector := range v.ConsumerVersionSelectors {
			if err := selector.Validate(); err != nil {
				return fmt.Errorf("invalid consumer version selector specified: %v", err)
			}
			body, err := json.Marshal(selector)
			if err != nil {
				return fmt.Errorf("invalid consumer version selector specified: %v", err)
			}

			v.args = append(v.args, "--consumer-version-selector", string(body))
		}
	}

	if v.ProviderBaseURL != "" {
		url, err := url.Parse(v.ProviderBaseURL)
		if err != nil {
			return err
		}
		v.args = append(v.args, "--hostname", url.Hostname())

		if url.Port() != "" {
			v.args = append(v.args, "--port", url.Port())
		} else if url.Scheme == "http" {
			v.args = append(v.args, "--port", "80")
		} else if url.Scheme == "https" {
			v.args = append(v.args, "--port", "443")
		}

		if url.Path != "" {
			v.args = append(v.args, "--base-path", url.Path)
		}
	} else {
		return fmt.Errorf("ProviderBaseURL is mandatory")
	}

	if v.BuildURL != "" {
		v.args = append(v.args, "--build-url", v.BuildURL)
	}

	if v.ProviderStatesSetupURL != "" {
		v.args = append(v.args, "--state-change-url", v.ProviderStatesSetupURL)
	}

	if v.BrokerUsername != "" {
		v.args = append(v.args, "--user", v.BrokerUsername)
	}

	if v.BrokerPassword != "" {
		v.args = append(v.args, "--password", v.BrokerPassword)
	}

	if v.BrokerURL != "" && ((v.BrokerUsername == "" && v.BrokerPassword != "") || (v.BrokerUsername != "" && v.BrokerPassword == "")) {
		return errors.New("both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given")
	}

	if v.BrokerURL != "" {
		v.args = append(v.args, "--broker-url", v.BrokerURL)
	}

	if v.BrokerToken != "" {
		v.args = append(v.args, "--token", v.BrokerToken)
	}

	if v.BrokerURL != "" && v.ProviderVersion == "" {
		return errors.New("both 'ProviderVersion' must be supplied if 'BrokerURL' given")
	}

	if v.ProviderVersion != "" {
		v.args = append(v.args, "--provider-version", v.ProviderVersion)
	}

	if v.Provider != "" {
		v.args = append(v.args, "--provider-name", v.Provider)
	}

	if v.PublishVerificationResults {
		v.args = append(v.args, "--publish")
	}

	if v.FilterNoState {
		v.args = append(v.args, "--filter-no-state")
	}

	for _, state := range v.FilterStates {
		v.args = append(v.args, "--filter-state", state)
	}

	for _, state := range v.FilterDescriptions {
		v.args = append(v.args, "--filter-description", state)
	}

	for _, state := range v.FilterConsumers {
		v.args = append(v.args, "--filter-consumer", state)
	}

	if len(v.ProviderTags) > 0 {
		v.args = append(v.args, "--provider-tags", strings.Join(v.ProviderTags, ","))
	}

	v.args = append(v.args, "--loglevel", strings.ToLower(string(logging.LogLevel())))

	if len(v.Tags) > 0 {
		v.args = append(v.args, "--consumer-version-tags", strings.Join(v.Tags, ","))
	}

	if v.EnablePending {
		v.args = append(v.args, "--enable-pending")
	}

	if v.IncludeWIPPactsSince != nil {
		v.args = append(v.args, "--include-wip-pacts-since", v.IncludeWIPPactsSince.Format(time.RFC3339))
	}

	return nil
}

type outputWriter interface {
	Log(args ...interface{})
}

func (v *VerifyRequest) Verify(writer outputWriter) error {
	err := v.validate()
	if err != nil {
		return err
	}

	address := getAddress(v.ProviderBaseURL)
	port := getPort(v.ProviderBaseURL)

	// TODO: parameterise client stuff here
	WaitForPort(port, "tcp", address, 10*time.Second,
		fmt.Sprintf(`Timed out waiting for Provider API to start on port %d - are you sure it's running?`, port))

	service := native.Verifier{}
	native.Init()
	res := service.Verify(v.args)

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
