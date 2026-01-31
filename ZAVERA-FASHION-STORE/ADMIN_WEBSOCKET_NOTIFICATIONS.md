# Admin WebSocket Real-Time Notifications

## Overview
Implementasi WebSocket untuk admin dashboard agar admin mendapat notifikasi real-time saat ada:
- Order baru
- Payment berhasil
- Payment expired
- Shipment updates
- Stock low alerts
- Refund requests
- Disputes

## Backend Implementation ✅ DONE

### 1. WebSocket Handler (`backend/handler/websocket_handler.go`)
- WebSocket server dengan gorilla/websocket
- Hub pattern untuk manage multiple admin connections
- Auto-reconnect support
- Ping/pong untuk keep-alive

### 2. Notification Service (`backend/service/notification_service.go`)
- `NotifyOrderCreated()` - Saat order baru dibuat
- `NotifyPaymentReceived()` - Saat payment berhasil
- `NotifyPaymentExpired()` - Saat payment expired
- `NotifyShipmentUpdate()` - Saat shipment status berubah
- `NotifyStockLow()` - Saat stock rendah
- `NotifyRefundRequest()` - Saat ada refund request
- `NotifyDisputeCreated()` - Saat ada dispute baru

### 3. Integration Points
- `checkout_service.go` - Trigger `NotifyOrderCreated()` setelah order dibuat
- `core_payment_service.go` - Trigger `NotifyPaymentReceived()` setelah payment settlement

### 4. WebSocket Endpoint
```
WS /api/admin/ws
Authorization: Bearer <admin_token>
```

## Frontend Implementation (TODO)

### 1. WebSocket Hook (`frontend/src/hooks/useAdminWebSocket.ts`)

```typescript
import { useEffect, useState, useCallback, useRef } from 'react';

interface AdminNotification {
  id: string;
  type: string;
  title: string;
  message: string;
  data?: any;
  timestamp: string;
  read: boolean;
}

export function useAdminWebSocket() {
  const [notifications, setNotifications] = useState<AdminNotification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) return;

    // WebSocket URL
    const wsUrl = `ws://localhost:8080/api/admin/ws?token=${token}`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('🔌 WebSocket connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      const notification: AdminNotification = JSON.parse(event.data);
      console.log('📢 Notification received:', notification);
      
      // Add to notifications list
      setNotifications(prev => [notification, ...prev]);
      setUnreadCount(prev => prev + 1);
      
      // Play sound
      playNotificationSound();
      
      // Show browser notification
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(notification.title, {
          body: notification.message,
          icon: '/favicon.ico',
        });
      }
    };

    ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };

    ws.onclose = () => {
      console.log('🔌 WebSocket disconnected');
      setIsConnected(false);
      
      // Auto-reconnect after 5 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        console.log('🔄 Reconnecting WebSocket...');
        connect();
      }, 5000);
    };
  }, []);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
  }, []);

  const markAsRead = useCallback((notificationId: string) => {
    setNotifications(prev =>
      prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
    
    // Send to server
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        action: 'mark_read',
        notification_id: notificationId,
      }));
    }
  }, []);

  const markAllAsRead = useCallback(() => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    setUnreadCount(0);
  }, []);

  const clearNotifications = useCallback(() => {
    setNotifications([]);
    setUnreadCount(0);
  }, []);

  useEffect(() => {
    connect();
    
    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission();
    }
    
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
    clearNotifications,
  };
}

function playNotificationSound() {
  const audio = new Audio('/sounds/notification.mp3');
  audio.volume = 0.5;
  audio.play().catch(err => console.log('Sound play failed:', err));
}
```

### 2. Notification Bell Component (`frontend/src/components/admin/NotificationBell.tsx`)

```typescript
'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useAdminWebSocket } from '@/hooks/useAdminWebSocket';

