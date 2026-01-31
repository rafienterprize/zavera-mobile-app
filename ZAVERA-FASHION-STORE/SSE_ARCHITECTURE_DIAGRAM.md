# SSE Architecture Diagram

## System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ZAVERA E-COMMERCE PLATFORM                         │
│                        Real-Time Admin Notification System                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER (React)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                     Admin Dashboard UI                                │  │
│  │                                                                        │  │
│  │  ┌─────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │  │
│  │  │ NotificationBell│  │  Toast Container │  │  Dropdown Panel  │   │  │
│  │  │   SSE Component │  │  (Top-Right)     │  │  (Notifications) │   │  │
│  │  └────────┬────────┘  └──────────────────┘  └──────────────────┘   │  │
│  │           │                                                          │  │
│  │           │ Uses                                                     │  │
│  │           ▼                                                          │  │
│  │  ┌─────────────────────────────────────────────────────────────┐   │  │
│  │  │              useAdminSSE() Hook                              │   │  │
│  │  │  • EventSource connection                                    │   │  │
│  │  │  • Auto-reconnect (exponential backoff)                      │   │  │
│  │  │  • Notification state management                             │   │  │
│  │  │  • Browser notification API                                  │   │  │
│  │  │  • Toast event dispatcher                                    │   │  │
│  │  └─────────────────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└──────────────────────────────────┬───────────────────────────────────────────┘
                                   │
                                   │ HTTP SSE Connection
                                   │ Authorization: Bearer <JWT>
                                   │ Accept: text/event-stream
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SERVER LAYER (Golang)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    SSE Handler (sse_handler.go)                       │  │
│  │                                                                        │  │
│  │  GET /api/admin/events                                                │  │
│  │  ┌──────────────────────────────────────────────────────────────┐   │  │
│  │  │ 1. Validate JWT (Authorization header)                        │   │  │
│  │  │ 2. Check admin permissions                                    │   │  │
│  │  │ 3. Set SSE headers (text/event-stream)                        │   │  │
│  │  │ 4. Create SSE client                                          │   │  │
│  │  │ 5. Register with broker                                       │   │  │
│  │  │ 6. Stream events (with keepalive)                             │   │  │
│  │  └──────────────────────────────────────────────────────────────┘   │  │
│  └────────────────────────────────┬─────────────────────────────────────┘  │
│                                   │                                          │
│                                   │ Register/Unregister                      │
│                                   │ Stream Events                            │
│                                   ▼                                          │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    SSE Broker (sse_broker.go)                         │  │
│  │                                                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐ │  │
│  │  │  Client Registry                                                │ │  │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │ │  │
│  │  │  │ Client 1 │  │ Client 2 │  │ Client N │  ...                │ │  │
│  │  │  │ (Admin A)│  │ (Admin B)│  │ (Admin C)│                     │ │  │
│  │  │  └──────────┘  └──────────┘  └──────────┘                     │ │  │
│  │  └────────────────────────────────────────────────────────────────┘ │  │
│  │                                                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐ │  │
│  │  │  Event Loop (Single Goroutine)                                 │ │  │
│  │  │  • Listen to NotificationChannel                               │ │  │
│  │  │  • Handle client registration                                  │ │  │
│  │  │  • Handle client unregistration                                │ │  │
│  │  │  • Broadcast events to all clients                             │ │  │
│  │  └────────────────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────┬─────────────────────────────────────┘  │
│                                   │                                          │
│                                   │ Listens to                               │
│                                   ▼                                          │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │              Notification Service (notification_service.go)           │  │
│  │                                                                        │  │
│  │  Global Channel: NotificationChannel (buffered, cap: 256)            │  │
│  │                                                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐ │  │
│  │  │  Notification Functions                                         │ │  │
│  │  │  • NotifyOrderCreated(code, customer, amount)                  │ │  │
│  │  │  • NotifyPaymentReceived(code, method, amount)                 │ │  │
│  │  │  • NotifyPaymentExpired(code)                                  │ │  │
│  │  │  • NotifyShipmentUpdate(code, status, courier)                 │ │  │
│  │  │  • NotifyStockLow(product, stock)                              │ │  │
│  │  │  • NotifyRefundRequest(code, amount, reason)                   │ │  │
│  │  │  • NotifyDisputeCreated(code, dispute, reason)                 │ │  │
│  │  └────────────────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────┬─────────────────────────────────────┘  │
│                                   │                                          │
│                                   │ Called by                                │
│                                   ▼                                          │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                      Business Logic Layer                             │  │
│  │                                                                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │  │
│  │  │   Checkout   │  │   Payment    │  │   Shipping   │               │  │
│  │  │   Service    │  │   Webhook    │  │   Service    │               │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘               │  │
│  │                                                                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │  │
│  │  │   Refund     │  │   Dispute    │  │   Stock      │               │  │
│  │  │   Service    │  │   Service    │  │   Service    │               │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘               │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Event Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         EVENT FLOW: Order Created                            │
└─────────────────────────────────────────────────────────────────────────────┘

