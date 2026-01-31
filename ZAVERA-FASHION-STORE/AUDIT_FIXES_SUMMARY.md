# Admin Audit Log Foreign Key Constraint Fix

## Problem
When attempting to cancel EXPIRED orders, the system was throwing a foreign key constraint error:
```
pq: insert or update on table "admin_audit_log" violates foreign key constraint "admin_audit_log_admin_user_id_fkey"
```

## Root Cause
The `admin_audit_log` table had `admin_user_id` defined as `NOT NULL` with a foreign key constraint to the `users` table. However, in some cases (like when admin context is not fully available), the `admin.UserID` was `0`, which doesn't exist in the `users` table, causing the foreign key violation.

## Solution

### 1. Database Schema Fix
Created migration `database/migrate_audit_fixes.sql` to make `admin_user_id` nullable:
```sql
ALTER TABLE admin_audit_log 
ALTER COLUMN admin_user_id DROP NOT NULL;
```

This allows the column to accept `NULL` values when admin user context is not available.

### 2. Backend Code Fix
Updated `backend/service/admin_service.go` and `backend/models/admin_audit.go`:

**Model Change:**
```go
// Changed from:
AdminUserID int `json:"admin_user_id" db:"admin_user_id"`

// To:
AdminUserID *int `json:"admin_user_id,omitempty" db:"admin_user_id"`
```

**Service Logic:**
Added null-safe handling in all admin action functions:
```go
var adminUserID *int
if admin.UserID > 0 {
    adminUserID = &admin.UserID
}

auditLog := &models.AdminAuditLog{
    AdminUserID: adminUserID,  // Will be nil if UserID is 0
    // ... rest of fields
}
```

This was applied to:
- `ForceCancel()` - line ~100
- `ForceRefund()` - line ~280
- `ForceReship()` - line ~410
- `ReconcilePayment()` - line ~550
- `logFailedAction()` - line ~650

## Files Modified

### Database
- `database/migrate_audit_fixes.sql` (NEW) - Migration to allow NULL admin_user_id

### Backend
- `backend/models/admin_audit.go` - Changed AdminUserID to nullable pointer
- `backend/service/admin_service.go` - Added null-safe handling for all admin actions

## Testing Steps

1. ✅ Run migration: `psql -U postgres -d zavera_db -f database/migrate_audit_fixes.sql`
2. ✅ Compile backend: `go build -o zavera.exe` in backend folder
3. ⏳ Stop old backend process
4. ⏳ Start new backend: `.\zavera.exe`
5. ⏳ Test cancel EXPIRED order from admin panel

## Expected Behavior
- Admin should be able to cancel EXPIRED orders without foreign key errors
- Audit log will be created with `admin_user_id = NULL` if user context is not available
- All other admin actions should continue to work normally

## Notes
- The `admin_email` field is still required and will always be populated
- This fix maintains audit trail integrity while allowing flexibility for system actions
- Foreign key constraint still exists, but now allows NULL values
