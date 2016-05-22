package daemon

import (
	"reflect"
	"testing"
)

func TestStartDaemon(t *testing.T) {

}

func TestStartDaemon_Fail(t *testing.T) {

}

func TestStartServer(t *testing.T) {
	daemon := &Daemon{}
	req := MockServer{Pid: 1234}
	res := MockServer{}
	err := daemon.StartServer(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if !reflect.DeepEqual(req, res) {
		t.Fatalf("Req and Res did not match")
	}
}

func TestStartServer_Fail(t *testing.T) {

}

func TestVerification(t *testing.T) {

}

func TestVerification_Fail(t *testing.T) {

}

func TestPublish(t *testing.T) {

}

func TestPublish_Fail(t *testing.T) {

}
