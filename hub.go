package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// boardSize is the width and height of the board, matching the table
// size in home.html
const boardSize = 100

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *InternalMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// board is an in-memory representation of the board
	// where each entry is a javascript color
	board [][]int

	// tileInfoBoard is an in-memory represenation of the board
	// where each each respresents metadata for a tile
	tileInfoBoard [][]TileInfo
}

type TileInfo struct {
	User       User
	LastUpdate time.Time
}

func newHub() *Hub {
	hub := &Hub{
		broadcast:     make(chan *InternalMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		board:         make([][]int, boardSize),
		tileInfoBoard: make([][]TileInfo, boardSize),
	}

	bytes, err := redisClient.Get(context.Background(), os.Getenv("REDIS_BOARD_KEY")).Bytes()
	if err != nil {
		log.Println(err)
		if err != redis.Nil {
			panic(err)
		}
		// initialize the bitfield
		for i := 0; i < boardSize*boardSize; i++ {
			if err := redisClient.BitField(context.Background(),
				os.Getenv("REDIS_BOARD_KEY"),
				"SET",
				"u4",
				fmt.Sprintf("#%d", i),
				5, // 5 is cornflowerblue
			).Err(); err != nil {
				panic(err)
			}
		}
		bytes, err = redisClient.Get(context.Background(), os.Getenv("REDIS_BOARD_KEY")).Bytes()
		if err != nil {
			panic(err)
		}
	}

	// initialize boards
	for i := 0; i < boardSize; i++ {
		hub.board[i] = make([]int, boardSize)
		hub.tileInfoBoard[i] = make([]TileInfo, boardSize)
	}

	for i := 0; i < len(bytes); i++ {
		firstColor, secondColor := getColorsFromByte(bytes[i])

		hub.board[i/(boardSize/2)][2*(i%(boardSize/2))] = firstColor
		hub.board[i/(boardSize/2)][2*(i%(boardSize/2))+1] = secondColor
	}

	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// parse and set color in memory
			webSocketMessage, err := h.parseAndSave(*message)
			if err != nil {
				log.Println(err)
			}
			for client := range h.clients {
				select {
				case client.send <- webSocketMessage:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// parseAndSave parses a message into x, y, and color and saves it to
// the board
func (h *Hub) parseAndSave(message InternalMessage) ([]byte, error) {
	h.board[message.Y][message.X] = message.Color
	h.tileInfoBoard[message.Y][message.X] = TileInfo{User: message.User, LastUpdate: message.Timestamp}

	offset := message.Y*boardSize + message.X
	_, err := redisClient.BitField(context.Background(), os.Getenv("REDIS_BOARD_KEY"), "SET", "u4", fmt.Sprintf("#%d", offset), strconv.Itoa(message.Color)).Result()
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("%d %d %d\n", message.X, message.Y, message.Color)), nil
}

func isInBounds(x, y int) error {
	if y < 0 || x < 0 || y >= boardSize || x >= boardSize {
		return errors.New("out of bounds")
	}
	return nil
}
