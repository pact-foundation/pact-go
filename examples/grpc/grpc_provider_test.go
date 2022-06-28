//go:build provider
// +build provider

package grpc

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide"
	"github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide/server"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var dir, _ = os.Getwd()

func TestGrpcProvider(t *testing.T) {
	go startProvider()

	verifier := provider.PluginVerifier{}

	err := verifier.VerifyProvider(t, provider.VerifyPluginRequest{
		ProviderAddress: "http://localhost:8222",
		Provider:        "grpcprovider",
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/../pacts/grpcconsumer-grpcprovider.json", dir)),
		},
	})

	assert.NoError(t, err)
}

func startProvider() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8222))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, server.NewServer())
	grpcServer.Serve(lis)
}
