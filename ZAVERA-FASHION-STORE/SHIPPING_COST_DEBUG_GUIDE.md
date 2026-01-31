# Shipping Cost Debug Guide

## Issue
Zavera shipping costs are 2x more expensive than Biteship dashboard for the same route and product dimensions.

**Test Parameters:**
- Origin: 50113 (Pedurungan, Semarang)
- Destination: 50122 (Semarang Timur)
- Product: Denim Jacket
- Weight: 1400g total (2 items × 700g each)
- Dimensions: 30×25×10 cm per item

**Expected (Biteship Dashboard):** Rp 14.000-16.000 (Regular)
**Actual (Zavera):** Rp 21.000-33.000 (almost 2x)

---

## Root Cause Analysis

The backend code is correctly configured to send dimensions to Biteship API. The issue is likely:

1. **Product dimensions not saved in database**
2. **Migration not run** (dimension columns don't exist)
3. **Dimensions are 0 or default values**

---

## Step 1: Check if Migration Was Run

Run this command to check if dimension columns exist:

```bash
psql -U postgres -d zavera_db -c "\d products"
```

Look for these columns:
- `length` (integer)
- `width` (integer)
- `height` (integer)

If they **don't exist**, run the migration:

```bash
psql -U postgres -d zavera_db -f database\migrate_product_dimensions.sql
```

---

## Step 2: Check Product Dimensions in Database

Run this query to see actual product dimensions:

```bash
psql -U postgres -d zavera_db -f database\check_product_dimensions.sql
```

Or manually:

```sql
SELECT 
    id,
    name,
    weight,
    length,
    width,
    height,
    stock,
    price
FROM products
WHERE name LIKE '%Denim%'
ORDER BY id DESC;
```

**Expected values for Denim Jacket:**
- Weight: 700 (grams per item)
- Length: 30 (cm)
- Width: 25 (cm)
- Height: 10 (cm)

If dimensions are **0 or wrong**, you need to update the product in Admin Dashboard.

---

## Step 3: Update Product Dimensions

1. Go to Admin Dashboard → Products
2. Click Edit on "Denim Jacket"
3. Set dimensions:
   - Weight: 700g
   - Length: 30cm
   - Width: 25cm
   - Height: 10cm
4. Click Save

---

## Step 4: Test Checkout with Backend Logs

1. **Start backend** (if not running):
   ```bash
   cd backend
   go run main.go
   ```

2. **Open frontend** and go to checkout

3. **Watch backend terminal** for these log messages:

   ```
   🛒 GetCartShippingOptions - SessionID: xxx, DestPostalCode: 50122
   📦 Cart ID: X, Items count: 1
   📡 Getting Biteship rates - Origin: 50113, Dest: 50122, Weight: 1400, Items: 1
     Item 1: Denim Jacket - Weight: 700g, Dimensions: 30x25x10 cm, Qty: 2
   📤 Biteship API Request [POST /v1/rates/couriers]: {...}
   📥 Biteship Rates Response: {...}
   ✅ Got X shipping rates from Biteship
   ```

4. **Check the log output:**
   - Does it show `Dimensions: 30x25x10 cm`?
   - Or does it show `Dimensions: 0x0x0 cm`?

---

## Step 5: Verify Biteship API Request

If dimensions show as `0x0x0`, the product dimensions are not loaded from database.

If dimensions show correctly (`30x25x10`), then check the Biteship API request payload in the logs:

```json
{
  "origin_postal_code": 50113,
  "destination_postal_code": 50122,
  "couriers": "jne,jnt,sicepat,tiki,anteraja",
  "items": [
    {
      "name": "Denim Jacket",
      "value": 250000,
      "weight": 700,
      "length": 30,
      "width": 25,
      "height": 10,
      "quantity": 2
    }
  ]
}
```

**Verify:**
- `length`, `width`, `height` are present and correct
- `weight` is per-item weight (700g, not 1400g)
- `quantity` is 2

---

## Step 6: Understanding Volumetric Weight

Biteship calculates volumetric weight using:

```
Volumetric Weight = (Length × Width × Height) / 6000
```

For 30×25×10 cm:
```
Volumetric Weight = (30 × 25 × 10) / 6000 = 1.25 kg = 1250g
```

**Biteship uses the HIGHER value:**
- Actual weight: 700g per item × 2 = 1400g
- Volumetric weight: 1250g per item × 2 = 2500g

So Biteship should use **2500g** (volumetric) for shipping calculation, not 1400g (actual).

**This might explain the price difference!**

---

## Step 7: Test with Correct Dimensions

If the issue is volumetric weight, you have two options:

### Option A: Reduce Dimensions
If the product can be packed smaller:
- Length: 30cm → 25cm
- Width: 25cm → 20cm
- Height: 10cm → 8cm

New volumetric: (25×20×8)/6000 = 0.67kg = 670g per item
Total: 670g × 2 = 1340g (close to actual weight 1400g)

### Option B: Accept Higher Shipping Cost
If dimensions are accurate, the higher shipping cost is correct because the package takes up more space in the courier's vehicle.

---

## Step 8: Compare with Biteship Dashboard

When testing in Biteship dashboard, make sure you're entering:

**Per-item values:**
- Weight: 700g (not 1400g)
- Dimensions: 30×25×10 cm
- Quantity: 2

**NOT total values:**
- Weight: 1400g
- Dimensions: 30×25×10 cm
- Quantity: 1

Biteship API expects per-item dimensions and calculates total volumetric weight automatically.

---

## Quick Debug Checklist

- [ ] Migration run (`migrate_product_dimensions.sql`)
- [ ] Dimension columns exist in database
- [ ] Product dimensions saved correctly (30×25×10)
- [ ] Backend logs show correct dimensions
- [ ] Biteship API request includes dimensions
- [ ] Volumetric weight calculated correctly
- [ ] Comparing apples-to-apples with Biteship dashboard

---

## Next Steps

1. Run `check_shipping_debug.bat` to verify database
2. Test checkout and copy backend terminal logs
3. Share the logs so we can verify the Biteship API request
4. Compare volumetric vs actual weight calculation

---

## Expected Outcome

After fixing dimensions, you should see:
- Backend logs showing `Dimensions: 30x25x10 cm`
- Biteship API request with correct dimensions
- Shipping costs matching Biteship dashboard (considering volumetric weight)

If costs are still different, it might be due to:
- Volumetric weight calculation (package size matters!)
- Different courier service types (Regular vs Express)
- Biteship account settings or pricing tier
