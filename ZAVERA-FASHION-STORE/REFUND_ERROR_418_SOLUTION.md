# Solusi Error 418 - Refund Manual Processing

## 🔴 Masalah yang Terjadi

Ketika admin mencoba memproses refund, muncul error:

```
MANUAL_PROCESSING_REQUIRED: Automatic refund failed. 
Please process manual bank transfer to customer and mark refund as completed after transfer is done
```

### Penyebab Error 418

Error ini terjadi karena **Midtrans mengembalikan status code 418** yang artinya:

> **"Payment provider requires additional settlement time before refund can be processed"**

**Kenapa ini terjadi?**
1. **Order terlalu baru** (< 24 jam) - Bank belum settle transaksi
2. **Payment provider (BCA, BNI, dll) butuh waktu settlement** - Biasanya 1-24 jam
3. **Refund otomatis tidak bisa diproses** - Harus tunggu atau proses manual

---

## ✅ Solusi yang Sudah Ditambahkan

Saya sudah menambahkan fitur **Manual Refund Completion** ke backend:

### 1. Backend Changes

#### ✅ Handler Baru: `MarkRefundCompleted`
**File:** `backend/handler/admin_refund_handler.go`

```go
// POST /admin/refunds/:id/mark-completed
func (h *AdminRefundHandler) MarkRefundCompleted(c *gin.Context)
```

Endpoint ini memungkinkan admin untuk:
- Mark refund sebagai COMPLETED setelah transfer bank manual
- Mencatat note/bukti transfer
- Update order status dan restore stock

#### ✅ Service Method Baru
**File:** `backend/service/refund_service.go`

```go
func (s *refundService) MarkRefundCompletedManually(refundID int, processedBy int, note string) error
```

#### ✅ Route Baru
**File:** `backend/routes/routes.go`

```go
admin.POST("/refunds/:id/mark-completed", refundHandler.MarkRefundCompleted)
```

---

## 🎯 Cara Menggunakan (Untuk Admin)

### Opsi 1: Tunggu dan Retry (Untuk Order Baru)

Jika order masih baru (< 24 jam):

1. **Tunggu 2-4 jam** untuk settlement
2. Kembali ke halaman refund
3. Klik tombol **"Retry Refund"**
4. Sistem akan coba proses otomatis lagi

### Opsi 2: Manual Bank Transfer (Untuk Order > 24 jam)

Jika order sudah > 24 jam tapi masih error 418:

1. **Lakukan transfer bank manual** ke rekening customer
2. Catat bukti transfer (nomor referensi, waktu, jumlah)
3. Kembali ke halaman refund
4. Klik tombol **"Mark as Completed"**
5. Masukkan note dengan detail transfer:
   ```
   Transfer manual via BCA
   Nomor Referensi: 1234567890
   Tanggal: 29 Jan 2026 14:30
   Jumlah: Rp 709.000
   Rekening tujuan: BCA 1234567890 a.n. Customer Name
   ```
6. Klik **"Confirm"**

---

## 🔧 Yang Perlu Ditambahkan di Frontend

Sekarang tinggal update UI admin untuk menampilkan opsi manual completion.

### File yang Perlu Diupdate

**File:** `frontend/src/app/admin/orders/[code]/page.tsx`

Tambahkan:

1. **State untuk manual completion dialog:**
```typescript
const [showManualCompleteDialog, setShowManualCompleteDialog] = useState(false);
const [manualNote, setManualNote] = useState("");
```

2. **Function untuk mark as completed:**
```typescript
const handleMarkRefundCompleted = async (refundId: number) => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/admin/refunds/${refundId}/mark-completed`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("admin_token")}`,
        },
        body: JSON.stringify({ note: manualNote }),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to mark refund as completed");
    }

    alert("✅ Refund marked as completed successfully!");
    // Refresh data
    fetchOrderDetail();
    setShowManualCompleteDialog(false);
    setManualNote("");
  } catch (error: any) {
    alert(`❌ Error: ${error.message}`);
  }
};
```

3. **UI Button di Refund Section:**
```tsx
{refund.status === "PENDING" && (
  <div className="mt-4 space-y-2">
    <button
      onClick={() => handleProcessRefund(refund.id)}
      className="w-full px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600"
    >
      🔄 Process Refund (Auto)
    </button>
    
    <button
      onClick={() => setShowManualCompleteDialog(true)}
      className="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
    >
      ✅ Mark as Completed (Manual Transfer)
    </button>
  </div>
)}
```

