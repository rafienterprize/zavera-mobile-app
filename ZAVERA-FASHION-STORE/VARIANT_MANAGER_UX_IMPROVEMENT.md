# Variant Manager UX Improvement

## Current Issues (Screenshot yang Anda tunjukkan):

1. ❌ **SKU** - Text input manual (harusnya auto-generate)
2. ❌ **Variant Name** - Text input manual (harusnya auto-generate)
3. ❌ **Size** - Text input manual (harusnya dropdown)
4. ❌ **Color** - Dropdown tapi masih perlu input manual color hex
5. ❌ **Color Hex** - Manual input (harusnya auto-fill dari color)

## Improvements to Implement:

### 1. Size Dropdown
```tsx
const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];

<select value={formData.size} onChange={handleSizeChange}>
  <option value="">Select Size</option>
  {SIZES.map(size => (
    <option key={size} value={size}>{size}</option>
  ))}
</select>
```

### 2. Color Dropdown with Auto Hex
```tsx
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  // ... more colors
];

<select value={formData.color} onChange={handleColorChange}>
  <option value="">Select Color</option>
  {COLORS.map(color => (
    <option key={color.name} value={color.name}>{color.name}</option>
  ))}
</select>

// Auto-fill color_hex when color selected
const handleColorChange = (colorName: string) => {
  const color = COLORS.find(c => c.name === colorName);
  setFormData({
    ...formData,
    color: colorName,
    color_hex: color?.hex || '',
  });
};
```

### 3. Auto-Generate SKU
```tsx
// Generate SKU from product name + size + color
const generateSKU = (productName: string, size: string, color: string) => {
  const sanitize = (str: string) => str.toUpperCase().replace(/[^A-Z0-9]/g, '-');
  return `${sanitize(productName)}-${sanitize(size)}-${sanitize(color)}`;
};

// Auto-generate when size or color changes
useEffect(() => {
  if (formData.size && formData.color) {
    const sku = generateSKU(productName, formData.size, formData.color);
    setFormData(prev => ({ ...prev, sku }));
  }
}, [formData.size, formData.color]);
```

### 4. Auto-Generate Variant Name
```tsx
// Generate variant name from size + color
const generateVariantName = (size: string, color: string) => {
  return `${size} - ${color}`;
};

// Auto-generate when size or color changes
useEffect(() => {
  if (formData.size && formData.color) {
    const variantName = generateVariantName(formData.size, formData.color);
    setFormData(prev => ({ ...prev, variant_name: variantName }));
  }
}, [formData.size, formData.color]);
```

### 5. Hide Auto-Generated Fields
```tsx
// Don't show SKU and Variant Name inputs
// They are auto-generated in the background

// Only show:
- Size (dropdown)
- Color (dropdown)
- Stock Quantity
- Price Override (optional)
- Shipping Dimensions (optional, collapsible)
```

## New UI Flow:

### Add Variant Form:
```
┌─────────────────────────────────────┐
│ Add New Variant                     │
├─────────────────────────────────────┤
│                                     │
│ Size *                              │
│ [Dropdown: XS, S, M, L, XL, ...]   │
│                                     │
│ Color *                             │
│ [Dropdown: Black, White, Navy, ...] │
│ ● #000000 (auto-filled)            │
│                                     │
│ Stock Quantity *                    │
│ [10                              ]  │
│                                     │
│ Price Override (optional)           │
│ [Leave empty to use product price]  │
│                                     │
│ ▼ Shipping Dimensions (optional)    │
│   Weight: [400] g                   │
│   Length: [30] cm                   │
│   Width:  [20] cm                   │
│   Height: [5] cm                    │
│                                     │
│ [Cancel]  [Add Variant]            │
└─────────────────────────────────────┘

Auto-generated (hidden):
- SKU: PRODUCT-NAME-M-BLACK
- Variant Name: M - Black
```

## Benefits:

1. ✅ **Faster Input** - Dropdown lebih cepat dari typing
2. ✅ **No Typos** - Dropdown prevents typos
3. ✅ **Consistent Data** - All sizes/colors standardized
4. ✅ **Auto SKU** - No need to think about SKU format
5. ✅ **Auto Color Hex** - No need to remember hex codes
6. ✅ **Cleaner UI** - Hide technical fields (SKU, variant_name)

