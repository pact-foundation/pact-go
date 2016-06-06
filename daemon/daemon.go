package daemon

// Runs the RPC daemon for remote communication

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"

	"github.com/pact-foundation/pact-go/types"
)

// Daemon wraps the commands for the RPC server.
type Daemon struct {
	pactMockSvcManager     Service
	verificationSvcManager Service
	signalChan             chan os.Signal
}

// NewDaemon returns a new Daemon with all instance variables initialised.
func NewDaemon(pactMockServiceManager Service, verificationServiceManager Service) *Daemon {
	pactMockServiceManager.Setup()
	verificationServiceManager.Setup()

	return &Daemon{
		pactMockSvcManager:     pactMockServiceManager,
		verificationSvcManager: verificationServiceManager,
		signalChan:             make(chan os.Signal, 1),
	}
}

// StartDaemon starts the daemon RPC server.
func (d *Daemon) StartDaemon(port int) {
	log.Println("[INFO] daemon - starting daemon on port", port)

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

	// Wait for sigterm
	signal.Notify(d.signalChan, os.Interrupt, os.Kill)
	s := <-d.signalChan
	log.Println("[INFO] daemon - received signal:", s, ", shutting down all services")

	d.Shutdown()
}

// StopDaemon allows clients to programmatically shuts down the running Daemon
// via RPC.
func (d *Daemon) StopDaemon(request string, reply *string) error {
	log.Println("[DEBUG] daemon - stop daemon")
	d.signalChan <- os.Interrupt
	return nil
}

// Shutdown ensures all services are cleanly destroyed.
func (d *Daemon) Shutdown() {
	log.Println("[DEBUG] daemon - shutdown")
	for _, s := range d.verificationSvcManager.List() {
		if s != nil {
			d.pactMockSvcManager.Stop(s.Process.Pid)
		}
	}
}

// StartServer starts a mock server and returns a pointer to atypes.PactMockServer
// struct.
func (d *Daemon) StartServer(request *types.PactMockServer, reply *types.PactMockServer) error {
	log.Println("[DEBUG] daemon - starting mock server")
	server := &types.PactMockServer{}
	port, svc := d.pactMockSvcManager.NewService(request.Args)
	server.Port = port
	server.Status = -1
	cmd := svc.Start()
	server.Pid = cmd.Process.Pid
	*reply = *server
	return nil
}

// VerifyProvider runs the Pact Provider Verification Process.
func (d *Daemon) VerifyProvider(request *types.VerifyRequest, reply *types.CommandResponse) error {
	log.Println("[DEBUG] daemon - verifying provider")
	exitCode := 1

	// Convert request into flags, and validate request
	err := request.Validate()
	if err != nil {
		*reply = *&types.CommandResponse{
			ExitCode: exitCode,
			Message:  err.Error(),
		}
		return nil
	}

	var out bytes.Buffer
	_, svc := d.verificationSvcManager.NewService(request.Args)
	cmd, err := svc.Run(&out)

	if cmd.ProcessState.Success() && err == nil {
		exitCode = 0
	}

	*reply = *&types.CommandResponse{
		ExitCode: exitCode,
		Message:  string(out.Bytes()),
	}

	return nil
}

// ListServers returns a slice of all running types.PactMockServers.
func (d *Daemon) ListServers(request types.PactMockServer, reply *types.PactListResponse) error {
	log.Println("[DEBUG] daemon - listing mock servers")
	var servers []*types.PactMockServer

	for port, s := range d.pactMockSvcManager.List() {
		servers = append(servers, &types.PactMockServer{
			Pid:  s.Process.Pid,
			Port: port,
		})
	}

	*reply = *&types.PactListResponse{
		Servers: servers,
	}

	return nil
}

// StopServer stops the given mock server.
func (d *Daemon) StopServer(request *types.PactMockServer, reply *types.PactMockServer) error {
	log.Println("[DEBUG] daemon - stopping mock server")
	success, err := d.pactMockSvcManager.Stop(request.Pid)
	if success == true && err == nil {
		request.Status = 0
	} else {
		request.Status = 1
	}
	*reply = *request

	return nil
}
