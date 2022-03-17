// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var addr = flag.String("addr", ":8080", "http service address")

// // stateMap is used for oauth redirection to protect users from CSRF
// // attacks. See https://pkg.go.dev/golang.org/x/oauth2#Config.AuthCodeURL
// var stateMap = map[string]string{}

// this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
var sessions = map[string]*session{}

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

// each session contains the username of the user and the time at which it expires
type session struct {
	id       int
	username string
	state    string
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
		http.ServeFile(w, r, "home.html")
		return
	}

	// authenticate user
	state := r.FormValue("state")
	code := r.FormValue("code")
	if state == "" || state != session.state {
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
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.id = j.Id
	session.username = j.Slug

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s session) isAuthenticated() bool {
	return s.id != 0
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
	log.Println(r.URL)
	if r.URL.Path != "/tile" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: authentication

	defer r.Body.Close()
	type jsonBody struct {
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Color string `json:"color"`
	}
	var j jsonBody
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	hub.broadcast <- []byte(fmt.Sprintf("%d %d %s", j.X, j.Y, j.Color))
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

	// Create a new random session token
	sessionToken := uuid.NewString()

	// Set the token in the session map, along with the session information
	sessions[sessionToken] = &session{
		state: uuid.NewString(),
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	url := conf.AuthCodeURL(sessions[sessionToken].state, oauth2.AccessTypeOnline)
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
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// generateRandomString generates a random hex string
func generateRandomString() (string, error) {
	n := 8
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := fmt.Sprintf("%x", b)
	return s, nil
}
