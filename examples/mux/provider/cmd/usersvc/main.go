package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/pact-foundation/pact-go/examples/mux/provider"
	"github.com/pact-foundation/pact-go/utils"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/login/", provider.UserLogin)

	port, _ := utils.GetFreePort()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", port, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))
}
