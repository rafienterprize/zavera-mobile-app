# SSE Migration Checklist

## 📋 Complete Migration Steps

### ✅ Phase 1: Backend Implementation (DONE)

- [x] Create `backend/service/sse_broker.go`
- [x] Create `backend/handler/sse_handler.go`
- [x] Update `backend/service/notification_service.go` (add severity)
- [x] Update `backend/routes/routes.go` (replace WebSocket route)
- [x] Update `backend/main.go` (initialize SSE broker)

### ✅ Phase 2: Frontend Implementation (DONE)

- [x] Create `frontend/src/hooks/useAdminSSE.ts`
- [x] Create `frontend/src/components/admin/NotificationBellSSE.tsx`

### ⏳ Phase 3: Integration (TODO - YOU NEED TO DO THIS)

- [ ] **Update admin layout to use new component**
  ```typescript
  // File: frontend/src/app/admin/layout.tsx
  
  // Replace this import:
  import { NotificationBell } from '@/components/admin/NotificationBell';
  
  // With this:
  import { NotificationBellSSE } from '@/components/admin/NotificationBellSSE';
  
  // Replace this in JSX:
  <NotificationBell />
  
  // With this:
  <NotificationBellSSE />
  ```

### ⏳ Phase 4: Testing (TODO - YOU NEED TO DO THIS)

- [ ] **Start backend**
  ```bash
  cd backend
  go run main.go
  ```
  
  Expected output:
  ```
  📡 SSE Broker initialized
  🚀 SSE Broker started
  🚀 Server starting on :8080...
  ```

- [ ] **Start frontend**
  ```bash
  cd frontend
  npm run dev
  ```

- [ ] **Open admin dashboard**
  - Navigate to: `http://localhost:3000/admin/dashboard`
  - Check: Connection indicator is green
  - Check: No console errors

- [ ] **Test notification flow**
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
  
  Expected:
  - [ ] Toast notification appears (top-right)
  - [ ] Notification in dropdown panel
  - [ ] Unread badge increments
  - [ ] Backend logs: "📢 Broadcast notification"

- [ ] **Test reconnection**
  - Stop backend (Ctrl+C)
  - Check: Connection indicator turns gray
  - Start backend again
  - Check: Connection indicator turns green (auto-reconnect)

- [ ] **Test multiple notification types**
  - [ ] Order created (checkout)
  - [ ] Payment received (payment webhook)
  - [ ] Shipment update (tracking webhook)
  - [ ] Low stock (manual trigger)

### ⏳ Phase 5: Cleanup (TODO - AFTER TESTING)

- [ ] **Delete old WebSocket files**
  ```bash
  # Backend
  rm backend/handler/websocket_handler.go
  
  # Frontend
  rm frontend/src/hooks/useAdminWebSocket.ts
  rm frontend/src/components/admin/NotificationBell.tsx
  ```

- [ ] **Remove WebSocket dependency (if not used elsewhere)**
  ```bash
  cd backend
  # Check if gorilla/websocket is used elsewhere
  grep -r "gorilla/websocket" .
  
  # If not used, remove from go.mod
  go mod tidy
  ```

### ⏳ Phase 6: Deployment (TODO - AFTER TESTING)

- [ ] **Build backend**
  ```bash
  cd backend
  go build -o zavera.exe
  ```

- [ ] **Build frontend**
  ```bash
  cd frontend
  npm run build
  ```

- [ ] **Deploy to production**
  - [ ] Deploy backend first
  - [ ] Verify SSE endpoint works
  - [ ] Deploy frontend
  - [ ] Monitor logs for 24 hours

### ⏳ Phase 7: Monitoring (TODO - AFTER DEPLOYMENT)

- [ ] **Check backend logs**
  - [ ] SSE broker started
  - [ ] Clients connecting
  - [ ] Notifications broadcasting
  - [ ] No errors

- [ ] **Check frontend**
  - [ ] Connection indicator green
  - [ ] Notifications appearing
  - [ ] No console errors
  - [ ] Auto-reconnect working

