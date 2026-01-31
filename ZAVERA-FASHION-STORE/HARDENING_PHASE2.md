# ZAVERA Shipping & Fulfillment Hardening - Phase 2 Complete

## Project Context

ZAVERA adalah e-commerce fashion store dengan stack:
- **Backend**: Go (Gin framework)
- **Frontend**: Next.js + TypeScript
- **Database**: PostgreSQL
- **Payment Gateway**: Midtrans (Snap)
- **Shipping**: RajaOngkir API integration

---

## PHASE 2 COMPLETED ✅

### What Was Implemented

#### 1. ENHANCED SHIPMENT STATE MACHINE

**15 Shipment Statuses:**
```
PENDING              → Waiting for payment/processing
PROCESSING           → Payment received, preparing package
PICKUP_SCHEDULED     → Courier pickup scheduled
PICKUP_FAILED        → Courier failed to pickup
SHIPPED              → Handed to courier
IN_TRANSIT           → On the way
OUT_FOR_DELIVERY     → Out for delivery
DELIVERED            → Successfully delivered
DELIVERY_FAILED      → Delivery attempt failed
HELD_AT_WAREHOUSE    → Held at courier warehouse
RETURNED_TO_SENDER   → Returned to sender
LOST                 → Package lost
INVESTIGATION        → Under investigation
REPLACED             → Replaced with new shipment
CANCELLED            → Cancelled
```

**Transition Validation:**
- Valid transitions defined in `shipment_status_transitions` table
- Admin-only transitions enforced
- Illegal transitions blocked
- Full status history tracking

---

#### 2. DATABASE SCHEMA (migrate_shipping_hardening.sql)

**New Tables:**
```sql
-- Shipment status history (audit trail)
CREATE TABLE shipment_status_history (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id),
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(100),  -- 'system', 'webhook', 'admin:email', 'cron'
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Courier failure log
CREATE TABLE courier_failure_log (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id),
    failure_type VARCHAR(50) NOT NULL,  -- 'pickup_failed', 'delivery_failed', 'lost'
    failure_reason TEXT,
    courier_code VARCHAR(50),
    courier_name VARCHAR(100),
    resolved BOOLEAN DEFAULT false,
    resolution_action TEXT,
    ...
);

-- Disputes table
CREATE TABLE disputes (
    id SERIAL PRIMARY KEY,
    dispute_code VARCHAR(50) UNIQUE NOT NULL,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    shipment_id INTEGER REFERENCES shipments(id),
    dispute_type dispute_type NOT NULL,
    status dispute_status DEFAULT 'OPEN',
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    customer_claim TEXT,
    evidence_urls TEXT[],
    investigation_notes TEXT,
    resolution dispute_status,
    resolution_notes TEXT,
    resolution_amount DECIMAL(12, 2),
    ...
);

-- Dispute messages (communication thread)
CREATE TABLE dispute_messages (
    id SERIAL PRIMARY KEY,
    dispute_id INTEGER NOT NULL REFERENCES disputes(id),
    sender_type VARCHAR(20) NOT NULL,  -- 'customer', 'admin', 'system'
    sender_id INTEGER REFERENCES users(id),
    message TEXT NOT NULL,
    attachment_urls TEXT[],
    is_internal BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Shipment alerts
CREATE TABLE shipment_alerts (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id),
    alert_type VARCHAR(50) NOT NULL,
    alert_level VARCHAR(20) NOT NULL,  -- 'warning', 'critical', 'urgent'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    acknowledged BOOLEAN DEFAULT false,
    resolved BOOLEAN DEFAULT false,
    auto_action_taken BOOLEAN DEFAULT false,
    ...
);

-- Valid status transitions
CREATE TABLE shipment_status_transitions (
    id SERIAL PRIMARY KEY,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    requires_admin BOOLEAN DEFAULT false,
    auto_allowed BOOLEAN DEFAULT true,
    UNIQUE(from_status, to_status)
);
```

