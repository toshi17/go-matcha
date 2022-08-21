package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
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

const roomSize = 3

type Match struct {
	RoomId string   `json:"room_id"`
	Users  []string `json:"users"`
}

func wshandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.CloseHandler()
	uId := c.Query("userId")
	log.Println("userID:" + uId)
	qId := c.Query("queueId")
	log.Println("queueId" + qId)
	qKey := "queue" + qId
	dbconn, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbconn.Close()

	// ユーザーのキューイング
	r, err := dbconn.Do("LPUSH", qKey, uId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)

	// マッチング
	l, err := redis.Int(dbconn.Do("LLEN", qKey))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("%s room size: %d", qKey, l)

	if l >= roomSize {
		us := []string{}
		for i := 0; i < roomSize; i++ {
			u, err := redis.String(dbconn.Do("RPOP", qKey))
			if err != nil {
				fmt.Println(err)
				return
			}
			us = append(us, u)
		}

		rid, err := uuid.NewRandom()
		if err != nil {
			fmt.Println(err)
			return
		}
		var match = Match{
			RoomId: rid.String(),
			Users:  us,
		}
		m, err := json.Marshal(match)
		if err != nil {
			fmt.Println(err)
			return
		}
		dbconn.Do("PUBLISH", "matches", string(m))
		conn.WriteMessage(websocket.TextMessage, m)
		return
	}

	//　mathesのサブスクライブ
	go func() {
		subConn, err := redis.Dial("tcp", "redis:6379")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer subConn.Close()

		psc := redis.PubSubConn{Conn: subConn}
		psc.Subscribe("matches")
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				m := new(Match)
				err := json.Unmarshal(v.Data, m)
				if err != nil {
					fmt.Println(err)
					return
				}
				for _, u := range m.Users {
					if u == uId {
						fmt.Println("match made!! : %s", m.Users)
						conn.WriteMessage(websocket.TextMessage, v.Data)
						if err != nil {
							fmt.Println(err)
						}
						return
					}
				}
			case redis.Subscription:
				fmt.Printf("subscribed %s: %s", v.Channel, uId)
			case error:
				fmt.Println(err)
				return
			}
		}
	}()
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
