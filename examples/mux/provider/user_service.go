package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pact-foundation/pact-go/examples/types"
)

var userRepository = &types.UserRepository{
	Users: map[string]*types.User{
		"billy": &types.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "Jean-Marie de La Beaujardière😀😍",
			Password: "issilly",
		},
	},
}

// UserLogin is the login route.
var UserLogin = func(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusOK)
		res := types.LoginResponse{User: user}
		resBody, _ := json.Marshal(res)
		w.Write(resBody)
	}
}
