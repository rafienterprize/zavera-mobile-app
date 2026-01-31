# ZAVERA Commercial Hardening Audit - Phase 1 Complete

## Project Context

ZAVERA adalah e-commerce fashion store dengan stack:
- **Backend**: Go (Gin framework)
- **Frontend**: Next.js + TypeScript
- **Database**: PostgreSQL
- **Payment Gateway**: Midtrans (Snap)
- **Shipping**: RajaOngkir API integration

---

## PHASE 1 COMPLETED ✅

### What Was Implemented

#### 1. DATABASE SCHEMA (migrate_hardening.sql)

**New Tables:**
```sql
-- Refund system
CREATE TABLE refunds (
    id SERIAL PRIMARY KEY,
    refund_code VARCHAR(50) UNIQUE NOT NULL,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    payment_id INTEGER REFERENCES payments(id),
    refund_type refund_type NOT NULL,  -- FULL, PARTIAL, SHIPPING_ONLY, ITEM_ONLY
    reason refund_reason NOT NULL,
    original_amount DECIMAL(12, 2) NOT NULL,
    refund_amount DECIMAL(12, 2) NOT NULL,
    shipping_refund DECIMAL(12, 2) DEFAULT 0,
    items_refund DECIMAL(12, 2) DEFAULT 0,
    status refund_status DEFAULT 'PENDING',
    gateway_refund_id VARCHAR(255),
    gateway_response JSONB,
    idempotency_key VARCHAR(100) UNIQUE,
    processed_by INTEGER REFERENCES users(id),
    ...
);

CREATE TABLE refund_items (
    id SERIAL PRIMARY KEY,
    refund_id INTEGER NOT NULL REFERENCES refunds(id),
    order_item_id INTEGER NOT NULL REFERENCES order_items(id),
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    refund_amount DECIMAL(12, 2) NOT NULL,
    stock_restored BOOLEAN DEFAULT false,
    ...
);

-- Admin audit (IMMUTABLE - has trigger to prevent updates)
CREATE TABLE admin_audit_log (
    id SERIAL PRIMARY KEY,
    admin_user_id INTEGER NOT NULL REFERENCES users(id),
    admin_email VARCHAR(255) NOT NULL,
    admin_ip VARCHAR(50),
    action_type admin_action_type NOT NULL,
    action_detail TEXT NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id INTEGER NOT NULL,
    state_before JSONB,
    state_after JSONB,
    success BOOLEAN NOT NULL,
    idempotency_key VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payment sync tracking
CREATE TABLE payment_sync_log (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER NOT NULL REFERENCES payments(id),
    order_id INTEGER NOT NULL REFERENCES orders(id),
    sync_type VARCHAR(50) NOT NULL,
    sync_status payment_sync_status DEFAULT 'PENDING',
    local_payment_status VARCHAR(50),
    gateway_status VARCHAR(50),
    has_mismatch BOOLEAN DEFAULT false,
    mismatch_type VARCHAR(100),
    resolved BOOLEAN DEFAULT false,
    retry_count INTEGER DEFAULT 0,
    ...
);

-- Daily reconciliation
CREATE TABLE reconciliation_log (
    id SERIAL PRIMARY KEY,
    reconciliation_date DATE NOT NULL UNIQUE,
    total_orders INTEGER,
    total_payments INTEGER,
    total_amount DECIMAL(14, 2),
    mismatches_found INTEGER,
    orphan_orders INTEGER,
    stuck_payments INTEGER,
    expected_revenue DECIMAL(14, 2),
    actual_revenue DECIMAL(14, 2),
    revenue_variance DECIMAL(14, 2),
    status VARCHAR(50),
    ...
);
```

