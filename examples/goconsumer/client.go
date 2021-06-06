package goconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	ex "github.com/pact-foundation/pact-go/examples/types"
)

// Client is a UI for the User Service.
type Client struct {
	user  *ex.User
	Host  string
	err   error
	token string
}

// Marshalling format for Users.
type loginResponse struct {
	User ex.User `json:"user"`
}

type templateData struct {
	User  *ex.User
	Error error
}

var loginTemplatePath = "login.html"
var templates = template.Must(template.ParseFiles(loginTemplatePath))

// getUser finds a user
func (c *Client) getUser(id string) (*ex.User, error) {

	u := fmt.Sprintf("%s/users/%s", c.Host, id)
	req, err := http.NewRequest("GET", u, nil)

	// NOTE: by default, request bodies are expected to be sent with a Content-Type
	// of application/json. If you don't explicitly set the content-type, you
	// will get a mismatch during Verification.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token)

	res, err := http.DefaultClient.Do(req)

	if res.StatusCode != 200 || err != nil {
		return nil, fmt.Errorf("get user failed")
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response ex.User
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response, err
}

// Login handles the login API call to the User Service.
func (c *Client) login(username string, password string) (*ex.User, error) {
	loginRequest := fmt.Sprintf(`
    {
      "username":"%s",
      "password": "%s"
    }`, username, password)

	res, err := http.Post(fmt.Sprintf("%s/login/10?foo=anything", c.Host), "application/json; charset=utf-8", bytes.NewReader([]byte(loginRequest)))
	if res.StatusCode != 200 || err != nil {
		return nil, fmt.Errorf("login failed")
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
	if err == nil && user != nil {
		c.user = user
		c.err = nil
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	c.err = fmt.Errorf("Invalid username/password")
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// Deal with the logout request.
func (c *Client) logoutHandler(w http.ResponseWriter, r *http.Request) {
	c.user = nil
	c.err = nil
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func renderTemplate(w http.ResponseWriter, tmpl string, u templateData) {

	err := templates.ExecuteTemplate(w, tmpl+".html", u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show the current user if logged in, otherwise display a login form.
func (c *Client) viewHandler(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		User:  c.user,
		Error: c.err,
	}
	renderTemplate(w, "login", data)
}

// Run the web application.
func (c *Client) Run() {
	http.HandleFunc("/login", c.loginHandler)
	http.HandleFunc("/logout", c.logoutHandler)
	http.HandleFunc("/", c.viewHandler)
	fmt.Println("User svc client running on port 8081")
	http.ListenAndServe(":8081", nil)
}
