// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

//go:embed home.html
var resources embed.FS
var home = template.Must(template.ParseFS(resources, "home.html"))

const updateLimit = 10 * time.Second

var addr = flag.String("addr", ":8080", "http service address")

// sessions stores user session information for browser login
var sessions = map[string]*session{}

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*user{}

var lastUpdateCache = map[string]time.Time{}

var conf = &oauth2.Config{
	RedirectURL:  os.Getenv("OAUTH_REDIRECT"),
	ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
	Scopes:       []string{},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.recurse.com/oauth/authorize",
		TokenURL: "https://www.recurse.com/oauth/token",
	},
}

type user struct {
	Id       int
	Username string
}

// Each session contains the user information and the oauth state
// to protect users from CSRF attacks.
// See https://pkg.go.dev/golang.org/x/oauth2#Config.AuthCodeURL
type session struct {
	user
	State string
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	if session.isAuthenticated() {
		home.Execute(w, nil)
		return
	}

	// authenticate user
	state := r.FormValue("state")
	code := r.FormValue("code")
	if state == "" || state != session.State {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	if code == "" {
		// access denied
		http.Error(w, "Unoauthorized", http.StatusUnauthorized)
		return
	}

	// get a token from the authorization code
	tok, err := conf.Exchange(context.TODO(), code)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// create a client to send authorized requests to recurse.com
	client := conf.Client(context.TODO(), tok)
	resp, err := client.Get("https://recurse.com/api/v1/profiles/me")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// read body and log / display
	defer resp.Body.Close()
	type jsonBody struct {
		Id   int    `json:"id"`
		Slug string `json:"slug"`
	}

	var j jsonBody
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.Id = j.Id
	session.Username = j.Slug

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s session) isAuthenticated() bool {
	return s.Id != 0
}

func (u *user) SetTile(hub *Hub, x, y int, color string) error {
	if time.Since(lastUpdateCache[u.Username]) < updateLimit {
		return errors.New("rate limited")
	}
	hub.broadcast <- []byte(fmt.Sprintf("%d %d %s", x, y, color))
	lastUpdateCache[u.Username] = time.Now()
	return nil
}

func getSession(r *http.Request) (*session, error) {
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

func serveTile(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// TODO: respond with JSON bodies always
	log.Println(r.URL)
	if r.URL.Path != "/tile" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// authenticate
	user, err := authPersonalAccessToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	type jsonBody struct {
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Color string `json:"color"`
	}
	var j jsonBody
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := user.SetTile(hub, j.X, j.Y, j.Color); err != nil {
		log.Println(err)
		http.Error(w, "Too Early", http.StatusTooEarly)
		return
	}
}

// authPersonalAccessToken will authenticate an Authorization header by
// forwarding a request to recurse.com API and cache a successful result
// in pacCache.
func authPersonalAccessToken(r *http.Request) (*user, error) {
	// get token
	pacToken := r.Header.Get("Authorization")
	if pacToken == "" {
		return nil, errors.New("missing authentication token")
	}
	// check cache
	if u, ok := pacCache[pacToken]; ok {
		return u, nil
	}
	// send request to recurse.com
	req, err := http.NewRequest(http.MethodGet, "https://recurse.com/api/v1/profiles/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", pacToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unauthorized")
	}

	// read body
	defer resp.Body.Close()
	type jsonBody struct {
		Id   int    `json:"id"`
		Slug string `json:"slug"`
	}

	var j jsonBody
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return nil, err
	}

	// update cache
	pacCache[pacToken] = &user{
		Id:       j.Id,
		Username: j.Slug,
	}
	return pacCache[pacToken], nil
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/login" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	sessions[sessionToken] = &session{
		State: uuid.NewString(),
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	url := conf.AuthCodeURL(sessions[sessionToken].State, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func main() {
	flag.Parse()

	// check oauth environment variables
	abort := false
	for _, env := range []string{"OAUTH_REDIRECT", "OAUTH_CLIENT_ID", "OAUTH_CLIENT_SECRET"} {
		if os.Getenv(env) == "" {
			log.Println("Required environment variable missing:", env)
			abort = true
		}
	}
	if abort {
		log.Println("Aborting")
		os.Exit(1)
	}

	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/tile", func(w http.ResponseWriter, r *http.Request) {
		serveTile(hub, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		session, err := getSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		serveWs(hub, &session.user, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
