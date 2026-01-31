# Frontend Payment Detail - 2 Column Layout Implementation

## Backend API Response (Already Done ✅)
Backend now returns `order_details` in payment response:
```typescript
interface PaymentDetails {
  payment_id: number;
  order_id: number;
  order_code: string;
  payment_method: string;
  bank: string;
  va_number: string;
  amount: number;
  expiry_time: string;
  status: string;
  // NEW: Order details for receipt
  order_details: {
    items: Array<{
      product_name: string;
      product_image: string;
      quantity: number;
      price_per_unit: number;
      subtotal: number;
    }>;
    subtotal: number;
    shipping_cost: number;
    total: number;
    customer_name: string;
    customer_email: string;
    customer_phone: string;
    shipping_address: string;
    courier_name: string;
    courier_service: string;
  };
}
```

## Frontend Changes Needed

### 1. Update Main Layout (Line ~501)
Change from single column to 2-column grid:

```tsx
<div className="max-w-7xl mx-auto px-6 py-8">
  {/* 2 Column Layout */}
  <div className="grid lg:grid-cols-2 gap-6">
    {/* LEFT COLUMN */}
    <div className="space-y-6">
      {/* Status Card + Payment Info */}
    </div>
    
    {/* RIGHT COLUMN */}
    <div className="space-y-6">
      {/* Order Summary + Instructions + Actions */}
    </div>
  </div>
</div>
```

### 2. Add Order Summary Component (Receipt Style)
Add this new component after `InstructionAccordion`:

```tsx
const OrderSummary = ({ orderDetails }: { orderDetails: PaymentDetails['order_details'] }) => {
  if (!orderDetails) return null;
  
  return (
    <div className="bg-white rounded-2xl shadow-sm border border-accent overflow-hidden">
      {/* Header */}
      <div className="p-6 border-b border-accent bg-secondary/30">
        <h3 className="text-lg font-serif font-bold text-primary">Ringkasan Belanja</h3>
      </div>
      
      {/* Items */}
      <div className="p-6 space-y-4">
        {orderDetails.items.map((item, idx) => (
          <div key={idx} className="flex gap-3">
            <div className="w-16 h-16 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
              {item.product_image ? (
                <Image
                  src={item.product_image}
                  alt={item.product_name}
                  width={64}
                  height={64}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
              )}
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium text-gray-900 truncate">{item.product_name}</p>
              <p className="text-sm text-muted">{item.quantity}x @ Rp{item.price_per_unit.toLocaleString("id-ID")}</p>
            </div>
            <div className="text-right flex-shrink-0">
              <p className="font-medium text-gray-900">Rp{item.subtotal.toLocaleString("id-ID")}</p>
            </div>
          </div>
        ))}
        
        {/* Breakdown */}
        <div className="space-y-2 pt-4 border-t border-accent">
          <div className="flex justify-between text-sm">
            <span className="text-muted">Subtotal ({orderDetails.items.length} barang)</span>
            <span className="text-gray-900">Rp{orderDetails.subtotal.toLocaleString("id-ID")}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted">Ongkos Kirim</span>
            <span className="text-gray-900">Rp{orderDetails.shipping_cost.toLocaleString("id-ID")}</span>
          </div>
          {orderDetails.courier_name && (
            <p className="text-xs text-muted">
              {orderDetails.courier_name} - {orderDetails.courier_service}
            </p>
          )}
        </div>
        
        {/* Total */}
        <div className="pt-4 border-t-2 border-primary">
          <div className="flex items-center justify-between">
            <span className="text-lg font-bold text-primary">Total Pembayaran</span>
            <span className="text-2xl font-serif font-bold text-primary">
              Rp{orderDetails.total.toLocaleString("id-ID")}
            </span>
          </div>
        </div>
        
        {/* Shipping Address */}
        {orderDetails.shipping_address && (
          <div className="pt-4 border-t border-accent">
            <p className="text-xs text-muted uppercase tracking-wider mb-2">Alamat Pengiriman</p>
            <p className="text-sm text-gray-900 font-medium">{orderDetails.customer_name}</p>
            <p className="text-sm text-muted">{orderDetails.customer_phone}</p>
            <p className="text-sm text-muted mt-1">{orderDetails.shipping_address}</p>
          </div>
        )}
      </div>
    </div>
  );
};
```

### 3. Reorganize Layout Structure

**LEFT COLUMN** should contain:
1. Status Card (with countdown)
2. Payment Info (QR/VA/GoPay)
3. Important Notes

**RIGHT COLUMN** should contain:
1. Order Summary (NEW - receipt style)
2. Payment Instructions (accordion)
3. Action Buttons (Check Status + View Order)

### 4. Move Components to Correct Columns

```tsx
{/* LEFT COLUMN */}
<div className="space-y-6">
  {/* Status Card */}
  <motion.div className="bg-white rounded-2xl...">
    {/* Status Header with Icon + Countdown */}
    {/* Payment Details (QR/VA) */}
    {/* Important Notes */}
  </motion.div>
</div>

{/* RIGHT COLUMN */}
<div className="space-y-6">
  {/* Order Summary - NEW */}
  <OrderSummary orderDetails={payment.order_details} />
  
  {/* Payment Instructions */}
  {!isExpired && !isPaid && !isCancelled && (
    <motion.div className="bg-white rounded-2xl...">
      <div className="p-6">
        <h3>Cara Pembayaran</h3>
        <InstructionAccordion instructions={payment.instructions} bank={payment.bank} />
      </div>
    </motion.div>
  )}
  
  {/* Action Buttons */}
  <motion.div className="bg-white rounded-2xl...">
    {/* Check Status + View Order buttons */}
  </motion.div>
</div>
```

### 5. Responsive Behavior
- Desktop (lg+): 2 columns side by side
- Mobile/Tablet: Stack vertically (payment info on top)

```css
grid lg:grid-cols-2 gap-6
```

## Summary of Changes

1. ✅ Backend API updated (includes order_details)
2. ⏳ Frontend layout: Change to 2-column grid
3. ⏳ Add OrderSummary component (receipt style)
4. ⏳ Move payment info to left column
5. ⏳ Move instructions + actions to right column
6. ⏳ Test responsive behavior

## Files to Modify
- `frontend/src/app/checkout/payment/detail/page.tsx` (main file)

## Testing Checklist
- [ ] Desktop view shows 2 columns
- [ ] Mobile view stacks vertically
- [ ] Order items display correctly
- [ ] Subtotal + shipping + total calculations correct
- [ ] Payment instructions still accessible
- [ ] Check status button works
- [ ] All payment methods (VA, GoPay, QRIS) display correctly
