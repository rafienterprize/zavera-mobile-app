# ZAVERA Commercial Hardening - Phase 1 Implementation

## Overview

Phase 1 implements the **Critical Safety Layer** for ZAVERA e-commerce platform, focusing on:
- Refund system with Midtrans integration
- Payment reconciliation
- Stuck payment recovery
- Admin force-actions with full audit trail

## 1️⃣ Database Schema

### New Tables

| Table | Purpose |
|-------|---------|
| `refunds` | Master refund records with gateway integration |
| `refund_items` | Individual items in partial refunds |
| `admin_audit_log` | Immutable audit trail for all admin actions |
| `payment_sync_log` | Payment status sync tracking |
| `reconciliation_log` | Daily reconciliation records |
| `refund_status_history` | Refund status change history |

### New Columns

**orders:**
- `refund_status` - Current refund status
- `refund_amount` - Total refunded amount
- `refunded_at` - Refund timestamp
- `is_refundable` - Refund eligibility flag
- `last_synced_at` - Last gateway sync time

**payments:**
- `refund_status` - Refund status
- `refunded_amount` - Amount refunded
- `refundable_amount` - Remaining refundable
- `gateway_status` - Gateway status
- `is_reconciled` - Reconciliation flag

**shipments:**
- `reship_count` - Number of reships
- `original_shipment_id` - Original shipment reference
- `is_replacement` - Replacement flag

### Migration

```bash
# Run migration
migrate_hardening.bat
```

## 2️⃣ Refund Engine

### Services

**RefundService** (`backend/service/refund_service.go`)
- Full refund
- Partial refund
- Shipping-only refund
- Item-specific refund
- Midtrans refund API integration
- Idempotency protection
- Stock restoration

### Refund Types

| Type | Description |
|------|-------------|
| `FULL` | Complete order refund |
| `PARTIAL` | Specific amount refund |
| `SHIPPING_ONLY` | Shipping cost only |
| `ITEM_ONLY` | Specific items refund |

### Refund Reasons

- `CUSTOMER_REQUEST`
- `OUT_OF_STOCK`
- `DAMAGED_ITEM`
- `WRONG_ITEM`
- `LATE_DELIVERY`
- `DUPLICATE_ORDER`
- `FRAUD_SUSPECTED`
- `ADMIN_DECISION`
- `SHIPPING_FAILED`
- `OTHER`

## 3️⃣ Admin Force Actions

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/admin/orders/:code/force-cancel` | POST | Force cancel order |
| `/api/admin/orders/:code/refund` | POST | Force refund order |
| `/api/admin/orders/:code/reship` | POST | Create replacement shipment |
| `/api/admin/payments/:id/reconcile` | POST | Manual payment reconciliation |

### Safety Features

- **Transaction Safety**: All actions run inside DB transactions
- **Audit Logging**: Every action logged to `admin_audit_log`
- **Idempotency**: Duplicate requests return cached result
- **State Capture**: Before/after state recorded
- **No Deletions**: Records are never deleted, only status changed

### Request Examples

**Force Cancel:**
```json
POST /api/admin/orders/ZVR-20260110-ABC123/force-cancel
{
  "reason": "Customer requested cancellation",
  "restore_stock": true,
  "idempotency_key": "cancel-abc123-001"
}
```

**Force Refund:**
```json
POST /api/admin/orders/ZVR-20260110-ABC123/refund
{
  "refund_type": "FULL",
  "reason": "Customer complaint",
  "skip_gateway": false,
  "idempotency_key": "refund-abc123-001"
}
```

**Reconcile Payment:**
```json
POST /api/admin/payments/123/reconcile
{
  "action": "MARK_PAID",
  "reason": "Manual verification - payment confirmed",
  "transaction_id": "TXN123456",
  "idempotency_key": "reconcile-123-001"
}
```

## 4️⃣ Payment Recovery Engine

### Services

**PaymentRecoveryService** (`backend/service/payment_recovery_service.go`)
- Midtrans status polling
- Stuck payment detection
- Orphan order/payment detection
- Auto-resolution of mismatches
- Background recovery job

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/admin/payments/:id/sync` | POST | Sync single payment |
| `/api/admin/payments/stuck` | GET | List stuck payments |
| `/api/admin/payments/sync-all` | POST | Sync all pending payments |

