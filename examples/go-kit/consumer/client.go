package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

type templateData struct {
	User *provider.User
}

var templates = template.Must(template.ParseFiles("../../login.html"))

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

func renderTemplate(w http.ResponseWriter, tmpl string, u templateData) {

	err := templates.ExecuteTemplate(w, tmpl+".html", u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show the current user if logged in, otherwise display a login form.
func (c *Client) viewHandler(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		User: c.user,
	}
	renderTemplate(w, "login", data)
}

// Run the web application.
func (c *Client) Run() {
	http.HandleFunc("/login", c.loginHandler)
	http.HandleFunc("/", c.viewHandler)
	http.ListenAndServe(":8081", nil)
}
