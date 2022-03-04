package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type UserResponse struct {
	Action    string          `json:"action"`
	Message   string          `json:"message"`
	Conn      *websocket.Conn `json:"-"`
	ChannelId uint64          `json:"-"`
}
type Room struct {
	ChannelId   uint64
	Connections map[*websocket.Conn]client
}

type UserConnection struct {
	Connection *websocket.Conn
	ChannelId  uint64
}

type client struct{} // Add more data to this type if needed

var rooms = make(map[uint64]Room)
var register = make(chan *UserConnection)
var unregister = make(chan *UserConnection)
var broadcast = make(chan UserResponse)

func runHub() {
	for {
		select {
		//https://stackoverflow.com/questions/42605337/cannot-assign-to-struct-field-in-a-map
		case connection := <-register:
			if entry, ok := rooms[connection.ChannelId]; ok {
				entry.Connections[connection.Connection] = client{}
				rooms[connection.ChannelId] = entry
				log.Println("Connection registered to existing room")

			} else {
				entry.ChannelId = connection.ChannelId
				if entry.Connections == nil {
					entry.Connections = make(map[*websocket.Conn]client)
				}
				entry.Connections[connection.Connection] = client{}
				rooms[connection.ChannelId] = entry

				log.Println("Connection registered to new room")
			}

		case message := <-broadcast:
			log.Println("message received:", message)
			for connection := range rooms[message.ChannelId].Connections {
				if err := connection.WriteJSON(message); err != nil {
					_ = connection.WriteMessage(websocket.CloseMessage, []byte{})
					_ = connection.Close()
					delete(rooms[message.ChannelId].Connections, connection)
				}

			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(rooms[connection.ChannelId].Connections, connection.Connection)
			log.Println("connection unregistered")
		}
	}
}

func main() {
	app := fiber.New()

	app.Use(logger.New())

	app.Static("/", "./static/home.html")

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	go runHub()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// When the function returns, unregister the client and close the connection
		fmt.Println("output", c.Params("id", "?"))

		channelId, err := strconv.ParseUint(c.Params("id", "0"), 10, 64)
		if err != nil {
			return
		}
		userConn := &UserConnection{
			Connection: c,
			ChannelId:  channelId,
		}
		// when we exit the function we'll remove these from continuing the broadcasted to
		defer func() {
			unregister <- userConn
			c.Close()
		}()

		register <- userConn

		response := UserResponse{
			Conn:      c,
			ChannelId: channelId,
		}
		for {

			err := c.ReadJSON(&response)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}

				return // Calls the deferred function, i.e. closes the connection on error
			}

			broadcast <- response
		}
	}))

	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	log.Fatal(app.Listen(*addr))
}
