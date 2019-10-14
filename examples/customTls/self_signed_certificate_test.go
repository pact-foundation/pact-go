package provider

import (
	"fmt"
	"os"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/utils"
	"path/filepath"
	"testing"
	"crypto/tls"
	"crypto/x509"
	"github.com/pact-foundation/pact-go/types"
	"io/ioutil"
	"log"
	"net/http"	
)

// An external HTTPS provider
func TestPact_SelfSignedTLSProvider(t *testing.T) {
	go startServer()
	
	pact := createPact()
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("https://localhost:%d", port),
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/consumer-selfsignedtls.json", pactDir))},
		CustomTLSConfig: &tls.Config{
			RootCAs: getCaCertPool(),
			// InsecureSkipVerify: true, // Disable SSL verification altogether
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func HelloServer(w http.ResponseWriter, req *http.Request) { }

func startServer() {
	http.HandleFunc("/hello", HelloServer)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		ClientCAs: getCaCertPool(),
		ClientAuth: tls.NoClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		TLSConfig: tlsConfig,
	}

	log.Fatalf("%v", server.ListenAndServeTLS("certs/server-cert.pem", "certs/server-key.pem"))
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer:                 "consumer",
		Provider:                 "selfsignedtls",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
		LogLevel:                 "DEBUG",
	}
}

// Custom certificate authority
func getCaCertPool() *x509.CertPool {
	caCert, err := ioutil.ReadFile("certs/ca.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return caCertPool
}