### Background Job

Enable in `.env`:
```
ENABLE_RECOVERY_JOB=true
```

Runs every 15 minutes:
1. Syncs pending payments with Midtrans
2. Recovers stuck payments
3. Resolves orphan orders
4. Sends alerts for critical issues

## 5️⃣ Daily Reconciliation

### Services

**ReconciliationService** (`backend/service/reconciliation_service.go`)
- Order/payment statistics
- Mismatch detection
- Orphan detection
- Revenue calculation
- Stuck payment tracking

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/admin/reconciliation/run` | POST | Run reconciliation |
| `/api/admin/reconciliation` | GET | Get summary |
| `/api/admin/reconciliation/mismatches` | GET | List mismatches |

### Background Job

Enable in `.env`:
```
ENABLE_RECONCILIATION_JOB=true
```

Runs daily at 2 AM for previous day.

### Reconciliation Summary

```json
{
  "date": "2026-01-09",
  "total_orders": 150,
  "total_payments": 148,
  "total_amount": 45000000,
  "orders_by_status": {
    "pending": 10,
    "paid": 120,
    "cancelled": 15,
    "refunded": 5
  },
  "mismatches_found": 2,
  "orphan_orders": 1,
  "stuck_payments": 3,
  "expected_revenue": 44500000,
  "actual_revenue": 44500000,
  "revenue_variance": 0
}
```

## 6️⃣ Audit Trail

### Admin Audit Log

Every admin action creates an immutable record:

```json
{
  "id": 1,
  "admin_user_id": 5,
  "admin_email": "admin@zavera.com",
  "admin_ip": "192.168.1.100",
  "action_type": "FORCE_CANCEL",
  "action_detail": "Force cancelled order: ZVR-20260110-ABC123",
  "target_type": "order",
  "target_id": 42,
  "target_code": "ZVR-20260110-ABC123",
  "state_before": {"status": "PENDING", "stock_reserved": true},
  "state_after": {"status": "CANCELLED", "stock_restored": true},
  "success": true,
  "idempotency_key": "cancel-abc123-001",
  "created_at": "2026-01-10T10:30:00Z"
}
```

### Endpoint

```
GET /api/admin/audit-logs?limit=50
```

## Quality Guarantees

This system survives:

| Scenario | Protection |
|----------|------------|
| Double webhooks | Idempotency keys |
| Duplicate refunds | Idempotency + status checks |
| Admin clicking twice | Idempotency keys |
| Race conditions | DB transactions + row locks |
| Network failures | Retry logic + sync jobs |

## File Structure

```
backend/
├── models/
│   ├── refund.go           # Refund models
│   └── admin_audit.go      # Audit & sync models
├── repository/
│   ├── refund_repository.go
│   └── admin_audit_repository.go
├── service/
│   ├── refund_service.go
│   ├── admin_service.go
│   ├── payment_recovery_service.go
│   └── reconciliation_service.go
├── handler/
│   └── admin_hardening_handler.go
├── dto/
│   └── hardening_dto.go
└── routes/
    └── routes.go           # Updated with new endpoints

database/
└── migrate_hardening.sql   # Migration script
```

## Environment Variables

```env
# Midtrans
MIDTRANS_SERVER_KEY=your_server_key
MIDTRANS_ENVIRONMENT=sandbox  # or production

# Background Jobs
ENABLE_RECOVERY_JOB=true      # Payment recovery job
ENABLE_RECONCILIATION_JOB=true # Daily reconciliation
```

## Data Flows

### Refund Flow
```
Admin Request → Idempotency Check → Create Refund Record
    → Process with Midtrans → Update Status → Restore Stock
    → Update Order → Audit Log
```

### Payment Recovery Flow
```
Cron Job → Find Pending Payments → Query Midtrans Status
    → Compare Local vs Gateway → Auto-resolve if possible
    → Log Sync Result → Alert if critical
```

### Reconciliation Flow
```
Daily Job → Query Orders/Payments → Calculate Stats
    → Find Mismatches → Find Orphans → Find Stuck
    → Calculate Revenue → Save Report → Alert if issues
```
