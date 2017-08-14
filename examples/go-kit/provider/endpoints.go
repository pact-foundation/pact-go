package provider

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints is a container to hold all of our Transport routes (HTTP, RPC etc.).
type Endpoints struct {
	LoginEndpoint endpoint.Endpoint
}

// MakeServerEndpoints creates all of the routes.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		LoginEndpoint: MakeLoginEndpoint(s),
	}
}

// MakeLoginEndpoint returns the Login endpoint on the service.
func MakeLoginEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loginRequest)
		u, err := s.Login(req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return loginResponse{u}, nil
	}
}

func decodeUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}