4. **Dialog untuk input note:**
```tsx
{showManualCompleteDialog && (
  <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div className="bg-white rounded-lg p-6 max-w-md w-full">
      <h3 className="text-lg font-bold mb-4">Mark Refund as Completed</h3>
      
      <p className="text-sm text-gray-600 mb-4">
        ⚠️ Only mark as completed after you have successfully transferred the refund amount to customer's bank account.
      </p>
      
      <label className="block mb-2 text-sm font-medium">
        Transfer Details / Note *
      </label>
      <textarea
        value={manualNote}
        onChange={(e) => setManualNote(e.target.value)}
        placeholder="Example:
Transfer manual via BCA
Ref: 1234567890
Date: 29 Jan 2026 14:30
Amount: Rp 709.000
To: BCA 1234567890 a.n. Customer Name"
        className="w-full border rounded p-2 mb-4 h-32"
        required
      />
      
      <div className="flex gap-2">
        <button
          onClick={() => {
            setShowManualCompleteDialog(false);
            setManualNote("");
          }}
          className="flex-1 px-4 py-2 border rounded hover:bg-gray-100"
        >
          Cancel
        </button>
        <button
          onClick={() => handleMarkRefundCompleted(currentRefund.id)}
          disabled={!manualNote.trim()}
          className="flex-1 px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:bg-gray-300"
        >
          ✅ Confirm Completion
        </button>
      </div>
    </div>
  </div>
)}
```

---

## 📋 Testing Steps

### 1. Build Backend
```bash
cd backend
go build -o zavera.exe
```

### 2. Test Manual Completion API
```bash
# Get refund ID yang PENDING
curl http://localhost:8080/admin/refunds

# Mark as completed
curl -X POST http://localhost:8080/admin/refunds/1/mark-completed \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "note": "Transfer manual via BCA\nRef: 1234567890\nDate: 29 Jan 2026\nAmount: Rp 709.000"
  }'
```

### 3. Verify Database
```sql
-- Check refund status
SELECT id, refund_code, status, gateway_refund_id, gateway_response 
FROM refunds 
WHERE id = 1;

-- Should show:
-- status: COMPLETED
-- gateway_refund_id: MANUAL_BANK_TRANSFER
-- gateway_response: {"manual_completion": true, "note": "...", ...}
```

---

## 🎨 UI Flow Diagram

```
┌─────────────────────────────────────────┐
│  Admin clicks "Process Refund"          │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Backend calls Midtrans API             │
└──────────────┬──────────────────────────┘
               │
               ▼
        ┌──────┴──────┐
        │   Success?   │
        └──────┬───────┘
               │
       ┌───────┴────────┐
       │                │
    ✅ YES            ❌ NO (Error 418)
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────────────────┐
│  COMPLETED  │  │  Show Error Message:     │
└─────────────┘  │  "MANUAL_PROCESSING_     │
                 │   REQUIRED"              │
                 └──────────┬───────────────┘
                            │
                            ▼
                 ┌──────────────────────────┐
                 │  Admin has 2 options:    │
                 │  1. Wait & Retry         │
                 │  2. Manual Transfer      │
                 └──────────┬───────────────┘
                            │
                    ┌───────┴────────┐
                    │                │
              Option 1          Option 2
                    │                │
                    ▼                ▼
         ┌──────────────┐  ┌─────────────────┐
         │ Wait 2-4 hrs │  │ Do bank transfer│
         │ Click Retry  │  │ Click "Mark as  │
         └──────────────┘  │  Completed"     │
                           │ Enter note      │
                           └─────────────────┘
                                    │
                                    ▼
                           ┌─────────────────┐
                           │   COMPLETED     │
                           │ (Manual)        │
                           └─────────────────┘
```

---

## 📝 Summary

**Backend:** ✅ SELESAI
- Handler untuk mark as completed
- Service method untuk manual completion
- Route sudah ditambahkan
- Error handling untuk 418

**Frontend:** ⏳ PERLU UPDATE
- Tambahkan button "Mark as Completed"
- Tambahkan dialog untuk input note
- Tambahkan function untuk call API

**Next Steps:**
1. Update frontend UI sesuai contoh di atas
2. Build dan test
3. Deploy

---

## 🔗 Related Files

- `backend/handler/admin_refund_handler.go` - Handler
- `backend/service/refund_service.go` - Service logic
- `backend/routes/routes.go` - Routes
- `frontend/src/app/admin/orders/[code]/page.tsx` - Admin UI (needs update)
