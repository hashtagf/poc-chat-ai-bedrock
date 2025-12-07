package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type MessageRequest struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

type StreamChunk struct {
	Type     string                 `json:"type"`
	Content  string                 `json:"content,omitempty"`
	Citation map[string]interface{} `json:"citation,omitempty"`
	Error    map[string]interface{} `json:"error,omitempty"`
}

func main() {
	sessionID := flag.String("session", "", "Session ID")
	message := flag.String("message", "Hello, world!", "Message to send")
	flag.Parse()

	if *sessionID == "" {
		log.Fatal("Session ID is required. Use -session flag")
	}

	// Connect to WebSocket
	url := "ws://localhost:8080/api/chat/stream"
	log.Printf("Connecting to %s", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Handle interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	// Read messages
	go func() {
		defer close(done)
		for {
			var chunk StreamChunk
			err := conn.ReadJSON(&chunk)
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}

			switch chunk.Type {
			case "content":
				fmt.Print(chunk.Content)
			case "citation":
				fmt.Printf("\n[Citation: %v]\n", chunk.Citation)
			case "error":
				fmt.Printf("\n[Error: %v]\n", chunk.Error)
				return
			case "done":
				fmt.Println("\n[Done]")
				return
			}
		}
	}()

	// Send message
	req := MessageRequest{
		SessionID: *sessionID,
		Content:   *message,
	}

	log.Printf("Sending message: %s", *message)
	if err := conn.WriteJSON(req); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Wait for completion or interrupt
	select {
	case <-done:
		log.Println("Connection closed")
	case <-interrupt:
		log.Println("Interrupt received, closing connection")
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Write close error: %v", err)
		}
	}
}
