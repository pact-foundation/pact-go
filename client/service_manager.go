package client

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ServiceManager is the default implementation of the Service interface.
type ServiceManager struct {
	Cmd                 string
	processMap          processMap
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
	s.processMap = processMap{processes: make(map[int]*exec.Cmd)}

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
				s.processMap.Set(p.Process.Pid, p)
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
				s.processMap.Delete(p.Process.Pid)
			}
		}
	}
}

// Stop a Service and returns the exit status.
func (s *ServiceManager) Stop(pid int) (bool, error) {
	log.Println("[DEBUG] stopping service with pid", pid)
	cmd := s.processMap.Get(pid)

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
	return s.processMap.processes
}

// Command creates an os command to be run
func (s *ServiceManager) Command() *exec.Cmd {
	cmd := exec.Command(s.Cmd, s.Args...)
	env := os.Environ()
	env = append(env, s.Env...)
	cmd.Env = env

	return cmd
}

// Start a Service and log its output.
func (s *ServiceManager) Start() *exec.Cmd {
	log.Println("[DEBUG] starting service")
	cmd := exec.Command(s.Cmd, s.Args...)
	env := os.Environ()
	env = append(env, s.Env...)
	cmd.Env = env

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[ERROR] unable to create output pipe for cmd: %s\n", err.Error())
	}

	cmdReaderErr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[ERROR] unable to create error pipe for cmd: %s\n", err.Error())
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Printf("[INFO] %s\n", scanner.Text())
		}
	}()

	scanner2 := bufio.NewScanner(cmdReaderErr)
	go func() {
		for scanner2.Scan() {
			log.Printf("[ERROR] service: %s\n", scanner2.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Println("[ERROR] service", err.Error())
	}

	// Add service to registry
	s.commandCreatedChan <- cmd

	return cmd
}

type processMap struct {
	sync.RWMutex
	processes map[int]*exec.Cmd
}

func (pm *processMap) Get(k int) *exec.Cmd {
	pm.RLock()
	defer pm.RUnlock()
	v, _ := pm.processes[k]
	return v
}

func (pm *processMap) Set(k int, v *exec.Cmd) {
	pm.Lock()
	defer pm.Unlock()
	pm.processes[k] = v
}

func (pm *processMap) Delete(k int) {
	pm.Lock()
	defer pm.Unlock()
	delete(pm.processes, k)
}