- [ ] **Monitor metrics**
  - [ ] Connection count
  - [ ] Notification delivery rate
  - [ ] Memory usage
  - [ ] CPU usage

---

## 🚨 Critical Steps (DO NOT SKIP)

### 1. Update Admin Layout (REQUIRED)

**File**: `frontend/src/app/admin/layout.tsx`

**Before**:
```typescript
import { NotificationBell } from '@/components/admin/NotificationBell';

// ...

<NotificationBell />
```

**After**:
```typescript
import { NotificationBellSSE } from '@/components/admin/NotificationBellSSE';

// ...

<NotificationBellSSE />
```

### 2. Verify Environment Variables (REQUIRED)

**Backend** (`.env`):
```env
ADMIN_GOOGLE_EMAIL=your-admin@gmail.com
JWT_SECRET=your-secret-key-change-this-in-production
```

**Frontend** (`.env.local`):
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 3. Test Before Deploying (REQUIRED)

- [ ] Test in development
- [ ] Test all notification types
- [ ] Test auto-reconnect
- [ ] Test with multiple admins
- [ ] Test under load (optional)

---

## 🐛 Troubleshooting

### Issue: "No connection" (gray dot)

**Checklist**:
- [ ] Backend is running
- [ ] SSE broker started (check logs)
- [ ] JWT token in localStorage
- [ ] User is admin (check ADMIN_GOOGLE_EMAIL)
- [ ] No CORS errors in console

**Fix**:
```bash
# Check backend logs
# Should see: "✅ SSE client connected"

# Check browser console
# Should see: "✅ SSE connected"

# Check token
localStorage.getItem('auth_token')
```

### Issue: "401 Unauthorized"

**Cause**: Invalid or missing token

**Fix**:
```typescript
// Check token
console.log(localStorage.getItem('auth_token'));

// Re-login if needed
```

### Issue: "403 Forbidden"

**Cause**: User is not admin

**Fix**:
```bash
# Update .env
ADMIN_GOOGLE_EMAIL=your-actual-email@gmail.com

# Restart backend
```

### Issue: No notifications received

**Cause**: SSE broker not started

**Fix**:
```go
// Check main.go has:
sseBroker := service.GetSSEBroker()
sseBroker.Start()

// Restart backend
```

---

## 📚 Documentation Reference

- **Quick Start**: `SSE_QUICK_START.md`
- **Full Migration Guide**: `SSE_MIGRATION_GUIDE.md`
- **Architecture Diagrams**: `SSE_ARCHITECTURE_DIAGRAM.md`
- **Implementation Summary**: `SSE_IMPLEMENTATION_SUMMARY.md`

---

## ✅ Success Criteria

### Functional
- [ ] All notification types work
- [ ] Multiple admins receive notifications
- [ ] Auto-reconnect works
- [ ] Toast notifications appear
- [ ] Dropdown panel works
- [ ] Severity-based styling applied

### Non-Functional
- [ ] <10ms latency for event delivery
- [ ] <1s reconnection time
- [ ] No memory leaks
- [ ] No CPU spikes
- [ ] 1000+ concurrent connections supported

### Quality
- [ ] No console errors
- [ ] No backend errors
- [ ] Clean code
- [ ] Proper documentation
- [ ] Security best practices

---

## 🎯 Next Steps

1. **Complete Phase 3**: Update admin layout
2. **Complete Phase 4**: Test thoroughly
3. **Complete Phase 5**: Cleanup old files
4. **Complete Phase 6**: Deploy to production
5. **Complete Phase 7**: Monitor for 24 hours

---

## 📞 Need Help?

- **Backend Issues**: Check `backend/handler/sse_handler.go` logs
- **Frontend Issues**: Check browser console
- **Architecture Questions**: Read `SSE_ARCHITECTURE_DIAGRAM.md`
- **Migration Questions**: Read `SSE_MIGRATION_GUIDE.md`

---

**Current Status**: ✅ Implementation Complete, ⏳ Integration Pending

**Last Updated**: January 21, 2026
