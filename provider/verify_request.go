package provider

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	logging "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
)

// defaultRequestTimeout is how long the verifier waits for each individual
// request to the provider when VerifyRequest.RequestTimeout is left unset.
const defaultRequestTimeout = 10 * time.Second

// Hook functions are used to tap into the lifecycle of a Consumer or Provider test.
type Hook func() error

// VerifyRequest contains the verification params.
type VerifyRequest struct {
	// Default URL to hit during provider verification.
	//
	// If your URL has a base path, make sure the URL includes the schema.
	// Otherwise, the [net/url] package will consider it as opaque, and the base path will be lost when the URL is parsed.
	ProviderBaseURL string

	// Specify one or more additional transports to communicate to the given provider
	// Providers may support multiple modes - e.g. HTTP, gRPC etc.
	Transports []Transport

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
	//
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

	// SoftFail, when true, renders a verification mismatch (one or more
	// consumer expectations not satisfied by the provider, returned by the
	// verifier as ErrVerifierFailed) as a SKIPped subtest rather than a
	// failed one. Use when the caller has a downstream gate — e.g. a Pact
	// Broker can-i-merge / can-i-deploy check — that owns the authoritative
	// compatibility decision, and local CI failure on the mismatch would
	// duplicate or contradict that gate.
	//
	// Infrastructure errors (ErrVerifierFailedToRun, panics, broker auth
	// failures, etc.) always fail the subtest regardless of this flag —
	// they signal that the verifier could not produce a result for the
	// downstream gate to act on.
	//
	// Default false preserves existing behavior: any verifier error fails
	// the subtest.
	SoftFail bool

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

	// MessageHandlers contains a mapped list of message handlers for a provider
	// that will be rable to produce the correct message format for a given
	// consumer interaction
	MessageHandlers message.Handlers

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

	// Sets the max time the framework will wait to issue requests to your provider API
	// as well as timeout for provider state actions
	RequestTimeout time.Duration

	// Disable SSL verification for HTTP requests
	DisableSSLVerification bool

	// If true, will disable colored output in console.
	DisableColoredOutput bool
}

// add in the PACT_URL env variable to support suggested webhook provider verification
// see https://docs.pact.io/pact_broker/webhooks/template_library#bitbucket---trigger-pipeline-run
// a generalized feature request added here https://github.com/pact-foundation/pact-reference/issues/250
func addPactUrlsFromEnvironment(v *VerifyRequest) {
	if pactURL := os.Getenv("PACT_URL"); pactURL != "" {
		v.PactURLs = append(v.PactURLs, pactURL)
	}
}

func valueOrFromEnvironment(value string, envKey string) string {
	if value != "" {
		return value
	}

	return os.Getenv(envKey)
}

type outputWriter interface {
	Log(args ...any)
}

// Verify configures handle from v's transports and provider state setup
// URL, then runs the verification. It shuts handle down before
// returning, so handle must not be reused afterwards.
func (v *VerifyRequest) Verify(handle *native.Verifier, writer outputWriter) error {
	for _, transport := range v.Transports {
		log.Println("[DEBUG] adding transport to verification", transport)
		handle.AddTransport(transport.Protocol, transport.Port, transport.Path, transport.Scheme)
	}

	if v.ProviderStatesSetupURL != "" {
		handle.SetProviderState(v.ProviderStatesSetupURL, true, true)
	}

	defer handle.Shutdown()
	//nolint:wrapcheck // Verifier.Execute reports the verification outcome through
	// internal/native's sentinels (ErrVerifierFailed, ErrVerifierFailedToRun),
	// selected by the pact_ffi return code; the mismatch detail goes to the output
	// writer, so the sentinel's own wording is all the user's test failure prints.
	return handle.Execute()
}

