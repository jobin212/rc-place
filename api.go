package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type tileResponse struct {
	Color string `json:"color"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*User{}

// serveTile serves the '/tile' API route for programatically updating a tile.
func serveTile(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		updateTile(hub, w, r)
	} else {
		getTile(hub, w, r)
	}
}

func getTile(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/tile") {
		return
	}

	// authenticate
	_, err := authPersonalAccessToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	x, err := strconv.Atoi(query.Get("x"))
	y, err := strconv.Atoi(query.Get("y"))

	if err != nil {
		log.Println("Missing or malformed query parameter")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err = isInBounds(x, y); err != nil {
		log.Println("Index out of bounds")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	color := hub.board[y][x]

	tile := tileResponse{Color: colorToName[color], X: x, Y: y}
	resp, err := json.Marshal(tile)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return
}

func updateTile(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/tile") {
		return
	}
	// TODO: respond with JSON bodies always

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
		if err.Error() == "unknown color" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		} else {
			http.Error(w, "Too Early", http.StatusTooEarly)
		}
		return
	}
}

func getTiles(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/tiles") {
		return
	}

	// authenticate
	_, err := authPersonalAccessToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type tilesResponse struct {
		Tiles  [][]int `json:"tiles"`
		Height int     `json:"height"`
		Width  int     `json:"width"`
	}

	board := tilesResponse{Tiles: hub.board, Height: boardSize, Width: boardSize}
	resp, err := json.Marshal(board)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return
}

// authPersonalAccessToken will authenticate an Authorization header by
// forwarding a request to recurse.com API and cache a successful result
// in pacCache.
func authPersonalAccessToken(r *http.Request) (*User, error) {
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
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// update cache
	pacCache[pacToken] = &user
	return pacCache[pacToken], nil
}
