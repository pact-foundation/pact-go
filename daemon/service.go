package daemon

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

// Service is a process wrapper for 3rd party binaries. It will spawn an instance
// of the binary and manage the life-cycle and IO of the process.
type Service interface {
	Setup()
	Stop(pid int) (bool, error)
	List() map[int]*exec.Cmd
	Start() *exec.Cmd
	NewService(args []string) (int, Service)
}

// ServiceManager is the default implementation of the Service interface.
type ServiceManager struct {
	Command             string
	processes           map[int]*exec.Cmd
	Args                []string
	Env                 []string
	commandCompleteChan chan *exec.Cmd
	commandCreatedChan  chan *exec.Cmd
}

// Setup the Management services.
func (s *ServiceManager) Setup() {
	s.commandCreatedChan = make(chan *exec.Cmd)
	s.commandCompleteChan = make(chan *exec.Cmd)
	s.processes = make(map[int]*exec.Cmd)

	// Listen for service create/kill
	go s.addServiceMonitor()
	go s.removeServiceMonitor()
}

// addServiceMonitor watches a channel to add services into operation.
func (s *ServiceManager) addServiceMonitor() {
	for {
		select {
		case p := <-s.commandCreatedChan:
			if p != nil && p.Process != nil {
				s.processes[p.Process.Pid] = p
			}
		}
	}
}

// removeServiceMonitor watches a channel to remove services from operation.
func (s *ServiceManager) removeServiceMonitor() {
	var p *exec.Cmd
	for {
		select {
		case p = <-s.commandCompleteChan:
			if p != nil && p.Process != nil {
				p.Process.Signal(os.Interrupt)
				delete(s.processes, p.Process.Pid)
			}
		}
	}
}

// Stop a Service and returns the exit status.
func (s *ServiceManager) Stop(pid int) (bool, error) {
	cmd := s.processes[pid]

	// Remove service from registry
	go func() {
		s.commandCompleteChan <- cmd
	}()

	err := cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return false, err
	}

	return cmd.ProcessState.Success(), nil
}

// List all Service PIDs.
func (s *ServiceManager) List() map[int]*exec.Cmd {
	return s.processes
}

// Start a Service and log its output.
func (s *ServiceManager) Start() *exec.Cmd {
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Env = s.Env

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	cmdReaderErr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("error: | %s\n", scanner.Text())
		}
	}()

	scanner2 := bufio.NewScanner(cmdReaderErr)
	go func() {
		for scanner2.Scan() {
			fmt.Printf("mock-service:  %s\n", scanner2.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	// Add service to registry
	s.commandCreatedChan <- cmd

	return cmd
}
