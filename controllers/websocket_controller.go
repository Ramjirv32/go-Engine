package controllers

import (
	"log"
	"net/http"
	"time"

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
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("🔌 WebSocket client connected for country: %s", country)

	// Set connection parameters for stability
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	services.RegisterClient(country, conn)
	services.SendCollegesUpdate(country, conn)

	// Heartbeat ticker
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send ping to keep connection alive
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("❌ WebSocket ping error for %s: %v", country, err)
				goto disconnect
			}

		default:
			// Read message with timeout
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					log.Printf("❌ WebSocket error for %s: %v", country, err)
				}
				goto disconnect
			}
		}
	}

disconnect:
	services.UnregisterClient(country, conn)
	log.Printf("🔌 WebSocket client disconnected for country: %s", country)
}

func HandleWebSocketCountries(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("🔌 WebSocket client connected for countries updates")

	// Set connection parameters for stability
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Send initial countries list
	services.SendCountriesUpdate(conn)

	// Heartbeat ticker
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	// Keep connection open for updates
	for {
		select {
		case <-ticker.C:
			// Send ping to keep connection alive
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("❌ WebSocket ping error for countries: %v", err)
				goto disconnect2
			}

		default:
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					log.Printf("❌ WebSocket error for countries: %v", err)
				}
				goto disconnect2
			}
		}
	}

disconnect2:
	log.Printf("🔌 WebSocket client disconnected for countries")
}
