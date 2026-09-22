//go:build provider
// +build provider

package plugin

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	"github.com/stretchr/testify/assert"
)

var pactDir = fmt.Sprintf("%s/../pacts", dir)

func TestPluginProvider(t *testing.T) {
	t.Skip()
	httpPort, _ := utils.GetFreePort()
	tcpPort, _ := utils.GetFreePort()

	// Start provider API in the background
	go startHTTPProvider(httpPort)
	go startTCPServer(tcpPort)

	verifier := provider.NewVerifier()

	// Verify the Provider with local Pact Files
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("http://127.0.0.1:%d", httpPort),
		// Provider:        "provider",
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/MattConsumer-MattProvider.json", pactDir)),
			filepath.ToSlash(fmt.Sprintf("%s/matttcpconsumer-matttcpprovider.json", pactDir)),
		},
		Transports: []provider.Transport{
			{
				Protocol: "matt",
				//nolint:gosec // G115: tcpPort comes from utils.GetFreePort(), which returns
				// net.TCPAddr.Port, a value the kernel always assigns in the 0-65535 range.
				Port:   uint16(tcpPort),
				Scheme: "tcp",
			},
		},
	})

	assert.NoError(t, err)
}

func startHTTPProvider(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/matt", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/matt")
		w.WriteHeader(200)
		_, err := fmt.Fprintf(w, `MATTworldMATT`)
		if err != nil {
			log.Println("ERROR writing response body:", err)
		}
	})

	log.Printf("started HTTP server on port: %d\n", port)
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func startTCPServer(port int) {
	log.Println("Starting TCP server on port", port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("ERROR:", err)
	}

	log.Println("TCP server started on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("TCP connection error:", err)
			continue
		}

		log.Println("TCP connection established with:", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Println("Handling TCP connection")
	defer func() { _ = conn.Close() }() // best-effort cleanup; error is not actionable in a test server

	s := bufio.NewScanner(conn)

	for s.Scan() {
		data := s.Text()
		log.Println("Data received from connection", data)

		if data == "" {
			continue
		}

		handleRequest(data, conn)
	}
}

func handleRequest(req string, conn net.Conn) {
	log.Println("TCP Server received request", req, "on connection", conn)

	if !isValidMessage(req) {
		log.Println("TCP Server received invalid request, erroring")
		_, err := conn.Write([]byte("ERROR\n"))
		if err != nil {
			log.Println("ERROR writing to connection:", err)
		}
	}
	log.Println("TCP Server received valid request, responding")

	// var expectedResponse = "badworld"
	expectedResponse := "tcpworld"
	_, err := conn.Write([]byte(generateMattMessage(expectedResponse)))
	if err != nil {
		log.Println("ERROR writing to connection:", err)
	}
}

func isValidMessage(str string) bool {
	matched, err := regexp.MatchString(`^MATT.*MATT$`, str)
	if err != nil {
		return false
	}

	return matched
}
