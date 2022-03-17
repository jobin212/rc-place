// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var addr = flag.String("addr", ":8080", "http service address")

// stateMap is used for oauth redirection to protect users from CSRF
// attacks. See https://pkg.go.dev/golang.org/x/oauth2#Config.AuthCodeURL
var stateMap = map[string]string{}

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
	// authenticate user
	state := r.FormValue("state")
	code := r.FormValue("code")
	if state == "" || state != stateMap[r.RemoteAddr] {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	if code == "" {
		// access denied
		http.Error(w, "Unoauthorized", http.StatusUnauthorized)
		return
	}
	// TODO: should we actually try to use the token to see if it's legitimate?
	http.ServeFile(w, r, "home.html")
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

	// generate random string
	state, err := generateRandomString()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// TODO: use cookie value instead of RemoteAddr
	stateMap[r.RemoteAddr] = state
	conf := &oauth2.Config{
		RedirectURL:  os.Getenv("OAUTH_REDIRECT"),
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://recurse.com/oauth/authorize",
			TokenURL: "https://recurse.com/oauth/token",
		},
	}
	url := conf.AuthCodeURL(stateMap[r.RemoteAddr], oauth2.AccessTypeOnline)
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
