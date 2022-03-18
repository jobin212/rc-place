package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

//go:embed home.html
var resources embed.FS
var home = template.Must(template.ParseFS(resources, "home.html"))

var (
	// sessions stores user session information for browser login
	sessions = map[string]*Session{}

	oauthConf = &oauth2.Config{
		RedirectURL:  os.Getenv("OAUTH_REDIRECT"),
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.recurse.com/oauth/authorize",
			TokenURL: "https://www.recurse.com/oauth/token",
		},
	}
)

// Each session contains the user information and the oauth state
// to protect users from CSRF attacks.
// See https://pkg.go.dev/golang.org/x/oauth2#Config.AuthCodeURL
type Session struct {
	User
	State string
}

func (s Session) isAuthenticated() bool {
	return s.Id != 0
}

// verifyRoute is a helper function to check that a request has the expected
// method and path.
func verifyRoute(w http.ResponseWriter, r *http.Request, method, path string) bool {
	log.Println(r.URL)
	if r.URL.Path != path {
		http.Error(w, "Not found", http.StatusNotFound)
		return false
	}
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// serveHome serves the '/' route and the main application.
func serveHome(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/") {
		return
	}

	if session, err := getSession(r); err != nil || !session.isAuthenticated() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	home.Execute(w, nil)
}

// serveLogin serves the '/login' route for initializing the oauth flow.
func serveLogin(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/login") {
		return
	}

	currentSession, err := getSession(r)
	if err == nil && currentSession.isAuthenticated() {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Create a new random session token
	sessionToken := uuid.NewString()

	// Set the token in the session map, along with the session information
	sessions[sessionToken] = &Session{
		State: uuid.NewString(),
	}

	// Set the client cookie for "session_token" as the session token
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	url := oauthConf.AuthCodeURL(sessions[sessionToken].State, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// serveAuth serves the '/auth' route, which is the oauth callback location.
func serveAuth(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/auth") {
		return
	}
	// if no session exists, redirect to /login
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// if already authenticated, redirect to /
	if session.isAuthenticated() {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// authenticate user
	state := r.FormValue("state")
	code := r.FormValue("code")
	if state == "" || state != session.State {
		// missing state or does not match our saved value
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	if code == "" {
		// access denied in oauth flow
		http.Error(w, "Unoauthorized", http.StatusUnauthorized)
		return
	}

	// get a token from the authorization code
	tok, err := oauthConf.Exchange(context.TODO(), code)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// create a client to send authorized requests to recurse.com
	client := oauthConf.Client(context.TODO(), tok)
	resp, err := client.Get("https://recurse.com/api/v1/profiles/me")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// read body and log / display
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// save user in session
	session.User = user

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// getSession is a helper function to get the session struct from the request
// cookie. This function will return an error if the session is not found.
func getSession(r *http.Request) (*Session, error) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	sessionToken := c.Value

	// We then get the session from our session map
	userSession, exists := sessions[sessionToken]
	if !exists {
		return nil, errors.New("Session not found")
	}

	return userSession, nil
}
