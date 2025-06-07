package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		CheckOrigin: checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	otps RetentionMap

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps: NewRetentionMap(ctx, 5*time.Second),
	}
	m.setupEventHandlers()
	return m
}



func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println(event)
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
		
	} else {
		return errors.New("there is no such event type")
	}
}



func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	otp := r.URL.Query().Get("otp")

	log.Println("OTP received:", otp)
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !m.otps.VerifyOTP(otp) {
		
		w.WriteHeader( http.StatusUnauthorized)
		return
	}
	log.Println("WebSocket connection established")
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return

	}

	client := NewClient(conn, m)
	m.addClient(client)

	go client.readMessages()

	go client.writeMessages()

}


func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request)  {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`	
	}

	var req userLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if req.Username == "admin" && req.Password == "123" {
		type response struct {
			OTP string `json:"otp"`
		}

		otp := m.otps.NewOTP()

		resp := response{
			OTP: otp.Key,
		}

		data, err := json.Marshal(resp)

		if err != nil {
			log.Println(err)
			return
		}


		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()

	defer m.Unlock()

	m.clients[client] = true
	log.Printf("Client added: %v, total clients: %d", client, len(m.clients))
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		delete(m.clients, client)
		log.Printf("Client removed: %v, total clients: %d", client, len(m.clients))
	}
}



func checkOrigin(r *http.Request) bool {
	// You can implement origin checking logic here if needed

	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
}
}