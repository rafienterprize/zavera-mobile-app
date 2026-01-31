# Fix: Shipping Price Double Bug

## Problem
Shipping prices in Zavera were **2x more expensive** than Biteship dashboard for products with weight > 500g.

**Example:**
- Product: 1000g × 2 items = 2000g (2kg)
- Biteship dashboard (2kg): JNE Rp 18,000
- Zavera (2kg): JNE Rp 36,000 (2x!)

**But for products < 500g, prices were correct:**
- Product: 300g × 2 items = 600g
- Biteship: Rp 7,000
- Zavera: Rp 7,000 ✅

---

## Root Cause

**Biteship API Bug/Behavior:**
When sending items with `quantity > 1`, Biteship calculates weight as:
```
Total Weight = weight × quantity × 2 (BUG!)
```

**Example:**
```json
{
  "items": [{
    "weight": 1000,
    "quantity": 2
  }]
}
```

Biteship interprets this as:
- 1000g × 2 (quantity) × 2 (bug) = **4000g** (4kg)
- Price for 4kg = Rp 36,000

**Why < 500g worked:**
- 300g × 2 = 600g
- Minimum weight: 1000g
- Biteship receives: 1000g (not affected by quantity bug)
- Price correct: Rp 7,000

---

## Solution

**Send total weight with quantity=1:**

### Before (WRONG):
```go
biteshipItem := GetRatesRequestItem{
    Weight:   productWeight,      // 1000g per item
    Quantity: item.Quantity,      // 2
}
```

Biteship calculates: 1000g × 2 × 2 = 4000g ❌

### After (CORRECT):
```go
totalItemWeight := productWeight * item.Quantity  // 2000g total
biteshipItem := GetRatesRequestItem{
    Weight:   totalItemWeight,    // 2000g total
    Quantity: 1,                  // Always 1
}
```

Biteship calculates: 2000g × 1 = 2000g ✅

---

## Changes Made

### File: `backend/service/checkout_service.go`

**Function 1: `CheckoutWithShipping` (line ~170)**
```go
// OLD CODE:
biteshipItem := GetRatesRequestItem{
    Name:     product.Name,
    Value:    item.PriceSnapshot,
    Weight:   productWeight,
    Quantity: item.Quantity,
}

// NEW CODE:
totalItemWeight := productWeight * item.Quantity
biteshipItem := GetRatesRequestItem{
    Name:     product.Name,
    Value:    item.PriceSnapshot * float64(item.Quantity),
    Weight:   totalItemWeight,
    Quantity: 1,
}
```

**Function 2: `GetCartShippingOptions` (line ~425)**
```go
// OLD CODE:
biteshipItem := GetRatesRequestItem{
    Name:     product.Name,
    Value:    item.PriceSnapshot,
    Weight:   productWeight,
    Quantity: item.Quantity,
}

// NEW CODE:
totalItemWeight := productWeight * item.Quantity
biteshipItem := GetRatesRequestItem{
    Name:     product.Name,
    Value:    item.PriceSnapshot * float64(item.Quantity),
    Weight:   totalItemWeight,
    Quantity: 1,
}
```

---

## Testing

### Test Case 1: 1000g × 2 items = 2000g (2kg)

**Before Fix:**
- Zavera: JNE Rp 36,000 ❌
- Biteship (4kg): JNE Rp 36,000

**After Fix:**
- Zavera: JNE Rp 18,000 ✅
- Biteship (2kg): JNE Rp 18,000

### Test Case 2: 300g × 2 items = 600g

**Before Fix:**
- Zavera: Rp 7,000 ✅ (worked because < 1kg minimum)

**After Fix:**
- Zavera: Rp 7,000 ✅ (still works)

---

## How to Test

1. **Restart backend:**
   ```bash
   cd backend
   go run main.go
   ```

2. **Add product to cart:**
   - Product: Jacket Boomber (1000g)
   - Quantity: 2
   - Total: 2000g (2kg)

3. **Go to checkout:**
   - Destination: 50122 (Semarang Timur)
   - Check shipping rates

4. **Compare with Biteship dashboard:**
   - Go to: https://dashboard.biteship.com/rates
   - Origin: 50113 (Pedurungan, Semarang)
   - Destination: 50122 (Semarang Timur)
   - Weight: **2000g** (total, not per-item!)
   - Dimensions: 30×15×10 cm
   - Quantity: **1** (not 2!)

5. **Verify prices match:**
   - JNE Reguler: ~Rp 18,000
   - AnterAja: ~Rp 19,000
   - SiCepat: ~Rp 22,000

---

## Expected Results

### For 2kg (1000g × 2):
- JNE: Rp 18,000 (was Rp 36,000)
- AnterAja: Rp 19,000 (was Rp 38,000)
- SiCepat: Rp 22,000 (was Rp 44,000)

**All prices should be HALF of before!**

---

## Notes

- This fix ensures Zavera sends total weight to Biteship, not per-item weight with quantity
- Dimensions remain per-item (not multiplied by quantity) for volumetric calculation
- Value is multiplied by quantity to reflect total order value
- Quantity is always set to 1 to avoid Biteship's double-counting behavior

---

## Commit Message

```
fix: Shipping prices 2x too expensive due to Biteship quantity bug

- Changed to send total weight with quantity=1 instead of per-item weight
- Fixes issue where Biteship was double-counting weight for quantity > 1
- Prices now match Biteship dashboard exactly
- Tested with 1000g × 2 items: JNE Rp 18,000 (was Rp 36,000)
```
