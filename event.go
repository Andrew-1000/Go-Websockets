package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `jsobn:"payload"`
}



// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type


type EventHandler func(event Event, c *Client) error



const (
	EventSendMessage = "send_message"
	// EventNewMessage is a response to send_message
	EventNewMessage = "new_message"
	// EventChangeRoom is event when switching rooms
	EventChangeRoom = "change_room"

	EventReceiveMessage = "receive_message"
	EventCloseConnection = "close_connection"
	EventError = "error"
	EventClientConnected = "client_connected"
	EventClientDisconnected = "client_disconnected"
	EventClientList = "client_list"
	EventClientListUpdate = "client_list_update"
	EventClientMessage = "client_message"
	EventServerMessage = "server_message"
	EventServerError = "server_error"
	EventServerStatus = "server_status"
	EventServerStatusUpdate = "server_status_update"
	EventServerConfig = "server_config"
	EventServerConfigUpdate = "server_config_update"
	EventServerCommand = "server_command"
	EventServerCommandResponse = "server_command_response"
	EventServerCommandError = "server_command_error"
	EventServerCommandStatus = "server_command_status"
	EventServerCommandStatusUpdate = "server_command_status_update"
	EventServerCommandList = "server_command_list"
	EventServerCommandListUpdate = "server_command_list_update"
	EventServerCommandExecution = "server_command_execution"
	EventServerCommandExecutionUpdate = "server_command_execution_update"
	EventServerCommandExecutionError = "server_command_execution_error"
	EventServerCommandExecutionStatus = "server_command_execution_status"
	EventServerCommandExecutionStatusUpdate = "server_command_execution_status_update"
	EventServerCommandExecutionList = "server_command_execution_list"
)





// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// NewMessageEvent is returned when responding to send_message
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}



// SendMessageHandler will send out a message to all other participants in the chat
func SendMessageHandler(event Event, c *Client) error {
	// Marshal Payload into wanted format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Prepare an Outgoing Message to others
	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now()
	broadMessage.Message = chatevent.Message
	broadMessage.From = chatevent.From

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage
	// Broadcast to all other Clients
	// for client := range c.manager.clients {
	// 	// Only send to clients inside the same chatroom
	// 	// if client.chatroom == c.chatroom {
	// 	// 	client.egress <- outgoingEvent
	// 	// }

	// }
	return nil
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

// ChatRoomHandler will handle switching of chatrooms between clients
// func ChatRoomHandler(event Event, c *Client) error {
// 	// Marshal Payload into wanted format
// 	var changeRoomEvent ChangeRoomEvent
// 	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
// 		return fmt.Errorf("bad payload in request: %v", err)
// 	}

// 	// Add Client to chat room
// 	c.chatroom = changeRoomEvent.Name

// 	return nil
// }