**New Columns on Shipments:**
```sql
-- Pickup control
pickup_scheduled_at TIMESTAMP
pickup_deadline TIMESTAMP
pickup_attempts INTEGER DEFAULT 0
last_pickup_attempt_at TIMESTAMP
pickup_notes TEXT

-- Tracking control
last_tracking_update TIMESTAMP
days_without_update INTEGER DEFAULT 0
tracking_stale BOOLEAN DEFAULT false

-- Investigation
investigation_opened_at TIMESTAMP
investigation_reason TEXT
marked_lost_at TIMESTAMP
lost_reason TEXT

-- Delivery control
delivery_attempts INTEGER DEFAULT 0
last_delivery_attempt_at TIMESTAMP
delivery_notes TEXT
recipient_name_confirmed VARCHAR(255)
delivery_photo_url TEXT

-- Reship tracking
reship_count INTEGER DEFAULT 0
original_shipment_id INTEGER
is_replacement BOOLEAN DEFAULT false
reship_reason TEXT
replaced_by_shipment_id INTEGER
reship_cost DECIMAL(12, 2)
reship_cost_bearer VARCHAR(50)  -- 'company' or 'customer'

-- Admin control
requires_admin_action BOOLEAN DEFAULT false
admin_action_reason TEXT
status_metadata JSONB
```

---

#### 3. BACKEND SERVICES

**Files Created:**

| File | Purpose |
|------|---------|
| `backend/models/dispute.go` | Dispute, DisputeMessage, CourierFailureLog, ShipmentStatusHistory, ShipmentAlert models |
| `backend/dto/shipping_hardening_dto.go` | All DTOs for fulfillment APIs |
| `backend/repository/dispute_repository.go` | Dispute, messages, alerts, failures, history CRUD |
| `backend/service/fulfillment_service.go` | Status management, investigation, mark lost, reship |
| `backend/service/dispute_service.go` | Dispute CRUD, investigation, resolution |
| `backend/service/shipment_monitor_service.go` | Automated detection jobs |
| `backend/handler/fulfillment_handler.go` | All fulfillment API handlers |

**Updated Files:**

| File | Changes |
|------|---------|
| `backend/models/shipping.go` | Added 15 statuses, transition validation, helper methods |
| `backend/routes/routes.go` | Added all Phase 2 routes |

---

#### 4. API ENDPOINTS

**Shipment Control:**
```
POST /api/admin/shipments/:id/investigate      # Open investigation
POST /api/admin/shipments/:id/mark-lost        # Mark as lost
POST /api/admin/shipments/:id/reship           # Create replacement shipment
POST /api/admin/shipments/:id/override-status  # Admin status override
POST /api/admin/shipments/:id/schedule-pickup  # Schedule courier pickup
POST /api/admin/shipments/:id/mark-shipped     # Mark as shipped
GET  /api/admin/shipments/:id/details          # Enhanced shipment details
GET  /api/admin/shipments/stuck                # Get stuck shipments
GET  /api/admin/shipments/pickup-failures      # Get pickup failures
```

**Dispute Management:**
```
POST /api/admin/disputes                       # Create dispute
GET  /api/admin/disputes/open                  # Get open disputes
GET  /api/admin/disputes/:id                   # Get dispute by ID
GET  /api/admin/disputes/code/:code            # Get dispute by code
POST /api/admin/disputes/:id/investigate       # Start investigation
POST /api/admin/disputes/:id/request-evidence  # Request evidence
POST /api/admin/disputes/:id/resolve           # Resolve dispute
POST /api/admin/disputes/:id/close             # Close dispute
POST /api/admin/disputes/:id/messages          # Add message
GET  /api/admin/disputes/:id/messages          # Get messages
```

**Fulfillment Dashboard:**
```
GET  /api/admin/fulfillment/dashboard          # Overview dashboard
POST /api/admin/fulfillment/run-monitors       # Trigger monitoring jobs
```

---

#### 5. AUTOMATED MONITORING JOBS

**Stuck Shipment Detector:**
- Updates `days_without_update` for all active shipments
- 7 days no update → Auto-move to INVESTIGATION
- 14 days no update → Auto-mark as LOST
- Creates alerts and disputes automatically

**Lost Shipment Detector:**
- Finds shipments in INVESTIGATION for 7+ days
- Auto-marks as LOST if no resolution

**Pickup Failure Detector:**
- Detects shipments past pickup deadline
- Auto-marks as PICKUP_FAILED
- 3+ failures → Requires admin intervention
- Creates alerts for each failure

