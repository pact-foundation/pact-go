package fakes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
)

// Publication is a verification result published to the Broker.
type Publication struct {
	Consumer string
	Body     map[string]any
}

// Broker is the part of a Pact Broker a provider verification uses: it
// serves pacts for one provider, and records the verification results
// published back.
type Broker struct {
	server    *httptest.Server
	provider  string
	mu        sync.Mutex
	pacts     map[string][]byte
	consumers []string
	published []Publication
}

// NewBroker starts a broker for provider.
func NewBroker(provider string) *Broker {
	b := &Broker{provider: provider, pacts: map[string][]byte{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", b.index)
	mux.HandleFunc("GET /pacts/provider/{provider}/for-verification", b.pactsForVerificationLink)
	mux.HandleFunc("POST /pacts/provider/{provider}/for-verification", b.pactsForVerification)
	mux.HandleFunc("GET /pacts/provider/{provider}/consumer/{consumer}", b.pact)
	mux.HandleFunc("POST /pacts/provider/{provider}/consumer/{consumer}/verification-results", b.publish)
	b.server = httptest.NewServer(mux)

	return b
}

// URL is the broker's base URL.
func (b *Broker) URL() string { return b.server.URL }

// Close stops the broker.
func (b *Broker) Close() { b.server.Close() }

// AddPact makes pact, a pact file between consumer and the broker's
// provider, available for verification.
func (b *Broker) AddPact(consumer string, pact []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.pacts[consumer]; !ok {
		b.consumers = append(b.consumers, consumer)
	}
	b.pacts[consumer] = pact
}

// Publications returns the verification results published so far.
func (b *Broker) Publications() []Publication {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.published)
}

func (b *Broker) pactURL(consumer string) string {
	return b.URL() + "/pacts/provider/" + b.provider + "/consumer/" + consumer
}

// HAL keys used in the documents the Broker serves.
const (
	halLinks = "_links"
	halHref  = "href"
	halSelf  = "self"
)

func (b *Broker) index(w http.ResponseWriter, _ *http.Request) {
	writeHAL(w, http.StatusOK, map[string]any{halLinks: map[string]any{
		"pb:provider-pacts-for-verification": map[string]any{
			halHref:     b.URL() + "/pacts/provider/{provider}/for-verification",
			"templated": true,
		},
	}})
}

// pactsForVerificationLink answers the verifier's GET of the for-verification
// URL with just a self link, as the reference implementation's broker fixture
// does; the pacts themselves come from the POST.
func (b *Broker) pactsForVerificationLink(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("provider") != b.provider {
		http.NotFound(w, r)

		return
	}

	writeHAL(w, http.StatusOK, map[string]any{halLinks: map[string]any{halSelf: map[string]any{
		halHref: b.URL() + "/pacts/provider/" + b.provider + "/for-verification",
		"title": "Pacts to be verified",
	}}})
}

func (b *Broker) pactsForVerification(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("provider") != b.provider {
		http.NotFound(w, r)

		return
	}

	b.mu.Lock()
	consumers := slices.Clone(b.consumers)
	b.mu.Unlock()

	pacts := make([]map[string]any, 0, len(consumers))
	for _, consumer := range consumers {
		pacts = append(pacts, map[string]any{
			"shortDescription":       "latest",
			"verificationProperties": map[string]any{"pending": false, "notices": []any{}},
			halLinks: map[string]any{halSelf: map[string]any{
				halHref: b.pactURL(consumer),
				"name":  "Pact between " + consumer + " and " + b.provider,
			}},
		})
	}
	writeHAL(w, http.StatusOK, map[string]any{
		"_embedded": map[string]any{"pacts": pacts},
		halLinks: map[string]any{halSelf: map[string]any{
			halHref: b.URL() + "/pacts/provider/" + b.provider + "/for-verification",
		}},
	})
}

func (b *Broker) pact(w http.ResponseWriter, r *http.Request) {
	consumer := r.PathValue("consumer")
	b.mu.Lock()
	raw, ok := b.pacts[consumer]
	b.mu.Unlock()
	if r.PathValue("provider") != b.provider || !ok {
		http.NotFound(w, r)

		return
	}

	var doc map[string]any
	err := json.Unmarshal(raw, &doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	doc[halLinks] = map[string]any{"pb:publish-verification-results": map[string]any{
		"title": "Publish verification results",
		halHref: b.pactURL(consumer) + "/verification-results",
	}}
	writeHAL(w, http.StatusOK, doc)
}

func (b *Broker) publish(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	b.mu.Lock()
	b.published = append(b.published, Publication{Consumer: r.PathValue("consumer"), Body: body})
	b.mu.Unlock()
	writeHAL(w, http.StatusCreated, map[string]any{})
}

func writeHAL(w http.ResponseWriter, status int, doc any) {
	data, err := json.Marshal(doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/hal+json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
