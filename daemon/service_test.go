package daemon

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
	env := []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=true")}
	mgr := &ServiceManager{
		Command: os.Args[0],
		Args:    cs,
		Env:     env,
	}
	mgr.Setup()
	return mgr
}

func TestServiceManager(t *testing.T) {
	var manager interface{}
	manager = new(ServiceManager)

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
	cmd.Start()
	mgr.processes = map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}

	mgr.commandCompleteChan <- cmd
	var timeout = time.After(channelTimeout)
	for {
		select {
		case <-time.After(10 * time.Millisecond):
			if len(mgr.processes) == 0 {
				return
			}
		case <-timeout:
			if len(mgr.processes) != 0 {
				t.Fatalf(`Expected 1 command to be removed from the queue. Have %d
          Timed out after 500millis`, len(mgr.processes))
			}
		}
	}
}

func TestServiceManager_addServiceMonitor(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start()
	mgr.commandCreatedChan <- cmd
	var timeout = time.After(channelTimeout)

	for {
		select {
		case <-time.After(10 * time.Millisecond):
			if len(mgr.processes) == 1 {
				return
			}
		case <-timeout:
			if len(mgr.processes) != 1 {
				t.Fatalf(`Expected 1 command to be added to the queue, but got: %d.
          Timed out after 500millis`, len(mgr.processes))
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
			if len(mgr.processes) != 0 {
				t.Fatalf(`Expected 0 command to be added to the queue, but got: %d.
          Timed out after 5 attempts`, len(mgr.processes))
			}
		case <-timeout:
			if len(mgr.processes) != 0 {
				t.Fatalf(`Expected 0 command to be added to the queue, but got: %d.
				Timed out after 50millis`, len(mgr.processes))
			}
			return
		}
	}
}

func TestServiceManager_Stop(t *testing.T) {
	mgr := createServiceManager()
	cmd := fakeExecCommand("", true, "")
	cmd.Start()
	mgr.processes = map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}

	mgr.Stop(cmd.Process.Pid)
	var timeout = time.After(channelTimeout)
	for {
		select {
		case <-time.After(10 * time.Millisecond):
			if len(mgr.processes) == 0 {
				return
			}
		case <-timeout:
			if len(mgr.processes) != 0 {
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
	cmd.Start()
	processes := map[int]*exec.Cmd{
		cmd.Process.Pid: cmd,
	}
	mgr.processes = processes

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
			if len(mgr.processes) == 1 {
				return
			}
		case <-timeout:
			if len(mgr.processes) != 1 {
				t.Fatalf(`Expected 1 command to be added to the queue, but got: %d.
          Timed out after 500millis`, len(mgr.processes))
			}
			return
		}
	}
}
