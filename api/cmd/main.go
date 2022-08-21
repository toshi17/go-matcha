package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.CloseHandler()
	for {
		userId := c.Query("userId")
		log.Println("userID:" + userId)
		mt, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		if string(message) == "ping" {
			message = []byte("pong")
		}
		err = conn.WriteMessage(mt, message)
		if err != nil {
			fmt.Println(err)
			break
		}
	}

}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/ws", wshandler)
	r.Run()
}
