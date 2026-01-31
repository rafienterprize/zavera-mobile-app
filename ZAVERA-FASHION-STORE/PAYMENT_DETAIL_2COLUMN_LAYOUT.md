# Payment Detail Page - 2 Column Layout Redesign

## Problem
Current payment detail page has vertical layout that requires too much scrolling. Left and right sides are empty, wasting screen space.

## Solution
Redesign to 2-column layout:
- **Left Column**: Payment info (QR Code/VA Number)
- **Right Column**: Order summary (receipt style) + Payment instructions + Action buttons

## Benefits
1. Less scrolling required
2. Better use of screen space
3. All important info visible at once
4. More professional look (like Tokopedia/Shopee)

## Implementation Plan

### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│                    Header (Back + Title)                 │
├──────────────────────┬──────────────────────────────────┤
│  LEFT COLUMN         │  RIGHT COLUMN                     │
│                      │                                   │
│  ┌────────────────┐  │  ┌────────────────────────────┐  │
│  │ Status Card    │  │  │ Order Summary (Receipt)    │  │
│  │ - Icon         │  │  │ - Order Code               │  │
│  │ - Status       │  │  │ - Items List               │  │
│  │ - Countdown    │  │  │ - Subtotal                 │  │
│  └────────────────┘  │  │ - Shipping Cost            │  │
│                      │  │ - Total                    │  │
│  ┌────────────────┐  │  └────────────────────────────┘  │
│  │ Payment Info   │  │                                   │
│  │ - Bank Logo    │  │  ┌────────────────────────────┐  │
│  │ - QR Code/VA   │  │  │ Payment Instructions       │  │
│  │ - Amount       │  │  │ - Accordion style          │  │
│  │ - Copy Button  │  │  │ - Step by step             │  │
│  └────────────────┘  │  └────────────────────────────┘  │
│                      │                                   │
│  ┌────────────────┐  │  ┌────────────────────────────┐  │
│  │ Important Note │  │  │ Action Buttons             │  │
│  └────────────────┘  │  │ - Check Status             │  │
│                      │  │ - View Order Detail        │  │
│                      │  └────────────────────────────┘  │
└──────────────────────┴──────────────────────────────────┘
```

### Key Changes

1. **Grid Layout**: `grid lg:grid-cols-2 gap-6`
2. **Left Column**: Payment-specific info (compact)
3. **Right Column**: Order details + Instructions + Actions
4. **Responsive**: Stack vertically on mobile

### Order Summary (Receipt Style)
```tsx
<div className="bg-white rounded-2xl shadow-sm border border-accent">
  <div className="p-6 border-b border-accent bg-secondary/30">
    <h3>Ringkasan Belanja</h3>
    <p>Order: {order_code}</p>
  </div>
  
  <div className="p-6 space-y-4">
    {/* Items */}
    {items.map(item => (
      <div className="flex gap-3">
        <img src={item.image} className="w-16 h-16" />
        <div className="flex-1">
          <p className="font-medium">{item.name}</p>
          <p className="text-sm text-muted">{item.quantity}x</p>
        </div>
        <p className="font-medium">Rp{item.price}</p>
      </div>
    ))}
    
    {/* Breakdown */}
    <div className="space-y-2 pt-4 border-t">
      <div className="flex justify-between text-sm">
        <span>Subtotal ({item_count} barang)</span>
        <span>Rp{subtotal}</span>
      </div>
      <div className="flex justify-between text-sm">
        <span>Ongkos Kirim</span>
        <span>Rp{shipping_cost}</span>
      </div>
    </div>
    
    {/* Total */}
    <div className="pt-4 border-t-2 border-primary">
      <div className="flex justify-between">
        <span className="text-lg font-bold">Total Pembayaran</span>
        <span className="text-2xl font-serif font-bold text-primary">
          Rp{total}
        </span>
      </div>
    </div>
  </div>
</div>
```

### Responsive Behavior
- **Desktop (lg+)**: 2 columns side by side
- **Tablet/Mobile**: Stack vertically (payment info on top, order summary below)

## Files to Modify
1. `frontend/src/app/checkout/payment/detail/page.tsx` - Main layout
2. Need to fetch order items data from backend

## Backend API Needed
Current `/api/payments/core/{order_id}` only returns payment info.
Need to also return order items for receipt display.

### Enhanced Response
```json
{
  "payment_id": 123,
  "order_id": 456,
  "order_code": "ZVR-20260124-ABC",
  "payment_method": "bca_va",
  "bank": "bca",
  "va_number": "1234567890",
  "amount": 309000,
  "expiry_time": "2026-01-25T10:00:00Z",
  "status": "PENDING",
  "order_details": {
    "items": [
      {
        "product_name": "Denim Jacket",
        "product_image": "https://...",
        "quantity": 2,
        "price_per_unit": 150000,
        "subtotal": 300000
      }
    ],
    "subtotal": 300000,
    "shipping_cost": 9000,
    "total": 309000,
    "shipping_address": {
      "recipient_name": "John Doe",
      "phone": "08123456789",
      "full_address": "Jl. Example No. 123"
    }
  }
}
```

## Next Steps
1. Update backend to include order items in payment response
2. Redesign frontend with 2-column layout
3. Test responsive behavior
4. Ensure all payment methods (VA, GoPay, QRIS) work correctly
