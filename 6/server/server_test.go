package main

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestTimeout(t *testing.T) {
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?job_id=test", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	// Wait longer than read timeout (simulate dead client)
	time.Sleep(70 * time.Second)
}
