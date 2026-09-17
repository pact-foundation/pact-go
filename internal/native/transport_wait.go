package native

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	// transportAcceptTimeout bounds how long waitForTransport waits for a mock
	// server to accept connections.
	transportAcceptTimeout = 10 * time.Second
	// transportPollInterval is how long waitForTransport waits between attempts.
	transportPollInterval = 10 * time.Millisecond
)

// waitForTransport blocks until the mock server on port accepts a connection,
// or transportAcceptTimeout elapses.
//
// pactffi_create_mock_server_for_transport returns as soon as the port is
// allocated. For a plugin transport that port belongs to the plugin, which
// reports it over gRPC before its listener is necessarily accepting, so a test
// that dials immediately can be refused on a loaded machine. Waiting here keeps
// that race out of every caller.
//
// A timeout is logged rather than returned: the caller's own connection error
// describes the failure better than a wait that gave up, and returning one here
// would turn a slow start into a hard failure.
func waitForTransport(address string, port int) {
	target := net.JoinHostPort(address, strconv.Itoa(port))

	ctx, cancel := context.WithTimeout(context.Background(), transportAcceptTimeout)
	defer cancel()

	dialer := &net.Dialer{}

	for attempt := 0; ; attempt++ {
		conn, err := dialer.DialContext(ctx, "tcp", target)
		if err == nil {
			_ = conn.Close()
			if attempt > 0 {
				log.Println("[DEBUG] mock server on", target, "accepted a connection after", attempt, "retries")
			}

			return
		}

		if ctx.Err() != nil {
			log.Println("[WARN] mock server on", target, "did not accept a connection within", transportAcceptTimeout)

			return
		}

		time.Sleep(transportPollInterval)
	}
}
