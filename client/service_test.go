package client

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

var channelTimeout = 50 * time.Millisecond

func createServiceManager() *ServiceManager {
	cs := []string{"-test.run=TestHelperProcess", "--", os.Args[0]}
	env := []string{"GO_WANT_HELPER_PROCESS=1", "GO_WANT_HELPER_PROCESS_TO_SUCCEED=true"}
	mgr := &ServiceManager{
		Cmd:  os.Args[0],
		Args: cs,
		Env:  env,
	}
	mgr.Setup()
	return mgr
}

func TestServiceManager(t *testing.T) {
	var manager interface{} = new(ServiceManager)
	if _, ok := manager.(*ServiceManager); !ok {
		t.Fatalf("Must be a ServiceManager")
	}
}

func TestServiceManager_Setup(t *testing.T) {
	mgr := createServiceManager()

	if mgr.commandCompleteChan == nil {
		t.Fatalf("Expected commandCompleteChan to be non-nil but got nil")
	}

	if mgr.commandCreatedChan == nil {
		t.Fatalf("Expected commandCreatedChan to be non-nil but got nil")
	}
}

func TestServiceManager_removeServiceMonitor(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start() // nolint:errcheck
	mgr.processMap.processes = map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}

	mgr.commandCompleteChan <- cmd
	var timeout = time.After(channelTimeout)
	for {

		select {
		case <-time.After(10 * time.Millisecond):
			mgr.processMap.Lock()
			if len(mgr.processMap.processes) == 0 {
				mgr.processMap.Unlock()
				return
			}
		case <-timeout:
			if len(mgr.processMap.processes) != 0 {
				t.Fatalf(`Expected 1 command to be removed from the queue. Have %d
          Timed out after 500millis`, len(mgr.processMap.processes))
			}
		}
	}
}

func TestServiceManager_addServiceMonitor(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start() // nolint:errcheck
	mgr.commandCreatedChan <- cmd
	var timeout = time.After(channelTimeout)

	for {

		select {
		case <-time.After(10 * time.Millisecond):
			mgr.processMap.Lock()
			defer mgr.processMap.Unlock()
			if len(mgr.processMap.processes) == 1 {
				return
			}
		case <-timeout:
			mgr.processMap.Lock()
			defer mgr.processMap.Unlock()
			if len(mgr.processMap.processes) != 1 {
				t.Fatalf(`Expected 1 command to be added to the queue, but got: %d.
          Timed out after 500millis`, len(mgr.processMap.processes))
			}
			return
		}
	}
}

func TestServiceManager_addServiceMonitorWithDeadJob(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	mgr.commandCreatedChan <- cmd
	var timeout = time.After(channelTimeout)

	for {

		select {
		case <-time.After(10 * time.Millisecond):

			if len(mgr.processMap.processes) != 0 {
				t.Fatalf(`Expected 0 command to be added to the queue, but got: %d.
        Timed out after 5 attempts`, len(mgr.processMap.processes))
			}
		case <-timeout:
			if len(mgr.processMap.processes) != 0 {
				t.Fatalf(`Expected 0 command to be added to the queue, but got: %d.
				Timed out after 50millis`, len(mgr.processMap.processes))
			}
			return
		}
	}
}

func TestServiceManager_Stop(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start() // nolint:errcheck
	mgr.processMap.processes = map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}

	mgr.Stop(cmd.Process.Pid) // nolint:errcheck
	var timeout = time.After(channelTimeout)
	for {
		mgr.processMap.Lock()
		defer mgr.processMap.Unlock()

		select {
		case <-time.After(10 * time.Millisecond):
			if len(mgr.processMap.processes) == 0 {
				return
			}
		case <-timeout:
			if len(mgr.processMap.processes) != 0 {
				t.Fatalf(`Expected 1 command to be removed from the queue.
          Timed out after 500millis`)
			}
			return
		}
	}
}

func TestServiceManager_List(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start() // nolint:errcheck
	processes := map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}
	mgr.processMap.Lock()
	mgr.processMap.processes = processes
	mgr.processMap.Unlock()

	if !reflect.DeepEqual(processes, mgr.List()) {
		t.Fatalf("Expected mgr.List() to equal processes")
	}
}

func TestServiceManager_Start(t *testing.T) {
	mgr := createServiceManager()
	mgr.Start()
	var timeout = time.After(channelTimeout)

	for {

		select {
		case <-time.After(10 * time.Millisecond):
			mgr.processMap.Lock()
			if len(mgr.processMap.processes) == 1 {
				mgr.processMap.Unlock()
				return
			}
		case <-timeout:
			mgr.processMap.Lock()
			defer mgr.processMap.Unlock()
			if len(mgr.processMap.processes) != 1 {
				t.Fatalf(`Expected 1 command to be added to the queue, but got: %d.
          Timed out after 500millis`, len(mgr.processMap.processes))
			}
			return
		}
	}
}

func fakeExecCommand(command string, success bool, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=%t", success)}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	<-time.After(250 * time.Millisecond)

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		fmt.Fprintf(os.Stdout, "COMMAND: oh noes!")
		os.Exit(1)
	}

	// Success :)
	fmt.Fprintf(os.Stdout, `{"summary_line":"1 examples, 0 failures"}`)
	os.Exit(0)
}
