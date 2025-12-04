package services

import (
	"context"
	"log"
	"sync"

	"gobackend/config"
	"gobackend/models"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	WsClients   = make(map[string]map[*websocket.Conn]bool)
	WsMutex     sync.RWMutex
	WsCloseOnce = make(map[*websocket.Conn]bool)
)

func RegisterClient(country string, conn *websocket.Conn) {
	WsMutex.Lock()
	if WsClients[country] == nil {
		WsClients[country] = make(map[*websocket.Conn]bool)
	}
	WsClients[country][conn] = true
	WsMutex.Unlock()
}

func UnregisterClient(country string, conn *websocket.Conn) {
	WsMutex.Lock()
	delete(WsClients[country], conn)
	WsMutex.Unlock()
}

func SendCollegesUpdate(country string, conn *websocket.Conn) {
	cursor, err := config.CollegeCollection.Find(context.TODO(), bson.M{
		"country": bson.M{"$regex": "^" + country + "$", "$options": "i"},
	})

	if err != nil {
		log.Printf(" Error fetching colleges: %v", err)
		return
	}
	defer cursor.Close(context.TODO())

	var colleges []map[string]interface{}
	for cursor.Next(context.TODO()) {
		var college models.CollegeStats
		if err := cursor.Decode(&college); err == nil {
			colleges = append(colleges, map[string]interface{}{
				"id":      college.CollegeName,
				"name":    college.CollegeName,
				"country": country,
				"data":    college.StudentStatistics,
			})
		}
	}

	message := map[string]interface{}{
		"type":     "colleges_update",
		"colleges": colleges,
		"country":  country,
		"count":    len(colleges),
	}

	conn.WriteJSON(message)
}

func BroadcastNewCollege(country string, college map[string]interface{}) {
	WsMutex.RLock()
	clients := WsClients[country]
	WsMutex.RUnlock()

	if len(clients) == 0 {
		return
	}

	message := map[string]interface{}{
		"type":    "new_college",
		"college": college,
		"country": country,
	}

	WsMutex.Lock()
	for client := range clients {
		client.WriteJSON(message)
	}
	WsMutex.Unlock()
}
