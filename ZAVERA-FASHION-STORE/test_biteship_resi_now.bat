@echo off
echo ========================================
echo TEST BITESHIP AUTO-RESI
echo ========================================
echo.

echo Step 1: Cek backend running...
tasklist | findstr zavera
if errorlevel 1 (
    echo [ERROR] Backend tidak running!
    echo Jalankan: cd backend ^&^& .\zavera_RESI_PLACED_FIX.exe
    pause
    exit /b 1
)
echo [OK] Backend running
echo.

echo Step 2: Cek order terbaru...
psql -U postgres -d zavera_db -c "SELECT o.order_code, o.status, s.biteship_draft_order_id FROM orders o LEFT JOIN shipments s ON o.id = s.order_id ORDER BY o.created_at DESC LIMIT 1;"
echo.

echo Step 3: Instruksi Test
echo ========================================
echo UNTUK ORDER LAMA (ZVR-20260129-DA7C29C4):
echo - Kemungkinan GAGAL (draft order stuck "placed")
echo - Coba test, tapi jangan expect berhasil
echo.
echo UNTUK ORDER BARU (RECOMMENDED):
echo 1. Buka http://localhost:3000
echo 2. Login customer
echo 3. Checkout dengan kurir apa saja
echo 4. Bayar order
echo 5. Admin pack order
echo 6. Admin kirim dengan resi KOSONG
echo 7. Cek modal muncul dengan resi
echo.
echo VERIFIKASI RESI:
echo - Format REAL: JNE1234567890 (no dash)
echo - Format DUMMY: JNE-123-xxx (with dash)
echo ========================================
echo.

pause
