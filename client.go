package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait = 10 * time.Second

	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

// Client represents a single WebSocket client connection.
// It holds the connection and a reference to the manager that handles it.
// The Client struct is used to manage individual WebSocket connections.
// It contains methods for sending messages, receiving messages, and closing the connection.
type Client struct {
	connection *websocket.Conn
	manager    *Manager

	//egress is used to avoid concurrrent writes to the WebSocket connection.

	egress chan Event
}


func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event), // Buffered channel to handle outgoing messages
	}
}

func (c *Client) readMessages() {

	defer func() {
		//cleanup connection
		c.manager.removeClient(c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
		return
	}

	c.connection.SetReadLimit(1024 * 1024) // Set a read limit of 1 MB
	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		var request Event 
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("Error unmarshalling mevent: %v", err)
			break
		}
	
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Printf("Error handling message: %v", err)
			
		}
	}
}


func (c *Client) pongHandler(pong string) error {
	log.Printf("Received pong from client: %s", c.connection.RemoteAddr())

	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for{
		select {
			case message, ok := <-c.egress:
				if !ok {
					if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
						log.Printf("Connection closed: %v", err)
					}

					return
				}

				data, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshalling message: %v", err)
					return
				}
				if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("Failed to send message: %v", err)
					
				}
				log.Printf("Sent message: %s", message)

			case <-ticker.C:
				log.Println("Sending ping to client:", c.connection.RemoteAddr())
				//send a ping to the client to keep the connection alive
				if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
					log.Printf("Failed to send ping: %v", err)
					return
				}
		}
	}
}

// SendMessage sends a message to the WebSocket client.
func (c *Client) SendMessage(message []byte) error {
	return c.connection.WriteMessage(websocket.TextMessage, message)
}

// ReceiveMessage receives a message from the WebSocket client.
func (c *Client) ReceiveMessage() ([]byte, error) {
	_, message, err := c.connection.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Close closes the WebSocket connection for the client.
func (c *Client) Close() error {
	err := c.connection.Close()
	if err != nil {
		return err
	}
	c.manager.removeClient(c)
	return nil
}
func (c *Client) String() string {
	return c.connection.RemoteAddr().String()
}
func (c *Client) GetConnection() *websocket.Conn {
	return c.connection
}
func (c *Client) GetManager() *Manager {
	return c.manager
}
func (c *Client) GetClients() ClientList {
	c.manager.RLock()
	defer c.manager.RUnlock()
	return c.manager.clients
}
func (c *Client) GetClientCount() int {
	c.manager.RLock()
	defer c.manager.RUnlock()
	return len(c.manager.clients)
}
func (c *Client) GetClientList() []string {
	c.manager.RLock()
	defer c.manager.RUnlock()

	clientList := make([]string, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		clientList = append(clientList, client.String())
	}
	return clientList
}
func (c *Client) GetClientByID(id string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == id {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByConnection(conn *websocket.Conn) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.GetConnection() == conn {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByManager(manager *Manager) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.GetManager() == manager {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByCount(count int) []string {
	c.manager.RLock()
	defer c.manager.RUnlock()

	clientList := make([]string, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		if len(client.GetClients()) == count {
			clientList = append(clientList, client.String())
		}
	}
	return clientList
}
func (c *Client) GetClientByName(name string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == name {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByAddress(address string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == address {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByPort(port string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == port {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByProtocol(protocol string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == protocol {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByStatus(status string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == status {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByType(clientType string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == clientType {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByVersion(version string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == version {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByOS(os string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == os {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByLocation(location string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == location {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByTime(time string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == time {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByIDAndConnection(id string, conn *websocket.Conn) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == id && client.GetConnection() == conn {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByIDAndManager(id string, manager *Manager) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == id && client.GetManager() == manager {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByConnectionAndManager(conn *websocket.Conn, manager *Manager) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.GetConnection() == conn && client.GetManager() == manager {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByIDAndConnectionAndManager(id string, conn *websocket.Conn, manager *Manager) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.String() == id && client.GetConnection() == conn && client.GetManager() == manager {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByIDAndCount(id string, count int) []string {
	c.manager.RLock()
	defer c.manager.RUnlock()

	clientList := make([]string, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		if client.String() == id && len(client.GetClients()) == count {
			clientList = append(clientList, client.String())
		}
	}
	return clientList
}
func (c *Client) GetClientByManagerAndCount(manager *Manager, count int) []string {
	c.manager.RLock()
	defer c.manager.RUnlock()

	clientList := make([]string, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		if client.GetManager() == manager && len(client.GetClients()) == count {
			clientList = append(clientList, client.String())
		}
	}
	return clientList
}
func (c *Client) GetClientByConnectionAndCount(conn *websocket.Conn, count int) []string {
	c.manager.RLock()
	defer c.manager.RUnlock()

	clientList := make([]string, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		if client.GetConnection() == conn && len(client.GetClients()) == count {
			clientList = append(clientList, client.String())
		}
	}
	return clientList
}
func (c *Client) GetClientByManagerAndConnection(manager *Manager, conn *websocket.Conn) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.GetManager() == manager && client.GetConnection() == conn {
			return client
		}
	}
	return nil
}
func (c *Client) GetClientByManagerAndID(manager *Manager, id string) *Client {
	c.manager.RLock()
	defer c.manager.RUnlock()

	for client := range c.manager.clients {
		if client.GetManager() == manager && client.String() == id {
			return client
		}
	}
	return nil
}