**New Columns Added:**
```sql
-- orders table
ALTER TABLE orders ADD COLUMN refund_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN refund_amount DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN refunded_at TIMESTAMP;
ALTER TABLE orders ADD COLUMN is_refundable BOOLEAN DEFAULT true;
ALTER TABLE orders ADD COLUMN last_synced_at TIMESTAMP;

-- payments table
ALTER TABLE payments ADD COLUMN refund_status VARCHAR(50);
ALTER TABLE payments ADD COLUMN refunded_amount DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE payments ADD COLUMN refundable_amount DECIMAL(12, 2);
ALTER TABLE payments ADD COLUMN gateway_status VARCHAR(100);
ALTER TABLE payments ADD COLUMN is_reconciled BOOLEAN DEFAULT false;

-- shipments table
ALTER TABLE shipments ADD COLUMN reship_count INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN original_shipment_id INTEGER REFERENCES shipments(id);
ALTER TABLE shipments ADD COLUMN is_replacement BOOLEAN DEFAULT false;
```

---

#### 2. BACKEND SERVICES

**Files Created:**

| File | Purpose |
|------|---------|
| `backend/models/refund.go` | Refund, RefundItem models + status enums |
| `backend/models/admin_audit.go` | AdminAuditLog, PaymentSyncLog, ReconciliationLog models |
| `backend/dto/hardening_dto.go` | All DTOs for hardening APIs |
| `backend/repository/refund_repository.go` | Refund CRUD operations |
| `backend/repository/admin_audit_repository.go` | Audit, PaymentSync, Reconciliation repos |
| `backend/service/refund_service.go` | Refund engine + Midtrans refund API |
| `backend/service/admin_service.go` | Force cancel/refund/reship/reconcile |
| `backend/service/payment_recovery_service.go` | Stuck payment detection & recovery |
| `backend/service/reconciliation_service.go` | Daily reconciliation engine |
| `backend/handler/admin_hardening_handler.go` | All admin API handlers |

---

#### 3. API ENDPOINTS

**Admin Force Actions:**
```
POST /api/admin/orders/:code/force-cancel
POST /api/admin/orders/:code/refund
POST /api/admin/orders/:code/reship
POST /api/admin/payments/:id/reconcile
```

**Refund Management:**
```
POST /api/admin/refunds
GET  /api/admin/refunds/:code
POST /api/admin/refunds/:code/process
```

**Payment Recovery:**
```
POST /api/admin/payments/:id/sync
GET  /api/admin/payments/stuck
POST /api/admin/payments/sync-all
```

**Reconciliation:**
```
POST /api/admin/reconciliation/run
GET  /api/admin/reconciliation
GET  /api/admin/reconciliation/mismatches
```

**Audit:**
```
GET /api/admin/audit-logs
```

---

#### 4. BACKGROUND JOBS

**Payment Recovery Job** (every 15 minutes):
- Syncs pending payments with Midtrans
- Detects stuck payments (>2 hours pending)
- Resolves orphan orders
- Auto-resolves status mismatches
- Sends alerts for critical issues

**Reconciliation Job** (daily at 2 AM):
- Calculates order/payment statistics
- Detects mismatches
- Tracks revenue variance
- Logs daily report

Enable via `.env`:
```
ENABLE_RECOVERY_JOB=true
ENABLE_RECONCILIATION_JOB=true
```

---

#### 5. SAFETY FEATURES IMPLEMENTED

| Feature | Implementation |
|---------|----------------|
| Idempotency | All admin actions support `idempotency_key` |
| Transaction Safety | All mutations wrapped in DB transactions |
| Audit Trail | Immutable `admin_audit_log` with state snapshots |
| No Deletions | Records never deleted, only status changed |
| Race Protection | `FOR UPDATE` row locks on critical operations |
| Double Webhook | Payment webhook has idempotency check |
| Stock Safety | Automatic stock restoration on cancel/refund |

---

## WHAT'S NOT IMPLEMENTED (For Phase 2+)

### Phase 2 - Operational Layer
- [ ] Inventory management system
- [ ] Low stock alerts
- [ ] Supplier management
- [ ] Purchase orders
- [ ] Warehouse management

### Phase 3 - Customer Experience
- [ ] Customer refund request portal
- [ ] Order tracking improvements
- [ ] Email notifications for refunds
- [ ] SMS notifications

### Phase 4 - Analytics & Reporting
- [ ] Sales dashboard
- [ ] Refund analytics
- [ ] Customer behavior tracking
- [ ] Financial reports

