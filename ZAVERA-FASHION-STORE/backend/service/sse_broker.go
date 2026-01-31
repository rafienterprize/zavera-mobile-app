package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	AdminID  int
	Username string
	Events   chan SSEEvent
	ctx      context.Context
	cancel   context.CancelFunc
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID    string      `json:"id"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
	Retry int         `json:"retry,omitempty"` // Milliseconds
}

// SSEBroker manages SSE client connections and event broadcasting
type SSEBroker struct {
	clients    map[string]*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan AdminNotification
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

var (
	// Global SSE broker instance
	ssebroker     *SSEBroker
	brokerOnce    sync.Once
	brokerStarted bool
)

// GetSSEBroker returns the singleton SSE broker instance
func GetSSEBroker() *SSEBroker {
	brokerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		ssebroker = &SSEBroker{
			clients:    make(map[string]*SSEClient),
			register:   make(chan *SSEClient, 10),
			unregister: make(chan *SSEClient, 10),
			broadcast:  make(chan AdminNotification, 256),
			ctx:        ctx,
			cancel:     cancel,
		}
	})
	return ssebroker
}

// Start begins the broker's event loop
func (b *SSEBroker) Start() {
	if brokerStarted {
		log.Println("⚠️ SSE Broker already started, skipping")
		return
	}
	brokerStarted = true
	
	log.Println("🚀 SSE Broker starting...")
	
	// Start the main event loop
	go b.run()
	
	// NOTE: We no longer listen to NotificationChannel
	// Notifications are sent directly to broker.broadcast via BroadcastNotification()
	
	// Send test notification after 3 seconds to verify broker works
	go func() {
		time.Sleep(3 * time.Second)
		log.Println("🧪 Sending test notification to verify broker...")
		BroadcastNotification(AdminNotification{
			ID:       fmt.Sprintf("test-startup-%d", time.Now().Unix()),
			Type:     "system",
			Title:    "🧪 System Test",
			Message:  "SSE Broker is running!",
			Severity: SeverityInfo,
			Data:     map[string]interface{}{"test": true},
			Timestamp: time.Now(),
			Read:     false,
		})
	}()
	
	log.Println("✅ SSE Broker started successfully")
}

// listenToNotifications bridges the old NotificationChannel to the new broker
func (b *SSEBroker) listenToNotifications() {
	log.Println("📡 SSE Broker: Listening to NotificationChannel...")
	log.Printf("📡 SSE Broker: NotificationChannel address: %p", NotificationChannel)
	
	for {
		select {
		case <-b.ctx.Done():
			log.Println("📡 SSE Broker: Stopping notification listener")
			return
		case notif := <-NotificationChannel:
			log.Printf("📡 SSE Broker: Received notification from channel: type=%s", notif.Type)
			// Forward to broker's broadcast channel
			select {
			case b.broadcast <- notif:
				log.Printf("📡 SSE Broker: Forwarded to broadcast channel")
			default:
				log.Println("⚠️ SSE Broker broadcast channel full, dropping notification")
			}
		}
	}
}

// run is the main event loop for the broker
func (b *SSEBroker) run() {
	for {
		select {
		case <-b.ctx.Done():
			log.Println("🛑 SSE Broker shutting down")
			b.closeAllClients()
			return

		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.ID] = client
			count := len(b.clients)
			b.mu.Unlock()
			log.Printf("✅ SSE client connected: %s (admin: %s, total: %d)", client.ID, client.Username, count)

		case client := <-b.unregister:
			b.mu.Lock()
			if _, exists := b.clients[client.ID]; exists {
				delete(b.clients, client.ID)
				close(client.Events)
				count := len(b.clients)
				b.mu.Unlock()
				log.Printf("❌ SSE client disconnected: %s (admin: %s, total: %d)", client.ID, client.Username, count)
			} else {
				b.mu.Unlock()
			}

		case notification := <-b.broadcast:
			log.Printf("📢 SSE Broker: Broadcasting notification to %d clients: type=%s", len(b.clients), notification.Type)
			
			// Convert AdminNotification to SSEEvent
			event := b.notificationToSSEEvent(notification)
			
			// Broadcast to all connected clients
			b.mu.RLock()
			sentCount := 0
			for _, client := range b.clients {
				select {
				case client.Events <- event:
					sentCount++
				case <-time.After(100 * time.Millisecond):
					// Client is slow/blocked, skip this event
					log.Printf("⚠️ Slow client detected: %s, skipping event", client.ID)
				}
			}
			b.mu.RUnlock()
			
			log.Printf("📢 Broadcast complete: sent to %d/%d clients", sentCount, len(b.clients))
		}
	}
}

// RegisterClient adds a new SSE client to the broker
func (b *SSEBroker) RegisterClient(client *SSEClient) {
	b.register <- client
}

// UnregisterClient removes an SSE client from the broker
func (b *SSEBroker) UnregisterClient(client *SSEClient) {
	b.unregister <- client
}

// NewClient creates a new SSE client
func (b *SSEBroker) NewClient(adminID int, username string) *SSEClient {
	ctx, cancel := context.WithCancel(b.ctx)
	
	client := &SSEClient{
		ID:       generateClientID(),
		AdminID:  adminID,
		Username: username,
		Events:   make(chan SSEEvent, 100), // Buffer to prevent blocking
		ctx:      ctx,
		cancel:   cancel,
	}
	
	return client
}

// notificationToSSEEvent converts AdminNotification to SSEEvent
func (b *SSEBroker) notificationToSSEEvent(notif AdminNotification) SSEEvent {
	return SSEEvent{
		ID:    notif.ID,
		Event: "notification",
		Data:  notif,
		Retry: 3000, // Retry after 3 seconds if connection drops
	}
}

// closeAllClients closes all connected clients
func (b *SSEBroker) closeAllClients() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for _, client := range b.clients {
		client.cancel()
		close(client.Events)
	}
	b.clients = make(map[string]*SSEClient)
}

// Shutdown gracefully shuts down the broker
func (b *SSEBroker) Shutdown() {
	log.Println("🛑 Shutting down SSE Broker...")
	b.cancel()
}

// GetClientCount returns the number of connected clients
func (b *SSEBroker) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("sse-%d-%d", time.Now().UnixNano(), time.Now().Unix()%10000)
}

// FormatSSEMessage formats an SSE event as per the SSE specification
func FormatSSEMessage(event SSEEvent) string {
	var msg string
	
	// Event ID (for Last-Event-ID support)
	if event.ID != "" {
		msg += fmt.Sprintf("id: %s\n", event.ID)
	}
	
	// Event type
	if event.Event != "" {
		msg += fmt.Sprintf("event: %s\n", event.Event)
	}
	
	// Retry interval
	if event.Retry > 0 {
		msg += fmt.Sprintf("retry: %d\n", event.Retry)
	}
	
	// Data (JSON encoded)
	if event.Data != nil {
		dataJSON, err := json.Marshal(event.Data)
		if err != nil {
			log.Printf("❌ Failed to marshal SSE data: %v", err)
			dataJSON = []byte(`{"error":"failed to encode data"}`)
		}
		msg += fmt.Sprintf("data: %s\n", string(dataJSON))
	}
	
	// SSE messages must end with double newline
	msg += "\n"
	
	return msg
}

// SendKeepAlive sends a comment to keep the connection alive
func SendKeepAlive() string {
	return fmt.Sprintf(": keepalive %d\n\n", time.Now().Unix())
}
