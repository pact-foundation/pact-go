package provider

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Request // Response types
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	User *User `json:"user"`
}

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a profilesvc server.
func MakeHTTPHandler(ctx context.Context, s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
	e := MakeServerEndpoints(s)

	r.Methods("POST").Path("/users/login/{id}").Handler(httptransport.NewServer(
		e.LoginEndpoint,
		decodeUserRequest,
		encodeResponse,
		options...,
	))

	return r
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Api-Correlation-Id", "1234")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Api-Correlation-Id", "1234")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrNotFound:
		return http.StatusNotFound
	case ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
	// if e, ok := err.(httptransport.Error); ok {
	// 	switch e.Err {domaind
	// 	switch e.Domain {
	// 	case httptransport.DomainDecode:
	// 		return http.StatusBadRequest
	// 	case httptransport.DomainDo:
	// 		return http.StatusServiceUnavailable
	// 	default:
	// 		return http.StatusInternalServerError
	// 	}
	// }
	// return http.StatusInternalServerError
}
