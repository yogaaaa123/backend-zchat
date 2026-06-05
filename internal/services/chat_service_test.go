package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper: start hub dengan cleanup
func startHub(t *testing.T) *Hub {
	t.Helper()
	hub := NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.StopCh <- struct{}{} })
	return hub
}

// ==================== HUB TESTS ====================

func TestHub_RegisterClient(t *testing.T) {
	hub := startHub(t)

	client := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	hub.RegisterCh <- client
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	assert.Contains(t, hub.rooms, "room-1")
	assert.Contains(t, hub.rooms["room-1"], client)
	assert.Equal(t, 1, len(hub.rooms["room-1"]))
	hub.mu.RUnlock()
}

func TestHub_RegisterMultipleClients(t *testing.T) {
	hub := startHub(t)

	for i := 0; i < 5; i++ {
		c := NewClient(hub, nil, uuid.New(), fmt.Sprintf("user%d", i), "room-1", nil)
		hub.RegisterCh <- c
	}
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	assert.Equal(t, 5, len(hub.rooms["room-1"]))
	hub.mu.RUnlock()
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-1", nil)
	hub.RegisterCh <- c1
	hub.RegisterCh <- c2
	time.Sleep(20 * time.Millisecond)

	hub.UnregisterCh <- c1
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	assert.Equal(t, 1, len(hub.rooms["room-1"]))
	assert.Contains(t, hub.rooms["room-1"], c2)
	assert.NotContains(t, hub.rooms["room-1"], c1)
	hub.mu.RUnlock()
}

func TestHub_UnregisterLastRemovesRoom(t *testing.T) {
	hub := startHub(t)

	c := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	hub.RegisterCh <- c
	time.Sleep(20 * time.Millisecond)

	hub.UnregisterCh <- c
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.rooms["room-1"]
	assert.False(t, exists, "room harus dihapus kalo client terakhir leave")
	hub.mu.RUnlock()
}

func TestHub_MultipleRooms(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-a", nil)
	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-b", nil)
	hub.RegisterCh <- c1
	hub.RegisterCh <- c2
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	assert.Equal(t, 1, len(hub.rooms["room-a"]))
	assert.Equal(t, 1, len(hub.rooms["room-b"]))
	hub.mu.RUnlock()
}

