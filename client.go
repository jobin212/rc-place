package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Time before a user can update the canvas again.
	updateLimit = 10 * time.Millisecond
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}

	lastUpdateCache = map[string]time.Time{}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	nameToColor = map[string]int{
		"black":          0,
		"forest":         1,
		"green":          2,
		"lime":           3,
		"blue":           4,
		"cornflowerblue": 5,
		"sky":            6,
		"cyan":           7,
		"red":            8,
		"burnt-orange":   9,
		"orange":         10,
		"yellow":         11,
		"purple":         12,
		"hot-pink":       13,
		"pink":           14,
		"white":          15,
	}

	colorToName = map[int]string{
		0:  "black",
		1:  "forest",
		2:  "green",
		3:  "lime",
		4:  "blue",
		5:  "cornflowerblue",
		6:  "sky",
		7:  "cyan",
		8:  "red",
		9:  "burnt-orange",
		10: "orange",
		11: "yellow",
		12: "purple",
		13: "hot-pink",
		14: "pink",
		15: "white",
	}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// User this websocket is associated with.
	user *User
}

type User struct {
	Id       int    `json:"id"`
	Username string `json:"slug"`
}

func (u *User) SetTile(hub *Hub, x, y int, color string) error {
	if time.Since(lastUpdateCache[u.Username]) < updateLimit {
		return errors.New("rate limited")
	}
	// validate color
	colInt, ok := nameToColor[color]
	if !ok {
		return errors.New("unknown color")
	}

	internalMessage, err := createInternalMessage(fmt.Sprintf("%d %d %d", x, y, colInt), *u, time.Now())
	if err != nil {
		log.Println("Failed to createInternalMessage")
		return err
	}

	hub.broadcast <- internalMessage
	return nil
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, webSocketMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// TODO: structured commands
		webSocketMessage = bytes.TrimSpace(bytes.Replace(webSocketMessage, newline, space, -1))

		message := string(webSocketMessage)
		// Try to parse message and send board if so.
		if message == "getTiles" {
			if err = c.conn.WriteJSON(c.hub.board); err != nil {
				log.Println("uh oh, failed to send board:", err)
			}
			continue
		}

		// check if this user can send a message
		if time.Since(lastUpdateCache[c.user.Username]) < updateLimit {
			continue
		}

		internalMessage, err := createInternalMessage(message, *c.user, time.Now())
		if err != nil {
			log.Printf("Failed to createInternalMessage %v\n", err)
			continue
		}

		c.hub.broadcast <- internalMessage
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// send messages to initialize board state
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}
	for y := range c.hub.board {
		for x := range c.hub.board[y] {
			msg := fmt.Sprintf("%d %d %d\n", x, y, c.hub.board[y][x])
			w.Write([]byte(msg))
		}
	}
	if err := w.Close(); err != nil {
		return
	}

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, user *User, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, user: user, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func createInternalMessage(message string, user User, timestamp time.Time) (*InternalMessage, error) {
	parts := strings.Fields(message)

	if len(parts) < 3 {
		// do nothing if we don't have enough information from the message
		return nil, errors.New("malformed message")
	}

	x, y, c := parts[0], parts[1], parts[2]
	var xPos, yPos, color int
	var err error

	// convert to integers
	if xPos, err = strconv.Atoi(x); err != nil {
		return nil, err
	}
	if yPos, err = strconv.Atoi(y); err != nil {
		return nil, err
	}
	if color, err = strconv.Atoi(c); err != nil {
		return nil, err
	}

	// check bounds
	if err = isInBounds(xPos, yPos); err != nil {
		return nil, err
	}

	if _, ok := colorToName[color]; !ok {
		return nil, errors.New("unknown color int")
	}

	internalMessage := &InternalMessage{X: xPos, Y: yPos, Color: color, User: user, Timestamp: timestamp}

	return internalMessage, nil
}
