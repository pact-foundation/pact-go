package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pact-foundation/pact-go/examples/go-kit/provider"
)

// Client is a UI for the User Service.
type Client struct {
	user *provider.User
	Host string
}

// Marshalling format for Users.
type loginResponse struct {
	User provider.User `json:"user"`
}

// Login handles the login API call to the User Service.
func (c *Client) login(username string, password string) (*provider.User, error) {
	loginRequest := fmt.Sprintf(`
    {
      "username":"%s",
      "password": "%s"
    }`, username, password)

	res, err := http.Post(fmt.Sprintf("%s/users/login", c.Host), "application/json", bytes.NewReader([]byte(loginRequest)))
	if res.StatusCode != 200 || err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response loginResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response.User, err
}

// Deal with the login request.
func (c *Client) loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := c.login(username, password)
	if err == nil {
		c.user = user
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// Show the current user if logged in, otherwise display a login form.
func (c *Client) viewHandler(w http.ResponseWriter, r *http.Request) {
	if c.user != nil {
		fmt.Fprintf(w, "<h1>Hello %s</h1>", c.user.Name)
	} else {
		fmt.Fprintf(w, "<h1>Login to my awesome website:</h1>"+
			"<form action=\"/login\" method=\"POST\">"+
			"<input type=\"text\" name=\"username\">"+
			"<input type=\"password\" name=\"password\">"+
			"<input type=\"submit\" value=\"login\">"+
			"</form>")

	}
}

// Run the web application.
func (c *Client) Run() {
	http.HandleFunc("/login", c.loginHandler)
	http.HandleFunc("/", c.viewHandler)
	http.ListenAndServe(":8081", nil)
}
