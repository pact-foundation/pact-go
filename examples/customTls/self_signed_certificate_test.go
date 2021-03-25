// +build provider

package provider

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// An external HTTPS provider
func TestExample_SelfSignedTLSProvider(t *testing.T) {
	go startServer()

	pact := createPact()
	// time.Sleep(100 * time.Second)

	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("https://localhost:%d", port),
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/consumer-selfsignedtls.json", pactDir))},
		CustomTLSConfig: &tls.Config{
			RootCAs:            getCaCertPool(),
			InsecureSkipVerify: true, // Disable SSL verification altogether
		},
		PactLogDir:   logDir,
		PactLogLevel: "INFO",
	})

	if err != nil {
		t.Fatal(err)
	}
}

func HelloServer(w http.ResponseWriter, req *http.Request) {}

func startServer() {
	http.HandleFunc("/hello", HelloServer)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		ClientCAs:  getCaCertPool(),
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

// Generate a certificate with self-signed CA
// openssl genrsa -des3 -out ca.key 2048
// openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.pem
// openssl genrsa -out server-key.pem 2048
// openssl req -new -key server-key.pem -out server.csr // Set "localhost" as the common name
// openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server-cert.pem -days 3650 -sha256
