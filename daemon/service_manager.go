package daemon

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

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
	log.Println("[DEBUG] setting up a service manager")
	s.commandCreatedChan = make(chan *exec.Cmd)
	s.commandCompleteChan = make(chan *exec.Cmd)
	s.processes = make(map[int]*exec.Cmd)

	// Listen for service create/kill
	go s.addServiceMonitor()
	go s.removeServiceMonitor()
}

// addServiceMonitor watches a channel to add services into operation.
func (s *ServiceManager) addServiceMonitor() {
	log.Println("[DEBUG] starting service creation monitor")
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
	log.Println("[DEBUG] starting service removal monitor")
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
	log.Println("[DEBUG] stopping service with pid", pid)
	cmd := s.processes[pid]

	// Remove service from registry
	go func() {
		s.commandCompleteChan <- cmd
	}()

	// Wait for error, kill if it takes too long
	var err error
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(3 * time.Second):
		if err = cmd.Process.Kill(); err != nil {
			log.Println("[ERROR] timeout reached, killing pid", pid)

			return false, err
		}
	case err = <-done:
		if err != nil {
			log.Println("[ERROR] error waiting for process to complete", err)
			return false, err
		}
	}

	return true, nil
}

// List all Service PIDs.
func (s *ServiceManager) List() map[int]*exec.Cmd {
	log.Println("[DEBUG] listing services")
	return s.processes
}

// Run runs a service synchronously and log its output to the given Pipe.
func (s *ServiceManager) Run(w io.Writer) (*exec.Cmd, error) {
	log.Println("[DEBUG] starting service")
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Env = s.Env
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()

	return cmd, err
}

// Start a Service and log its output.
func (s *ServiceManager) Start() *exec.Cmd {
	log.Println("[DEBUG] starting service")
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Env = s.Env

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[ERROR] unable to create output pipe for cmd: %s\n", err.Error())
		os.Exit(1)
	}

	cmdReaderErr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[ERROR] unable to create error pipe for cmd: %s\n", err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Printf("[ERROR] %s\n", scanner.Text())
		}
	}()

	scanner2 := bufio.NewScanner(cmdReaderErr)
	go func() {
		for scanner2.Scan() {
			log.Printf("[INFO] service: %s\n", scanner2.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Println("[ERROR] service", err.Error())
		os.Exit(1)
	}

	// Add service to registry
	s.commandCreatedChan <- cmd

	return cmd
}
