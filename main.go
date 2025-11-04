package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)
var mutex = &sync.Mutex{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()
	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
		broadcast <- message
	}
}
func handleMessages() {
	for {
		message := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				if err := client.Close(); err != nil {
					return
				}
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLFiles("index.html")
	// websocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	// serve HTML page
	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "Chat Room",
			"User":  "Aibar",
		})
	})
	return r
}
func main() {
	go handleMessages()
	fmt.Println("WebSocket server started on :8080")
	err := InitRouter().Run(":8080")
	if err != nil {
		return
	}

}