## Implementation Steps:

1. ✅ Define SIZES and COLORS constants
2. ✅ Create handleSizeChange and handleColorChange
3. ✅ Add auto-generate logic for SKU and variant_name
4. ✅ Update form UI to use dropdowns
5. ✅ Hide SKU and variant_name inputs
6. ✅ Make shipping dimensions collapsible
7. ✅ Add visual feedback (color preview)

## Code Changes Needed:

### File: `frontend/src/components/admin/VariantManager.tsx`

**Add Constants:**
```tsx
const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' },
  { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' },
  { name: 'Green', hex: '#008000' },
  { name: 'Gray', hex: '#808080' },
  { name: 'Pink', hex: '#FFC0CB' },
];
```

**Add Helper Functions:**
```tsx
const generateSKU = (productName: string, size: string, color: string) => {
  const sanitize = (str: string) => str.toUpperCase().replace(/[^A-Z0-9]/g, '-');
  return `${sanitize(productName)}-${sanitize(size)}-${sanitize(color)}`;
};

const generateVariantName = (size: string, color: string) => {
  return `${size} - ${color}`;
};
```

**Add Change Handlers:**
```tsx
const handleSizeChange = (size: string) => {
  setFormData(prev => {
    const newData = { ...prev, size };
    if (newData.size && newData.color) {
      newData.sku = generateSKU(productName, newData.size, newData.color);
      newData.variant_name = generateVariantName(newData.size, newData.color);
    }
    return newData;
  });
};

const handleColorChange = (colorName: string) => {
  const color = COLORS.find(c => c.name === colorName);
  setFormData(prev => {
    const newData = {
      ...prev,
      color: colorName,
      color_hex: color?.hex || '',
    };
    if (newData.size && newData.color) {
      newData.sku = generateSKU(productName, newData.size, newData.color);
      newData.variant_name = generateVariantName(newData.size, newData.color);
    }
    return newData;
  });
};
```

**Update Form UI:**
```tsx
{/* Size Dropdown */}
<div>
  <label>Size *</label>
  <select value={formData.size} onChange={(e) => handleSizeChange(e.target.value)}>
    <option value="">Select Size</option>
    {SIZES.map(size => (
      <option key={size} value={size}>{size}</option>
    ))}
  </select>
</div>

{/* Color Dropdown */}
<div>
  <label>Color *</label>
  <select value={formData.color} onChange={(e) => handleColorChange(e.target.value)}>
    <option value="">Select Color</option>
    {COLORS.map(color => (
      <option key={color.name} value={color.name}>{color.name}</option>
    ))}
  </select>
  {formData.color_hex && (
    <div className="flex items-center gap-2 mt-2">
      <div 
        className="w-6 h-6 rounded border border-white/20" 
        style={{ backgroundColor: formData.color_hex }}
      />
      <span className="text-white/60 text-sm">{formData.color_hex}</span>
    </div>
  )}
</div>

{/* Remove SKU and Variant Name inputs - they're auto-generated */}
```

## Testing Checklist:

- [ ] Size dropdown shows all sizes
- [ ] Color dropdown shows all colors
- [ ] Color hex auto-fills when color selected
- [ ] Color preview shows correct color
- [ ] SKU auto-generates (check in console/network)
- [ ] Variant name auto-generates
- [ ] Can create variant successfully
- [ ] Variant appears in list with correct data
- [ ] No duplicate SKU errors

## Future Enhancements:

1. **Custom Colors** - Allow admin to add custom colors
2. **Custom Sizes** - Allow admin to add custom sizes
3. **Size Charts** - Show size chart reference
4. **Bulk Edit** - Edit multiple variants at once
5. **Import/Export** - Import variants from CSV
6. **Templates** - Save variant templates for reuse

## Notes:

- SKU format: `PRODUCT-NAME-SIZE-COLOR`
- Variant name format: `SIZE - COLOR`
- Color hex is required for proper display
- Shipping dimensions default to clothing standards
- Can override price per variant if needed