Enable via `.env`:
```
ENABLE_SHIPMENT_MONITOR=true
```

---

#### 6. DISPUTE SYSTEM

**Dispute Types:**
```
LOST_PACKAGE      → Package lost in transit
DAMAGED_PACKAGE   → Package damaged
WRONG_ITEM        → Wrong item received
MISSING_ITEM      → Item missing from package
NOT_DELIVERED     → Marked delivered but not received
LATE_DELIVERY     → Significantly late delivery
FAKE_DELIVERY     → Fake delivery confirmation
OTHER             → Other issues
```

**Dispute Statuses:**
```
OPEN              → Dispute opened
INVESTIGATING     → Under investigation
EVIDENCE_REQUIRED → Waiting for evidence
PENDING_RESOLUTION → Waiting for decision
RESOLVED_REFUND   → Resolved with refund
RESOLVED_RESHIP   → Resolved with reship
RESOLVED_REJECTED → Dispute rejected
CLOSED            → Closed
```

**Resolution Actions:**
- Auto-create refund on RESOLVED_REFUND
- Auto-create reship on RESOLVED_RESHIP
- Link refund/reship to dispute record

---

#### 7. RESHIP ENGINE

**Features:**
- Creates replacement shipment from existing order
- Marks original shipment as REPLACED
- Tracks cost bearer (company/customer)
- Links original and replacement shipments
- Increments reship count

**Valid Reship From:**
- LOST
- RETURNED_TO_SENDER
- INVESTIGATION

---

## SAFETY FEATURES

| Feature | Implementation |
|---------|----------------|
| Status Validation | Transitions validated against allowed list |
| Admin-Only Actions | Certain transitions require admin flag |
| Full Audit Trail | All status changes logged with metadata |
| Courier Failure Log | All failures tracked for analysis |
| Alert System | Critical issues create alerts |
| Auto-Detection | Background jobs detect problems |
| Dispute Tracking | Full communication thread preserved |
| No Deletions | Records never deleted |

---

## HOW TO RUN

```bash
# 1. Run Phase 2 migration
migrate_shipping_hardening.bat

# Or manually:
psql -U postgres -d zavera -f database/migrate_shipping_hardening.sql

# 2. Enable monitoring in .env
ENABLE_SHIPMENT_MONITOR=true

# 3. Start backend
cd backend
go run main.go
```

---

## ENVIRONMENT VARIABLES

```env
# Existing variables...

# Phase 2 - Shipment Monitoring
ENABLE_SHIPMENT_MONITOR=true   # Enable background monitoring jobs
```

---

## WHAT'S NEXT (Phase 3+)

### Phase 3 - Customer Experience
- [ ] Customer dispute portal (self-service)
- [ ] Real-time tracking updates
- [ ] Email notifications for status changes
- [ ] SMS notifications for delivery

### Phase 4 - Courier Integration
- [ ] Direct courier API integration (JNE, J&T, etc.)
- [ ] Auto-pickup scheduling
- [ ] Real-time tracking sync
- [ ] Courier performance analytics

### Phase 5 - Analytics
- [ ] Fulfillment performance dashboard
- [ ] Courier comparison reports
- [ ] Dispute analytics
- [ ] Lost package trends

---

## FILE STRUCTURE (Phase 2 Additions)

```
backend/
├── models/
│   ├── shipping.go        # [UPDATED] 15 statuses + validation
│   └── dispute.go         # [NEW] Dispute models
├── repository/
│   └── dispute_repository.go  # [NEW]
├── service/
│   ├── fulfillment_service.go      # [NEW]
│   ├── dispute_service.go          # [NEW]
│   └── shipment_monitor_service.go # [NEW]
├── handler/
│   └── fulfillment_handler.go      # [NEW]
├── dto/
│   └── shipping_hardening_dto.go   # [NEW]
└── routes/
    └── routes.go                   # [UPDATED]

database/
└── migrate_shipping_hardening.sql  # [NEW]

migrate_shipping_hardening.bat      # [NEW]
```

---

*Document generated: January 10, 2026*
*Phase 2 Implementation by: Kiro AI*