export function NotificationBell() {
  const { notifications, unreadCount, markAsRead, markAllAsRead } = useAdminWebSocket();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="relative">
      {/* Bell Icon with Badge */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-gray-400 hover:text-white transition-colors"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        
        {/* Badge */}
        {unreadCount > 0 && (
          <motion.span
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center"
          >
            {unreadCount > 9 ? '9+' : unreadCount}
          </motion.span>
        )}
      </button>

      {/* Dropdown Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-xl border border-gray-200 z-50"
          >
            {/* Header */}
            <div className="p-4 border-b border-gray-200 flex items-center justify-between">
              <h3 className="font-semibold text-gray-900">Notifications</h3>
              {unreadCount > 0 && (
                <button
                  onClick={markAllAsRead}
                  className="text-sm text-blue-600 hover:text-blue-700"
                >
                  Mark all as read
                </button>
              )}
            </div>

            {/* Notifications List */}
            <div className="max-h-96 overflow-y-auto">
              {notifications.length === 0 ? (
                <div className="p-8 text-center text-gray-500">
                  <svg className="w-12 h-12 mx-auto mb-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                  </svg>
                  <p>No notifications yet</p>
                </div>
              ) : (
                notifications.map((notif) => (
                  <div
                    key={notif.id}
                    onClick={() => markAsRead(notif.id)}
                    className={`p-4 border-b border-gray-100 hover:bg-gray-50 cursor-pointer transition-colors ${
                      !notif.read ? 'bg-blue-50' : ''
                    }`}
                  >
                    <div className="flex items-start gap-3">
                      <div className="flex-1">
                        <p className="font-medium text-gray-900">{notif.title}</p>
                        <p className="text-sm text-gray-600 mt-1">{notif.message}</p>
                        <p className="text-xs text-gray-400 mt-2">
                          {new Date(notif.timestamp).toLocaleString()}
                        </p>
                      </div>
                      {!notif.read && (
                        <div className="w-2 h-2 bg-blue-500 rounded-full mt-2"></div>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* Footer */}
            {notifications.length > 0 && (
              <div className="p-3 border-t border-gray-200 text-center">
                <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
                  View all notifications
                </button>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
```

### 3. Integration in Admin Layout

Update `frontend/src/app/admin/layout.tsx`:

```typescript
import { NotificationBell } from '@/components/admin/NotificationBell';

// In the header section:
<div className="flex items-center gap-4">
  <NotificationBell />
  {/* Other header items */}
</div>
```

### 4. Auto-Refresh Dashboard on Notification

In `frontend/src/app/admin/dashboard/page.tsx`:

```typescript
import { useAdminWebSocket } from '@/hooks/useAdminWebSocket';

export default function DashboardPage() {
  const { notifications } = useAdminWebSocket();
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // Auto-refresh when new order or payment notification
  useEffect(() => {
    const lastNotif = notifications[0];
    if (lastNotif && (
      lastNotif.type === 'order_created' ||
      lastNotif.type === 'payment_received'
    )) {
      setRefreshTrigger(prev => prev + 1);
    }
  }, [notifications]);

  // Use refreshTrigger in data fetching
  useEffect(() => {
    fetchDashboardData();
  }, [refreshTrigger]);

  // ...
}
```

## Notification Types & Icons

| Type | Icon | Color | Sound |
|------|------|-------|-------|
| order_created | 🛍️ | Blue | Ding |
| payment_received | 💰 | Green | Cash |
| payment_expired | ⏰ | Orange | Alert |
| shipment_update | 📦 | Purple | Notification |
| stock_low | ⚠️ | Yellow | Warning |
| refund_request | 💸 | Red | Alert |
| dispute_created | ⚠️ | Red | Urgent |

## Testing

### Backend Test:
```bash
# Start backend
cd backend
./zavera.exe

# Check WebSocket endpoint
wscat -c ws://localhost:8080/api/admin/ws -H "Authorization: Bearer <token>"
```

### Frontend Test:
1. Login as admin
2. Open admin dashboard
3. In another browser, create an order
4. Admin should see notification bell badge increase
5. Click bell to see notification
6. Click notification to mark as read

## Files Created/Modified

### Backend:
- ✅ `backend/handler/websocket_handler.go` (NEW)
- ✅ `backend/service/notification_service.go` (NEW)
- ✅ `backend/service/checkout_service.go` (MODIFIED - added NotifyOrderCreated)
- ✅ `backend/service/core_payment_service.go` (MODIFIED - added NotifyPaymentReceived)
- ✅ `backend/routes/routes.go` (MODIFIED - added /admin/ws route)
- ✅ `backend/go.mod` (MODIFIED - added gorilla/websocket)

### Frontend (TODO):
- `frontend/src/hooks/useAdminWebSocket.ts` (NEW)
- `frontend/src/components/admin/NotificationBell.tsx` (NEW)
- `frontend/src/app/admin/layout.tsx` (MODIFY)
- `frontend/src/app/admin/dashboard/page.tsx` (MODIFY)
- `frontend/public/sounds/notification.mp3` (ADD)

## Next Steps

1. ✅ Backend WebSocket server - DONE
2. ✅ Notification triggers - DONE
3. ⏳ Frontend WebSocket hook - TODO
4. ⏳ Notification bell component - TODO
5. ⏳ Integration in admin layout - TODO
6. ⏳ Auto-refresh dashboard - TODO
7. ⏳ Sound alerts - TODO
8. ⏳ Browser notifications - TODO

Mau saya lanjutkan implementasi frontend nya sekarang?