### Phase 5 - Security Hardening
- [ ] Admin role-based access control (RBAC)
- [ ] Rate limiting on admin APIs
- [ ] IP whitelisting for admin
- [ ] Two-factor authentication
- [ ] API key management

### Phase 6 - Infrastructure
- [ ] Redis caching
- [ ] Message queue (for async operations)
- [ ] Logging aggregation
- [ ] Monitoring & alerting (Prometheus/Grafana)
- [ ] Backup automation

---

## EXISTING CODEBASE STRUCTURE

```
backend/
├── main.go                 # Entry point
├── config/                 # Database config
├── models/
│   ├── models.go          # Order, Payment, User, Product, Cart models
│   ├── shipping.go        # Shipment, TrackingEvent models
│   ├── refund.go          # [NEW] Refund models
│   └── admin_audit.go     # [NEW] Audit models
├── repository/
│   ├── order_repository.go
│   ├── payment_repository.go
│   ├── product_repository.go
│   ├── cart_repository.go
│   ├── user_repository.go
│   ├── shipping_repository.go
│   ├── refund_repository.go      # [NEW]
│   └── admin_audit_repository.go # [NEW]
├── service/
│   ├── order_service.go
│   ├── payment_service.go        # Midtrans Snap integration
│   ├── product_service.go
│   ├── cart_service.go
│   ├── auth_service.go           # JWT + Google OAuth
│   ├── shipping_service.go       # RajaOngkir integration
│   ├── checkout_service.go
│   ├── tracking_job.go           # Shipment tracking cron
│   ├── refund_service.go         # [NEW]
│   ├── admin_service.go          # [NEW]
│   ├── payment_recovery_service.go # [NEW]
│   └── reconciliation_service.go   # [NEW]
├── handler/
│   ├── order_handler.go
│   ├── payment_handler.go
│   ├── product_handler.go
│   ├── cart_handler.go
│   ├── auth_handler.go
│   ├── shipping_handler.go
│   ├── checkout_handler.go
│   └── admin_hardening_handler.go # [NEW]
├── dto/
│   ├── dto.go
│   ├── shipping_dto.go
│   └── hardening_dto.go          # [NEW]
└── routes/
    └── routes.go                 # All route definitions

database/
├── schema.sql              # Base schema
├── migrate_shipping.sql    # Shipping tables
├── migrate_auth.sql        # Auth tables
└── migrate_hardening.sql   # [NEW] Hardening tables

frontend/
└── src/
    ├── app/                # Next.js app router
    ├── components/         # React components
    ├── lib/               # API client, utils
    └── types/             # TypeScript types
```

---

## ENVIRONMENT VARIABLES

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=zavera

# Midtrans
MIDTRANS_SERVER_KEY=SB-Mid-server-xxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxx
MIDTRANS_ENVIRONMENT=sandbox

# JWT
JWT_SECRET=your-secret-key

# RajaOngkir
RAJAONGKIR_API_KEY=your-api-key

# Background Jobs
ENABLE_TRACKING_JOB=true
ENABLE_RECOVERY_JOB=true
ENABLE_RECONCILIATION_JOB=true
```

---

## NOTES FOR NEXT PHASE

1. **Admin RBAC is NOT implemented** - Currently all authenticated users can access admin endpoints. Need to add role checking.

2. **Email notifications NOT implemented** - Refunds don't send email to customers yet.

3. **Midtrans refund API** - Implemented but needs testing with real transactions (sandbox has limitations).

4. **Frontend admin panel NOT implemented** - All admin features are API-only, no UI.

5. **Rate limiting NOT implemented** - Admin APIs have no rate limits.

---

## HOW TO RUN

```bash
# 1. Run migration
cd database
psql -U postgres -d zavera -f migrate_hardening.sql

# 2. Start backend
cd backend
go run main.go

# 3. Start frontend
cd frontend
npm run dev
```

---

*Document generated: January 10, 2026*
*Phase 1 Implementation by: Kiro AI*