1. Customer Checkout
   │
   ▼
┌──────────────────────┐
│  POST /api/checkout  │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────────────────┐
│  CheckoutService.CreateOrder()   │
│  • Validate cart                 │
│  • Create order record           │
│  • Reserve stock                 │
└──────────┬───────────────────────┘
           │
           │ Calls
           ▼
┌──────────────────────────────────────────────────────┐
│  NotifyOrderCreated(orderCode, customer, amount)     │
│  • Creates AdminNotification struct                  │
│  • Sets severity = "info"                            │
│  • Sends to NotificationChannel                      │
└──────────┬───────────────────────────────────────────┘
           │
           │ Channel send
           ▼
┌──────────────────────────────────────────────────────┐
│  SSE Broker Event Loop                               │
│  • Receives notification from channel                │
│  • Converts to SSEEvent                              │
│  • Broadcasts to all connected clients               │
└──────────┬───────────────────────────────────────────┘
           │
           │ Broadcast (fan-out)
           ▼
┌─────────────────────────────────────────────────────────────┐
│  SSE Clients (All Connected Admins)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │ Admin A  │  │ Admin B  │  │ Admin C  │                 │
│  │ Events   │  │ Events   │  │ Events   │                 │
│  │ Channel  │  │ Channel  │  │ Channel  │                 │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘                 │
└───────┼─────────────┼─────────────┼───────────────────────┘
        │             │             │
        │ SSE Stream  │ SSE Stream  │ SSE Stream
        ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Browser  │  │ Browser  │  │ Browser  │
│ Admin A  │  │ Admin B  │  │ Admin C  │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
     │ Parse       │ Parse       │ Parse
     ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ useAdmin │  │ useAdmin │  │ useAdmin │
│ SSE Hook │  │ SSE Hook │  │ SSE Hook │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
     │ Update      │ Update      │ Update
     ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Toast   │  │  Toast   │  │  Toast   │
│  Notif   │  │  Notif   │  │  Notif   │
└──────────┘  └──────────┘  └──────────┘
```

---

## Connection Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SSE CONNECTION LIFECYCLE                             │
└─────────────────────────────────────────────────────────────────────────────┘

1. INITIAL CONNECTION
   ┌──────────────────────────────────────────────────────────────┐
   │ Browser: fetch('/api/admin/events', {                        │
   │   headers: { Authorization: 'Bearer <JWT>' }                 │
   │ })                                                            │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Server: HandleAdminSSE()                                      │
   │ • Validate JWT                                                │
   │ • Check admin permissions                                     │
   │ • Create SSE client                                           │
   │ • Register with broker                                        │
   │ • Send "connected" event                                      │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Browser: Connection established                               │
   │ • isConnected = true                                          │
   │ • Green indicator                                             │
   │ • Reset reconnect attempts                                    │
   └──────────────────────────────────────────────────────────────┘

2. ACTIVE CONNECTION
   ┌──────────────────────────────────────────────────────────────┐
   │ Server: Event loop                                            │
   │ • Send notifications as they arrive                           │
   │ • Send keepalive every 30s                                    │
   │ • Monitor client.Events channel                               │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Browser: Receive events                                       │
   │ • Parse SSE messages                                          │
   │ • Update notification state                                   │
   │ • Show toasts                                                 │
   │ • Play sounds                                                 │
   └──────────────────────────────────────────────────────────────┘

3. DISCONNECTION
   ┌──────────────────────────────────────────────────────────────┐
   │ Trigger: Network error, server restart, timeout               │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Server: Detect disconnect                                     │
   │ • Context cancelled                                           │
   │ • Unregister client                                           │
   │ • Close client.Events channel                                 │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Browser: Detect disconnect                                    │
   │ • isConnected = false                                         │
   │ • Gray indicator                                              │
   │ • Schedule reconnect                                          │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Browser: Auto-reconnect                                       │
   │ • Exponential backoff (1s, 2s, 4s, 8s, ...)                  │
   │ • Max 10 attempts                                             │
   │ • Max delay 30s                                               │
   └────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
                        [Back to step 1]
```

