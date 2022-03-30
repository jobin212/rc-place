package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
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
		"PG_DATABASE_URL",
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
		log.Println("Error setting up postgres:", err)
		os.Exit(1) // TODO make postgres not required
	}
	fmt.Println("Client success")
	defer postgresClient.Close()

	// update postgres
	res, err := postgresClient.Exec(
		"INSERT INTO tile_info(username, x, y, color) VALUES ($1, $2, $3, $4) ON CONFLICT (x, y) DO UPDATE SET username=excluded.username, timestamp=now(), color=excluded.color",
		"user123",
		10,
		11,
		13)
	if err != nil {
		log.Println(err)
	}

	log.Println(res)

	var timestamp time.Time
	var username string
	err = postgresClient.QueryRow("SELECT username, timestamp FROM tile_info WHERE x = $1 AND y = $2", 10, 11).Scan(&username, &timestamp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	log.Printf("%s %v\n", username, timestamp)

	// // setup redis connection
	// if err := setupRedisClient(); err != nil {
	// 	log.Println("Error setting up redis:", err)
	// 	os.Exit(1)
	// }

	// hub := newHub()
	// go hub.run()
	// http.HandleFunc("/", serveHome)
	// http.HandleFunc("/login", serveLogin)
	// http.HandleFunc("/auth", serveAuth)
	// http.HandleFunc("/tile", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTile(hub, w, r)
	// })
	// http.HandleFunc("/tiles", func(w http.ResponseWriter, r *http.Request) {
	// 	getTiles(hub, w, r)
	// })
	// http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
	// 	serveFavicon(hub, w, r)
	// })
	// http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	session, err := getSession(r)
	// 	if err != nil {
	// 		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	// 		return
	// 	}
	// 	serveWs(hub, &session.User, w, r)
	// })
	// log.Printf("Running on port %s\n", *addr)

	// err := http.ListenAndServe(*addr, nil)
	// if err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
}
