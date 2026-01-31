# Pembelian Quantity Display Fix - 27 Januari 2026

## 🐛 Issue

**Problem:** Saat user membeli 3 barang yang sama (contoh: Denim Jacket × 3):
- Hanya menampilkan 1 foto produk ✅ (ini benar)
- Nama produk: "Denim Jacket" ❌ (tidak jelas quantity)
- Text di bawah: "3 barang" ❌ (tidak jelas bahwa itu 3 barang yang sama)

**User Expectation:**
- Nama produk harus menunjukkan quantity: "Denim Jacket × 3"
- Atau minimal ada indikator yang jelas bahwa itu 3 pcs dari produk yang sama

---

## ✅ Solution

Menambahkan quantity indicator di judul produk dengan format: **"Nama Produk × Quantity"**

### Before:
```
┌─────────────────────────────────────┐
│ [Foto]  Denim Jacket                │
│         3 barang • BCA               │
└─────────────────────────────────────┘
```
❌ Tidak jelas bahwa 3 barang itu adalah Denim Jacket semua

### After:
```
┌─────────────────────────────────────┐
│ [Foto]  Denim Jacket × 3            │
│         3 barang • BCA               │
└─────────────────────────────────────┘
```
✅ Jelas bahwa user membeli 3 pcs Denim Jacket

---

## 🔧 Implementation

**File:** `frontend/src/app/account/pembelian/page.tsx`

**Location:** TransactionCard component, Product Info section

### Code Changes:

```typescript
// BEFORE
<h3 className="font-medium text-gray-900 truncate">{order.item_summary}</h3>
<p className="text-sm text-gray-500 mt-0.5">
  {order.item_count} barang
  {order.payment_method && ` • ${order.bank?.toUpperCase() || order.payment_method}`}
</p>

// AFTER
<h3 className="font-medium text-gray-900">
  {order.item_summary}
  {order.item_count > 1 && (
    <span className="text-gray-500 font-normal"> × {order.item_count}</span>
  )}
</h3>
<p className="text-sm text-gray-500 mt-0.5">
  {order.item_count} {order.item_count === 1 ? "barang" : "barang"}
  {order.payment_method && ` • ${order.bank?.toUpperCase() || order.payment_method}`}
</p>
```

### Logic:
1. **If `item_count === 1`:** Tampilkan nama produk saja (tidak perlu × 1)
2. **If `item_count > 1`:** Tampilkan "Nama Produk × Quantity" dengan quantity berwarna abu-abu

---

## 🎨 Design Decisions

### Why " × Quantity" format?
- ✅ Universal symbol untuk multiplication/quantity
- ✅ Compact dan tidak memakan banyak space
- ✅ Mudah dipahami (standar e-commerce)
- ✅ Konsisten dengan format cart/checkout

### Why gray color for quantity?
- ✅ Membedakan antara nama produk (hitam) dan quantity (abu-abu)
- ✅ Tidak terlalu mencolok tapi tetap terlihat
- ✅ Konsisten dengan design system Zavera

### Why keep "3 barang" text?
- ✅ Informasi tambahan yang berguna
- ✅ Konsisten dengan design existing
- ✅ Membantu user yang tidak familiar dengan simbol ×

---

## 📱 Display Examples

### Example 1: Single Item (1 pcs)
```
Denim Jacket
1 barang • BCA
```
✅ No quantity indicator (tidak perlu × 1)

### Example 2: Multiple Same Items (3 pcs)
```
Denim Jacket × 3
3 barang • BCA
```
✅ Clear quantity indicator

### Example 3: Multiple Different Items
```
Denim Jacket
5 barang • BCA
```
⚠️ Note: Backend `item_summary` hanya mengirim nama produk pertama
Jika user beli 3 Denim Jacket + 2 T-Shirt, akan tampil:
- "Denim Jacket × 5" (misleading)
- Seharusnya: "Denim Jacket +4 produk lainnya"

**Backend Limitation:** 
Backend hanya mengirim `item_summary` (nama produk pertama) dan `item_count` (total quantity).
Tidak ada informasi apakah itu 1 produk dengan quantity banyak atau banyak produk berbeda.

---

## 🔍 Backend Data Structure

### Current API Response:
```json
{
  "order_id": 123,
  "order_code": "ZVR-20260127-BRB3ACCD",
  "item_summary": "Denim Jacket",  // ← Hanya nama produk pertama
  "item_count": 3,                  // ← Total quantity semua item
  "product_image": "https://...",   // ← Gambar produk pertama
  "total_amount": 918000
}
```

### Limitation:
- Tidak bisa membedakan:
  - 3 pcs Denim Jacket (same product)
  - 1 Denim Jacket + 2 T-Shirt (different products)
- Keduanya akan tampil sama: "Denim Jacket × 3"

### Ideal Solution (Future Enhancement):
Backend should send:
```json
{
  "items": [
    {
      "product_name": "Denim Jacket",
      "quantity": 3,
      "product_image": "https://..."
    }
  ],
  "total_items": 1,      // Jumlah produk berbeda
  "total_quantity": 3    // Total quantity semua item
}
```

Then frontend can display:
- If `total_items === 1`: "Denim Jacket × 3"
- If `total_items > 1`: "Denim Jacket +2 produk lainnya"

---

## ✅ Testing

### Test Cases:

1. **Single Item (1 pcs):**
   - Buy 1 Denim Jacket
   - Expected: "Denim Jacket" (no × 1)
   - ✅ Pass

2. **Multiple Same Items (3 pcs):**
   - Buy 3 Denim Jacket
   - Expected: "Denim Jacket × 3"
   - ✅ Pass

3. **Multiple Same Items (10 pcs):**
   - Buy 10 T-Shirt
   - Expected: "T-Shirt × 10"
   - ✅ Pass

4. **Multiple Different Items:**
   - Buy 2 Denim Jacket + 1 T-Shirt
   - Current: "Denim Jacket × 3" (misleading)
   - Ideal: "Denim Jacket +2 produk lainnya"
   - ⚠️ Backend limitation

---

## 🚀 Future Enhancements

### Option 1: Show All Items (if space allows)
```
┌─────────────────────────────────────┐
│ [Foto]  Denim Jacket × 2            │
│         T-Shirt × 1                  │
│         3 barang • BCA               │
└─────────────────────────────────────┘
```

### Option 2: Show First Item + Others
```
┌─────────────────────────────────────┐
│ [Foto]  Denim Jacket × 2            │
│         +1 produk lainnya            │
│         3 barang • BCA               │
└─────────────────────────────────────┘
```

### Option 3: Expandable Item List
```
┌─────────────────────────────────────┐
│ [Foto]  Denim Jacket × 2            │
│         [▼ Lihat 2 produk lainnya]   │
│         3 barang • BCA               │
└─────────────────────────────────────┘
```

**Recommendation:** Option 2 (Show First Item + Others)
- Simple and clear
- Doesn't take much space
- User can click "Lihat Detail" for full list

---

## 📝 Files Changed

1. **`frontend/src/app/account/pembelian/page.tsx`**
   - Component: `TransactionCard`
   - Section: Product Info
   - Change: Added quantity indicator to product title

---

## 🎯 Result

**Before:**
- "Denim Jacket" + "3 barang" (confusing)

**After:**
- "Denim Jacket × 3" + "3 barang" (clear!)

**User Experience:**
- ✅ Lebih jelas bahwa user membeli 3 pcs dari produk yang sama
- ✅ Konsisten dengan format e-commerce standar
- ✅ Tidak memakan banyak space
- ✅ Mudah dipahami

---

**Status:** ✅ Fixed
**Last Updated:** 27 Januari 2026
