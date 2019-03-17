package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pact-foundation/pact-go/examples/types"
)

var userRepository = &types.UserRepository{
	Users: map[string]*types.User{
		"jmarie": &types.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "jmarie",
			Password: "issilly",
			Type:     "admin",
			ID:       10,
		},
	},
}

// Crude time-bound "bearer" token
func getAuthToken() string {
	return time.Now().Format("2006-01-02")
}

// Simple authentication middleware
func IsAuthenticated(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == fmt.Sprintf("Bearer %s", getAuthToken()) {
			h.ServeHTTP(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

// UserLogin logs a user in, returning an auth token and the user object
func UserLogin(w http.ResponseWriter, r *http.Request) {
	var login types.LoginRequest
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Api-Correlation-Id", "1234")

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	err = json.Unmarshal(body, &login)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	user, err := userRepository.ByUsername(login.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else if user.Username != login.Username || user.Password != login.Password {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.Header().Set("X-Auth-Token", getAuthToken())
		w.WriteHeader(http.StatusOK)
		res := types.LoginResponse{User: user}
		resBody, _ := json.Marshal(res)
		w.Write(resBody)
	}
}

// GetUser fetches a user if authenticated and exists
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Api-Correlation-Id", "1234")

	// Get username from path
	a := strings.Split(r.URL.Path, "/")
	id, _ := strconv.Atoi(a[len(a)-1])

	user, err := userRepository.ByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal(user)
		w.Write(resBody)
	}
}