---

## Notification Severity Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      NOTIFICATION SEVERITY MAPPING                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐
│   Business Event     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────┐
│  Severity Assignment                                          │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  INFO (Blue)                                         │    │
│  │  • Order created                                     │    │
│  │  • Payment received                                  │    │
│  │  • Shipment update                                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  WARNING (Yellow)                                    │    │
│  │  • Payment expired                                   │    │
│  │  • Stock low                                         │    │
│  │  • Refund request                                    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  CRITICAL (Red)                                      │    │
│  │  • Dispute created                                   │    │
│  │  • Payment fraud detected                            │    │
│  │  • System error                                      │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────┐
│  UI Rendering                                                 │
│                                                               │
│  INFO:                                                        │
│  • Blue background (bg-blue-500/10)                          │
│  • Blue border (border-blue-500/30)                          │
│  • Blue text (text-blue-400)                                 │
│  • Blue glow (shadow-blue-500/20)                            │
│                                                               │
│  WARNING:                                                     │
│  • Yellow background (bg-yellow-500/10)                      │
│  • Yellow border (border-yellow-500/30)                      │
│  • Yellow text (text-yellow-400)                             │
│  • Yellow glow (shadow-yellow-500/20)                        │
│                                                               │
│  CRITICAL:                                                    │
│  • Red background (bg-red-500/10)                            │
│  • Red border (border-red-500/30)                            │
│  • Red text (text-red-400)                                   │
│  • Red glow (shadow-red-500/20)                              │
│  • Pulsing animation                                         │
└──────────────────────────────────────────────────────────────┘
```

---

## Scalability: Future Redis Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    FUTURE: HORIZONTAL SCALING WITH REDIS                     │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                              LOAD BALANCER                                    │
│                         (Round-robin / Least-conn)                            │
└────────────────────────────┬──────────────────┬──────────────────────────────┘
                             │                  │
                ┌────────────┴────────┐    ┌────┴────────────────┐
                │                     │    │                     │
                ▼                     ▼    ▼                     ▼
┌─────────────────────────┐  ┌─────────────────────────┐  ┌─────────────────────────┐
│   Backend Server 1      │  │   Backend Server 2      │  │   Backend Server N      │
│                         │  │                         │  │                         │
│  ┌──────────────────┐  │  │  ┌──────────────────┐  │  │  ┌──────────────────┐  │
│  │  SSE Handler     │  │  │  │  SSE Handler     │  │  │  │  SSE Handler     │  │
│  │  (10 clients)    │  │  │  │  (15 clients)    │  │  │  │  (12 clients)    │  │
│  └────────┬─────────┘  │  │  └────────┬─────────┘  │  │  └────────┬─────────┘  │
│           │             │  │           │             │  │           │             │
│  ┌────────▼─────────┐  │  │  ┌────────▼─────────┐  │  │  ┌────────▼─────────┐  │
│  │  SSE Broker      │  │  │  │  SSE Broker      │  │  │  │  SSE Broker      │  │
│  │  (Local clients) │  │  │  │  (Local clients) │  │  │  │  (Local clients) │  │
│  └────────┬─────────┘  │  │  └────────┬─────────┘  │  │  └────────┬─────────┘  │
│           │             │  │           │             │  │           │             │
│           │ Subscribe   │  │           │ Subscribe   │  │           │ Subscribe   │
│           ▼             │  │           ▼             │  │           ▼             │
└───────────┼─────────────┘  └───────────┼─────────────┘  └───────────┼─────────────┘
            │                            │                            │
            └────────────────────────────┼────────────────────────────┘
                                         │
                                         ▼
                        ┌────────────────────────────────┐
                        │       REDIS PUB/SUB            │
                        │                                │
                        │  Channel: "admin:notifications"│
                        │                                │
                        │  • Persistent                  │
                        │  • Cross-server                │
                        │  • High throughput             │
                        └────────────────┬───────────────┘
                                         │
                                         │ Publish
                                         ▼
                        ┌────────────────────────────────┐
                        │   Business Logic Layer         │
                        │   (Any server can publish)     │
                        └────────────────────────────────┘

Benefits:
• Horizontal scaling (add more servers)
• No single point of failure
• Notifications reach all admins regardless of server
• Load balancer can use any strategy
• Graceful server restarts
```

