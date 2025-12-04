package controllers

import (
	"log"
	"net/http"

	"gobackend/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocketColleges(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		http.Error(w, "country parameter required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(" WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("🔌 WebSocket client connected for country: %s", country)

	services.RegisterClient(country, conn)
	services.SendCollegesUpdate(country, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf(" WebSocket error: %v", err)
			}
			break
		}
	}

	services.UnregisterClient(country, conn)
	log.Printf("🔌 WebSocket client disconnected for country: %s", country)
}
