package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	// check oauth environment variables
	abort := false
	for _, env := range []string{
		"OAUTH_REDIRECT",
		"OAUTH_CLIENT_ID",
		"OAUTH_CLIENT_SECRET",
		"REDIS_HOST",     // TODO: make optional; use map if not provided
		"REDIS_PASSWORD", // TODO: make optional; use map if not provided
		"REDIS_BOARD_KEY",
	} {
		if _, ok := os.LookupEnv(env); !ok {
			log.Println("Required environment variable missing:", env)
			abort = true
		}
	}
	if abort {
		log.Println("Aborting")
		os.Exit(1)
	}

	// setup postgres connection
	if err := setupPostgresConnection(); err != nil {
		log.Println("Error setting up postgres:", err) // TODO make postgres not required
	}
	defer postgresClient.Close()

	// setup redis connection
	if err := setupRedisClient(); err != nil {
		log.Println("Error setting up redis:", err)
		os.Exit(1)
	}

	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/auth", serveAuth)
	http.HandleFunc("/tile", func(w http.ResponseWriter, r *http.Request) {
		serveTile(hub, w, r)
	})
	http.HandleFunc("/tiles", func(w http.ResponseWriter, r *http.Request) {
		getTiles(hub, w, r)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		serveFavicon(hub, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		session, err := getSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		serveWs(hub, &session.User, w, r)
	})
	log.Printf("Running on port %s\n", *addr)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
