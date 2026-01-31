# ZAVERA Database Maintenance Guide

## Overview
Dokumen ini menjelaskan prosedur maintenance database untuk sistem ZAVERA e-commerce.

## Cleanup Jobs

### 1. Expire Pending Orders (Wajib - Jalankan setiap jam)
Order yang PENDING lebih dari 24 jam akan di-expire dan stock dikembalikan.

```sql
SELECT expire_pending_orders();
```

**Cron schedule:** `0 * * * *` (setiap jam)

### 2. Cleanup Abandoned Carts (Jalankan harian)
Menghapus cart yang tidak diupdate selama 30 hari.

```sql
SELECT cleanup_abandoned_carts();
```

**Cron schedule:** `0 3 * * *` (jam 3 pagi setiap hari)

### 3. Cleanup Expired Tokens (Jalankan harian)
Menghapus token verifikasi email dan reset password yang sudah expired.

```sql
SELECT cleanup_expired_tokens();
```

**Cron schedule:** `0 4 * * *` (jam 4 pagi setiap hari)

## Monitoring Views

### Order Summary
```sql
SELECT * FROM v_order_summary WHERE order_date >= CURRENT_DATE - INTERVAL '7 days';
```

### Pending Orders Alert
```sql
SELECT * FROM v_pending_orders_alert;
```

## Manual Maintenance

### Check Stock Integrity
```sql
-- Cek apakah ada stock negatif
SELECT id, name, stock FROM products WHERE stock < 0;

-- Cek order dengan stock_reserved yang salah
SELECT id, order_code, status, stock_reserved 
FROM orders 
WHERE (status IN ('CANCELLED', 'FAILED', 'EXPIRED') AND stock_reserved = true)
   OR (status IN ('PENDING', 'PAID', 'PROCESSING', 'SHIPPED') AND stock_reserved = false);
```

### Fix Stock Reserved Flag
```sql
-- Jalankan jika ada inkonsistensi
UPDATE orders SET stock_reserved = false 
WHERE status IN ('CANCELLED', 'FAILED', 'EXPIRED', 'COMPLETED', 'DELIVERED') 
AND stock_reserved = true;
```

## Migration Order
Jalankan migrasi dalam urutan berikut:
1. `schema.sql` - Base schema
2. `migrate_auth.sql` - Authentication tables
3. `migrate_phase2.sql` - Order status enhancements
4. `migrate_shipping.sql` - Shipping system
5. `migrate_categories.sql` - Product categories
6. `migrate_fixes.sql` - Production fixes & cleanup functions

## Backup Recommendations
- Full backup: Daily
- Transaction log backup: Every 15 minutes
- Retention: 30 days minimum