// Validate checks that the minimum fields are provided.
func (v *VerifyRequest) validate(handle *native.Verifier) error {
	if v.ProviderBaseURL == "" {
		logging.PactCrash(errors.New("ProviderBaseURL is a required field"))
	} else {
		url, err := url.Parse(v.ProviderBaseURL)
		if err != nil {
			return fmt.Errorf("parsing ProviderBaseURL %q: %w", v.ProviderBaseURL, err)
		}

		port := getPort(v.ProviderBaseURL)
		switch port {
		case portOutOfRange:
			return fmt.Errorf("port in 'ProviderBaseURL' %q is out of range, must be between 0 and 65535", v.ProviderBaseURL)
		case portUnknownScheme:
			return fmt.Errorf("unknown scheme '%s' given to 'ProviderBaseURL', unable to determine default port. Use 'Transports' for non-HTTP providers instead", url.Scheme)
		}

		//nolint:gosec // G115: getPort returns either portOutOfRange or portUnknownScheme
		// (both handled above) or a value already checked to be within 0-65535, so this
		// conversion cannot overflow uint16.
		handle.SetProviderInfo(v.Provider, url.Scheme, url.Hostname(), uint16(port), url.Path)

		log.Println("[DEBUG] v.Transports", v.Transports)
	}

	addPactUrlsFromEnvironment(v)

	filterDescription := valueOrFromEnvironment(v.FilterDescription, "PACT_DESCRIPTION")
	filterState := valueOrFromEnvironment(v.FilterState, "PACT_PROVIDER_STATE")
	filterNoState := valueOrFromEnvironment(strconv.FormatBool(v.FilterNoState), "PACT_PROVIDER_NO_STATE") == "true"

	if filterDescription != "" || filterState != "" || os.Getenv("PACT_PROVIDER_NO_STATE") != "" {
		handle.SetFilterInfo(filterDescription, filterState, filterNoState)
	}

	if v.RequestTimeout == 0 {
		v.RequestTimeout = defaultRequestTimeout
	}

	handle.SetVerificationOptions(v.DisableSSLVerification, v.RequestTimeout.Milliseconds())

	if v.PublishVerificationResults && v.ProviderVersion != "" {
		handle.SetPublishOptions(v.ProviderVersion, v.BuildURL, v.ProviderTags, v.ProviderBranch)
	}

	if len(v.FilterConsumers) > 0 {
		handle.SetConsumerFilters(v.FilterConsumers)
	}

	// TODO:
	// AddCustomHeader: 7,

	for _, url := range v.PactURLs {
		handle.AddURLSource(url, valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME"), valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD"), valueOrFromEnvironment(v.BrokerToken, "PACT_BROKER_TOKEN"))
	}

	for _, file := range v.PactFiles {
		handle.AddFileSource(file)
	}

	for _, dir := range v.PactDirs {
		handle.AddDirectorySource(dir)
	}

	if len(v.PactURLs) == 0 && len(v.PactFiles) == 0 && len(v.PactDirs) == 0 && v.BrokerURL == "" {
		return errors.New("one of 'PactURLs', 'PactFiles', 'PactDIRs' or 'BrokerURL' must be specified")
	}

	selectors := make([]string, len(v.ConsumerVersionSelectors))

	if len(v.ConsumerVersionSelectors) != 0 {
		for i, selector := range v.ConsumerVersionSelectors {
			body, err := json.Marshal(selector)
			if err != nil {
				return fmt.Errorf("invalid consumer version selector specified: %w", err)
			}

			selectors[i] = string(body)
		}
	}

	if valueOrFromEnvironment(v.BrokerURL, "PACT_BROKER_URL") != "" && (v.ProviderVersion == "" || v.Provider == "") {
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
		handle.BrokerSourceWithSelectors(valueOrFromEnvironment(v.BrokerURL, "PACT_BROKER_URL"), valueOrFromEnvironment(v.BrokerUsername, "PACT_BROKER_USERNAME"), valueOrFromEnvironment(v.BrokerPassword, "PACT_BROKER_PASSWORD"), valueOrFromEnvironment(v.BrokerToken, "PACT_BROKER_TOKEN"), v.EnablePending, includeWIPPactsSince, v.ProviderTags, v.ProviderBranch, selectors, v.Tags)
	}

	handle.SetNoPactsIsError(v.FailIfNoPactsFound)
	handle.SetColoredOutput(!v.DisableColoredOutput)

	return nil
}

// Sentinel returns from getPort, distinct from any valid port (0-65535).
const (
	// portUnknownScheme means the URL did not parse, or its scheme has no
	// default port. A caller that has already parsed the URL itself only ever
	// sees the latter.
	portUnknownScheme = -1
	// portOutOfRange means the URL carried an explicit port outside 0-65535.
	portOutOfRange = -2
)

// defaultHTTPSPort and defaultHTTPPort are the well-known ports assumed
// when a provider URL has no explicit port.
const (
	defaultHTTPSPort = 443
	defaultHTTPPort  = 80
)

// maxPort is the highest port number representable as a uint16.
const maxPort = 65535

// getPort returns the port of a URL, falling back to the default port for the
// scheme when the URL carries none. Only http and https have a default port.
func getPort(rawURL string) int {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return portUnknownScheme
	}

	// Port is empty unless the URL carries an explicit port, and strips the
	// brackets from an IPv6 host.
	if rawPort := parsedURL.Port(); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil || port < 0 || port > maxPort {
			return portOutOfRange
		}
		return port
	}

	switch parsedURL.Scheme {
	case "https":
		return defaultHTTPSPort
	case "http":
		return defaultHTTPPort
	default:
		return portUnknownScheme
	}
}