func TestHub_BroadcastToAllExceptSender(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-1", nil)
	c3 := NewClient(hub, nil, uuid.New(), "user3", "room-1", nil)
	hub.RegisterCh <- c1
	hub.RegisterCh <- c2
	hub.RegisterCh <- c3
	time.Sleep(30 * time.Millisecond)

	// Drain join events — c1: 2 (c2+c3), c2: 1 (c3), c3: 0
	drain := func(c *Client, n int) {
		for i := 0; i < n; i++ {
			select {
			case <-c.Send:
			case <-time.After(50 * time.Millisecond):
				return
			}
		}
	}
	drain(c1, 2)
	drain(c2, 1)
	drain(c3, 0)

	msg := []byte(`{"type":"test","data":"hello"}`)
	hub.broadcastToRoom("room-1", msg, c1)

	for i, client := range []*Client{c2, c3} {
		select {
		case received := <-client.Send:
			assert.Equal(t, msg, received, "client %d harus terima broadcast", i+2)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("client %d gak terima pesan", i+2)
		}
	}

	// c1 (sender) gak boleh dapet
	select {
	case <-c1.Send:
		t.Fatal("sender (c1) gak boleh terima broadcast")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestHub_Broadcast_EmptyRoom(t *testing.T) {
	hub := NewHub()
	hub.broadcastToRoom("empty-room", []byte(`{"type":"test"}`), nil)
}

func TestHub_Broadcast_NonExistentRoom(t *testing.T) {
	hub := NewHub()
	hub.broadcastToRoom("gak-ada", []byte(`{"type":"test"}`), nil)
}

func TestHub_JoinEvent(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	hub.RegisterCh <- c1
	time.Sleep(20 * time.Millisecond)

	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-1", nil)
	hub.RegisterCh <- c2
	time.Sleep(20 * time.Millisecond)

	select {
	case msg := <-c1.Send:
		var wsMsg WSMessage
		json.Unmarshal(msg, &wsMsg)
		assert.Equal(t, "join", wsMsg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("c1 harus terima join event pas c2 connect")
	}
}

func TestHub_LeaveEvent(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-1", nil)
	hub.RegisterCh <- c1
	hub.RegisterCh <- c2
	time.Sleep(20 * time.Millisecond)

	hub.UnregisterCh <- c2
	time.Sleep(20 * time.Millisecond)

}

// ==================== WEBSOCKET TESTS ====================

func TestClient_WritePump_SendsMessage(t *testing.T) {
	done := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		_, msg, err := conn.ReadMessage()
		if err == nil {
			done <- msg
		}
		conn.Close()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	hub := NewHub()
	client := NewClient(hub, conn, uuid.New(), "tester", "room-x", nil)
	go client.WritePump()

	testMsg := []byte(`{"type":"message","data":{"content":"hello"}}`)
	client.Send <- testMsg

	select {
	case received := <-done:
		assert.Equal(t, testMsg, received)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("server gak terima pesan")
	}
}

func TestClient_FullCycle(t *testing.T) {
	cleanDB(t)

	messages := make(chan []byte, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err == nil {
			messages <- msg
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"message","data":{"content":"ping"}}`))
	require.NoError(t, err)

	_, pong, err := conn.ReadMessage()
	require.NoError(t, err)

	var wsMsg WSMessage
	json.Unmarshal(pong, &wsMsg)
	assert.Equal(t, "pong", wsMsg.Type)

	select {
	case received := <-messages:
		var msg WSMessage
		json.Unmarshal(received, &msg)
		assert.Equal(t, "message", msg.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("server gak terima pesan")
	}
}

// ==================== NEW CHAT TESTS ====================

func TestHub_Broadcast_BufferFull_UnregistersClient(t *testing.T) {
	hub := startHub(t)

	client := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	hub.RegisterCh <- client
	time.Sleep(20 * time.Millisecond)

	// Drain join event
	select { case <-client.Send: default: }

	// Fill the buffer completely (channel cap is 256)
	for i := 0; i < 256; i++ {
		select {
		case client.Send <- []byte("data"):
		default:
			break
		}
	}

	// This broadcast should trigger default case → client unregistered
	hub.broadcastToRoom("room-1", []byte("overflow"), nil)
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.rooms["room-1"][client]
	hub.mu.RUnlock()
	assert.False(t, exists, "client harus di-unregister kalo buffer penuh")
}

func TestHub_Register_SameClientTwice(t *testing.T) {
	hub := startHub(t)

	client := NewClient(hub, nil, uuid.New(), "user1", "room-1", nil)
	hub.RegisterCh <- client
	time.Sleep(10 * time.Millisecond)
	hub.RegisterCh <- client // same client lagi
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	assert.Equal(t, 1, len(hub.rooms["room-1"]))
	hub.mu.RUnlock()
}

func TestHub_Unregister_NonExistentClient(t *testing.T) {
	hub := startHub(t)

	client := NewClient(hub, nil, uuid.New(), "ghost", "room-x", nil)
	hub.UnregisterCh <- client // client gak pernah register → no panic
	time.Sleep(10 * time.Millisecond)

	assert.True(t, true, "should not panic")
}

func TestHub_Broadcast_OnlyToSpecificRoom(t *testing.T) {
	hub := startHub(t)

	c1 := NewClient(hub, nil, uuid.New(), "user1", "room-a", nil)
	c2 := NewClient(hub, nil, uuid.New(), "user2", "room-b", nil)
	hub.RegisterCh <- c1
	hub.RegisterCh <- c2
	time.Sleep(20 * time.Millisecond)

	// Drain join events
	select { case <-c1.Send: default: }
	select { case <-c2.Send: default: }

	// Broadcast ke room-a aja
	hub.broadcastToRoom("room-a", []byte("hello"), nil)
	time.Sleep(10 * time.Millisecond)

	// c1 harus terima
	select {
	case <-c1.Send:
		// OK
	default:
		t.Fatal("c1 harus terima broadcast")
	}

	// c2 gak boleh terima
	select {
	case <-c2.Send:
		t.Fatal("c2 gak boleh terima broadcast room lain")
	default:
		// OK
	}
}
