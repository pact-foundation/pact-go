package daemon

// Runs the RPC daemon for remote communication

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"os/signal"
)

// PactMockServer contains the RPC client interface to a Mock Server
type PactMockServer struct {
	Pid    int
	Port   int
	Status int
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(ServiceMock{})
}

// PublishRequest contains the details required to Publish Pacts to a broker.
type PublishRequest struct {
	// Array of local Pact files or directories containing them. Required.
	PactUrls []url.URL

	// URL to fetch the provider states for the given provider API. Optional.
	PactBroker url.URL

	// Username for Pact Broker basic authentication. Optional
	PactBrokerUsername string

	// Password for Pact Broker basic authentication. Optional
	PactBrokerPassword string
}

// PactResponse contains the exit status and any message from the Broker.
type PactResponse struct {
	// System exit code from the Publish task.
	ExitCode int

	// Error message (if any) from the publish process.
	Message string
}

// PactListResponse contains a list of all running Servers.
type PactListResponse struct {
	// System exit code from the Publish task.
	Servers []*PactMockServer
}

// VerifyRequest contains the verification params.
type VerifyRequest struct{}

// Daemon wraps the commands for the RPC server.
type Daemon struct {
	pactMockSvcManager Service
	signalChan         chan os.Signal
}

// NewDaemon returns a new Daemon with all instance variables initialised.
func NewDaemon(pactMockServiceManager Service) *Daemon {
	pactMockServiceManager.Setup()

	return &Daemon{
		pactMockSvcManager: pactMockServiceManager,
		signalChan:         make(chan os.Signal, 1),
	}
}

// StartDaemon starts the daemon RPC server.
func (d *Daemon) StartDaemon(port int) {
	fmt.Println("Starting daemon on port", port)
	// rpc.Register(d)
	// rpc.HandleHTTP()
	//
	// // Start daemon in background
	// go func() {
	// 	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	// 	if e != nil {
	// 		log.Fatal("listen error:", e)
	// 	}
	// 	http.Serve(l, nil)
	// }()

	serv := rpc.NewServer()
	serv.Register(d)

	// ===== workaround ==========
	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	// ===========================

	serv.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	// ===== workaround ==========
	http.DefaultServeMux = oldMux
	// ===========================

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	go http.Serve(l, mux)

	// d.pactMockSvcManager.Start()

	// Wait for sigterm
	signal.Notify(d.signalChan, os.Interrupt, os.Kill)
	s := <-d.signalChan
	fmt.Println("Got signal:", s, ". Shutting down all services")

	d.Shutdown()
	fmt.Println("done")
}

// Shutdown ensures all services are cleanly destroyed.
func (d *Daemon) Shutdown() {
	for _, s := range d.pactMockSvcManager.List() {
		if s != nil {
			d.pactMockSvcManager.Stop(s.Process.Pid)
		}
	}
}

// StartServer starts a mock server and returns a pointer to aPactMockServer
// struct.
func (d *Daemon) StartServer(request *PactMockServer, reply *PactMockServer) error {
	reply = &PactMockServer{}
	port, svc := d.pactMockSvcManager.NewService()
	reply.Port = port
	cmd := svc.Start()
	reply.Pid = cmd.Process.Pid

	return nil
}

// ListServers returns a slice of all running PactMockServers.
func (d *Daemon) ListServers(request PactMockServer, reply *PactListResponse) error {
	var servers []*PactMockServer

	for port, s := range d.pactMockSvcManager.List() {
		servers = append(servers, &PactMockServer{
			Pid:  s.Process.Pid,
			Port: port,
		})
	}

	*reply = *&PactListResponse{
		Servers: servers,
	}

	return nil
}

// StopServer stops the given mock server.
func (d *Daemon) StopServer(request *PactMockServer, reply *PactMockServer) error {
	d.pactMockSvcManager.Stop(request.Pid)
	*reply = *request
	return nil
}

// Publish publishes Pact files from a given location (file/http).
func (d *Daemon) Publish(request *PublishRequest, reply *PactResponse) error {
	*reply = *&PactResponse{
		ExitCode: 0,
		Message:  "Success",
	}
	return nil
}

// Verify runs the Pact verification process against a given API Provider.
func (d *Daemon) Verify(request *VerifyRequest, reply *PactResponse) error {
	*reply = *&PactResponse{
		ExitCode: 0,
		Message:  "Success",
	}
	return nil
}
