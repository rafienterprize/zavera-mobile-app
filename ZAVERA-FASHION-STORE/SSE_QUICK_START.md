# SSE Quick Start Guide

## 🚀 5-Minute Setup

### 1. Update Admin Layout (Frontend)

**File**: `frontend/src/app/admin/layout.tsx`

```typescript
// Replace this:
import { NotificationBell } from '@/components/admin/NotificationBell';

// With this:
import { NotificationBellSSE } from '@/components/admin/NotificationBellSSE';

// In the component:
<NotificationBellSSE />
```

### 2. Start Backend

```bash
cd backend
go run main.go
```

You should see:
```
📡 SSE Broker initialized
🚀 SSE Broker started
🚀 Server starting on :8080...
```

### 3. Start Frontend

```bash
cd frontend
npm run dev
```

### 4. Test

1. Open admin dashboard: `http://localhost:3000/admin/dashboard`
2. Check connection indicator (green dot = connected)
3. Trigger a test notification:

```bash
# Create a test order
curl -X POST http://localhost:8080/api/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Test User",
    "email": "test@example.com",
    "phone": "08123456789",
    "items": [{"product_id": 1, "quantity": 1}]
  }'
```

4. You should see a toast notification appear!

---

## 📋 Checklist

- [ ] Backend SSE broker initialized
- [ ] Frontend using `NotificationBellSSE` component
- [ ] Connection indicator shows green
- [ ] Test notification received
- [ ] Toast appears in top-right corner
- [ ] Notification dropdown works
- [ ] Unread badge updates

---

## 🎨 UI Features

### Toast Notifications
- Appear in top-right corner
- Auto-dismiss after 5 seconds
- Severity-based colors:
  - 🔵 Blue = Info
  - 🟡 Yellow = Warning
  - 🔴 Red = Critical

### Notification Dropdown
- Click bell icon to open
- Shows all notifications
- Unread notifications highlighted
- Click notification to navigate
- "Mark all as read" button

### Connection Indicator
- 🟢 Green = Connected
- ⚪ Gray = Disconnected
- Auto-reconnects on disconnect

---

## 🔧 Configuration

### Backend (.env)

```env
# Admin email (required for SSE access)
ADMIN_GOOGLE_EMAIL=your-admin@gmail.com

# JWT secret (must match frontend)
JWT_SECRET=your-secret-key-change-this-in-production
```

### Frontend (.env.local)

```env
# Backend URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 🐛 Troubleshooting

### "No connection" (gray dot)

**Check**:
1. Backend is running
2. JWT token in localStorage
3. User is admin (check ADMIN_GOOGLE_EMAIL)
4. No CORS errors in console

**Fix**:
```bash
# Check backend logs
# Should see: "✅ SSE client connected"

# Check browser console
# Should see: "✅ SSE connected"
```

### "401 Unauthorized"

**Cause**: Invalid or missing token

**Fix**:
```typescript
// Check token
console.log(localStorage.getItem('auth_token'));

// Re-login if needed
```

### "403 Forbidden"

**Cause**: User is not admin

**Fix**:
```bash
# Update .env
ADMIN_GOOGLE_EMAIL=your-actual-email@gmail.com

# Restart backend
```

### No notifications received

**Cause**: SSE broker not started

**Fix**:
```go
// Check main.go has:
sseBroker := service.GetSSEBroker()
sseBroker.Start()
```

---

## 📊 Testing Notifications

### Trigger Order Created
```bash
curl -X POST http://localhost:8080/api/checkout \
  -H "Content-Type: application/json" \
  -d '{"customer_name":"Test","email":"test@test.com","phone":"08123456789","items":[{"product_id":1,"quantity":1}]}'
```

### Trigger Payment Received
```bash
# Complete payment via Midtrans
# Or manually call NotifyPaymentReceived in code
```

### Trigger Low Stock
```go
// In product service
service.NotifyStockLow("Product Name", 3)
```

---

## 🎯 Next Steps

1. ✅ Basic setup complete
2. ⏭️ Test all notification types
3. ⏭️ Customize notification styling
4. ⏭️ Add notification persistence (database)
5. ⏭️ Implement Redis Pub/Sub for scaling
6. ⏭️ Add notification preferences
7. ⏭️ Add notification history page

---

## 📚 Full Documentation

See `SSE_MIGRATION_GUIDE.md` for:
- Complete architecture details
- Scalability considerations
- Security best practices
- Performance metrics
- Rollback procedures

---

**Ready to go!** 🎉
