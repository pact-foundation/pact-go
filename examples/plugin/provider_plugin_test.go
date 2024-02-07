//go:build provider
// +build provider

package plugin

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	// "github.com/pact-foundation/pact-go/v2/log"
	"log"

	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()
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
			provider.Transport{
				Protocol: "matt",
				Port:     uint16(tcpPort),
				Scheme:   "tcp",
			},
		},
	})

	assert.NoError(t, err)
}

func startHTTPProvider(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/matt", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/matt")
		fmt.Fprintf(w, `MATTworldMATT`)
		w.WriteHeader(200)
	})

	log.Printf("started HTTP server on port: %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux))
}

func startTCPServer(port int) {
	log.Println("Starting TCP server on port", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
	defer conn.Close()

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
		conn.Write([]byte("ERROR\n"))
	}
	log.Println("TCP Server received valid request, responding")

	// var expectedResponse = "badworld"
	var expectedResponse = "tcpworld"
	conn.Write([]byte(generateMattMessage(expectedResponse)))
	conn.Write([]byte("\n"))
}

func generateMattMessage(message string) string {
	return fmt.Sprintf("MATT%sMATT", message)
}

func parseMattMessage(message string) string {
	return strings.ReplaceAll(message, "MATT", "")
}

func isValidMessage(str string) bool {
	matched, err := regexp.MatchString(`^MATT.*MATT$`, str)
	if err != nil {
		return false
	}

	return matched
}
