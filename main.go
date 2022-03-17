// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

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

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
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