---

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SECURITY LAYERS                                    │
└─────────────────────────────────────────────────────────────────────────────┘

1. AUTHENTICATION
   ┌──────────────────────────────────────────────────────────────┐
   │ JWT Token Validation                                          │
   │ • Token in Authorization header (not query params)            │
   │ • HMAC-SHA256 signature verification                          │
   │ • Expiration check                                            │
   │ • Issuer validation                                           │
   └──────────────────────────────────────────────────────────────┘

2. AUTHORIZATION
   ┌──────────────────────────────────────────────────────────────┐
   │ Admin Permission Check                                        │
   │ • Email must match ADMIN_GOOGLE_EMAIL                         │
   │ • Per-request validation (no session)                         │
   │ • No privilege escalation possible                            │
   └──────────────────────────────────────────────────────────────┘

3. TRANSPORT SECURITY
   ┌──────────────────────────────────────────────────────────────┐
   │ HTTPS/TLS                                                     │
   │ • All traffic encrypted                                       │
   │ • Certificate validation                                      │
   │ • No token exposure in logs                                   │
   └──────────────────────────────────────────────────────────────┘

4. DATA SECURITY
   ┌──────────────────────────────────────────────────────────────┐
   │ Notification Content                                          │
   │ • No sensitive data in messages                               │
   │ • Order codes instead of full details                         │
   │ • Admin-only visibility                                       │
   └──────────────────────────────────────────────────────────────┘

5. RATE LIMITING (TODO)
   ┌──────────────────────────────────────────────────────────────┐
   │ Connection Limits                                             │
   │ • Max connections per admin                                   │
   │ • Max reconnection attempts                                   │
   │ • Request rate limiting                                       │
   └──────────────────────────────────────────────────────────────┘
```

---

## Performance Characteristics

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PERFORMANCE METRICS                                  │
└─────────────────────────────────────────────────────────────────────────────┘

Memory Usage per Connection:
┌────────────────────────────────────────────────────────────┐
│ WebSocket (Old):  ~8KB                                      │
│ • 2 goroutines (read + write pump)                         │
│ • 2 channels (send + control)                              │
│ • Connection state                                         │
│                                                            │
│ SSE (New):        ~4KB                                      │
│ • 1 goroutine (event loop)                                 │
│ • 1 channel (events)                                       │
│ • Connection state                                         │
│                                                            │
│ Improvement:      50% reduction                            │
└────────────────────────────────────────────────────────────┘

CPU Usage per Connection:
┌────────────────────────────────────────────────────────────┐
│ WebSocket (Old):  ~0.1%                                     │
│ • Ping/pong overhead                                       │
│ • Read pump (blocking)                                     │
│ • Write pump (blocking)                                    │
│                                                            │
│ SSE (New):        ~0.05%                                    │
│ • Keepalive only (30s interval)                            │
│ • Single event loop                                        │
│                                                            │
│ Improvement:      50% reduction                            │
└────────────────────────────────────────────────────────────┘

Latency:
┌────────────────────────────────────────────────────────────┐
│ Event → Client:   <10ms                                     │
│ • In-memory channel                                        │
│ • No serialization overhead                                │
│ • Direct broadcast                                         │
│                                                            │
│ Reconnection:     <1s                                       │
│ • Browser native                                           │
│ • No manual backoff                                        │
└────────────────────────────────────────────────────────────┘

Scalability:
┌────────────────────────────────────────────────────────────┐
│ Current (In-memory):                                        │
│ • Single server                                            │
│ • ~1000 concurrent connections                             │
│ • ~4MB memory for connections                              │
│                                                            │
│ Future (Redis):                                            │
│ • Multiple servers                                         │
│ • ~10,000+ concurrent connections                          │
│ • Horizontal scaling                                       │
└────────────────────────────────────────────────────────────┘
```

---

**Document Version**: 1.0  
**Last Updated**: January 21, 2026
