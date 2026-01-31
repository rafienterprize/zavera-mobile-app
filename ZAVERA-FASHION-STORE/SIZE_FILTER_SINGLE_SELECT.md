# Update: Size Filter - Single Select

## 🔄 Perubahan

Filter ukuran diubah dari **multi-select** (bisa pilih banyak) menjadi **single-select** (hanya bisa pilih satu).

## ❌ Behavior Sebelumnya (Multi-Select)

```
User bisa klik: XS, XL, XXL sekaligus
Produk yang muncul: yang punya XS ATAU XL ATAU XXL
```

**Masalah:**
- User bingung kenapa bisa pilih banyak
- Tidak sesuai dengan UX pemilihan size (biasanya pilih satu)

## ✅ Behavior Sekarang (Single-Select)

```
User klik: L
  → Hanya L yang selected
  → Produk yang muncul: hanya yang punya size L

User klik: M (saat L sudah selected)
  → L otomatis unselect
  → M yang selected
  → Produk yang muncul: hanya yang punya size M

User klik: M lagi (saat M sudah selected)
  → M unselect
  → Tidak ada size selected
  → Semua produk muncul kembali
```

## 🔧 Technical Changes

### 1. ProductFilters Interface

**Before:**
```typescript
export interface ProductFilters {
  sizes: string[];  // Array - multi-select
  priceRange: { min: number; max: number } | null;
  subcategory: string | null;
}
```

**After:**
```typescript
export interface ProductFilters {
  size: string | null;  // Single value - single-select
  priceRange: { min: number; max: number } | null;
  subcategory: string | null;
}
```

### 2. Handler Logic

**Before (Multi-Select):**
```typescript
const handleSizeToggle = (size: string) => {
  const newSizes = activeFilters.sizes.includes(size)
    ? activeFilters.sizes.filter((s) => s !== size)  // Remove if exists
    : [...activeFilters.sizes, size];                // Add if not exists
  onFilterChange({ ...activeFilters, sizes: newSizes });
};
```

**After (Single-Select):**
```typescript
const handleSizeChange = (size: string | null) => {
  // If clicking the same size, deselect it (set to null)
  // Otherwise, select the new size
  const newSize = activeFilters.size === size ? null : size;
  onFilterChange({ ...activeFilters, size: newSize });
};
```

### 3. Filter Logic

**Before:**
```typescript
if (filters.sizes.length > 0) {
  result = result.filter((p) => {
    return filters.sizes.some((selectedSize) =>
      p.available_sizes!.includes(selectedSize)
    );
  });
}
```

**After:**
```typescript
if (filters.size) {
  result = result.filter((p) => {
    return p.available_sizes?.includes(filters.size!);
  });
}
```

### 4. UI Rendering

**Before:**
```typescript
className={`... ${
  activeFilters.sizes.includes(size)  // Check if in array
    ? "border-primary bg-primary text-white"
    : "border-gray-200 text-gray-600"
}`}
```

**After:**
```typescript
className={`... ${
  activeFilters.size === size  // Check if equals
    ? "border-primary bg-primary text-white"
    : "border-gray-200 text-gray-600"
}`}
```

### 5. Active Filter Tags

**Before:**
```typescript
{filters.sizes.map((size) => (
  <button onClick={() => onRemoveFilter("size", size)}>
    Ukuran: {size}
  </button>
))}
```

**After:**
```typescript
{filters.size && (
  <button onClick={() => onRemoveFilter("size")}>
    Ukuran: {filters.size}
  </button>
)}
```

## 📁 Files Modified

1. ✅ `frontend/src/components/FilterPanel.tsx`
   - Update `ProductFilters` interface
   - Change `handleSizeToggle` to `handleSizeChange`
   - Update UI rendering logic
   - Update `ActiveFilters` component

2. ✅ `frontend/src/components/CategoryPage.tsx`
   - Update initial state: `sizes: []` → `size: null`
   - Update filter logic for single size
   - Update `handleRemoveFilter`
   - Update `handleClearAllFilters`
   - Update `hasActiveFilters` check
   - Update filter count display

## 🧪 Testing

### Test Scenario 1: Select Single Size
1. Buka `/pria`
2. Klik size "L"
3. ✅ Hanya "L" yang selected (hitam)
4. ✅ Produk yang muncul: Hip Hop Baggy Jeans, Hip Hop Baggy Jeans 22, Jacket Parasut 22

### Test Scenario 2: Change Size
1. Size "L" sudah selected
2. Klik size "M"
3. ✅ "L" otomatis unselect
4. ✅ "M" yang selected
5. ✅ Produk berubah: hanya yang punya size M

### Test Scenario 3: Deselect Size
1. Size "M" sudah selected
2. Klik "M" lagi
3. ✅ "M" unselect
4. ✅ Semua produk muncul kembali

### Test Scenario 4: Active Filter Tag
1. Klik size "L"
2. ✅ Tag "Ukuran: L" muncul di atas
3. Klik X pada tag
4. ✅ Filter removed, semua produk muncul

### Test Scenario 5: Clear All Filters
1. Pilih size "L" dan kategori "Celana"
2. Klik "Hapus Semua"
3. ✅ Semua filter cleared
4. ✅ Semua produk muncul

## 📊 Comparison

| Feature | Multi-Select (Before) | Single-Select (After) |
|---------|----------------------|----------------------|
| **Pilih Size** | Bisa pilih banyak (XS, L, XL) | Hanya satu (L) |
| **Klik Size Lain** | Tambah ke selection | Replace selection |
| **Klik Size yang Sama** | Remove dari selection | Deselect (clear) |
| **Produk yang Muncul** | Punya salah satu size | Punya size yang dipilih |
| **Use Case** | Lihat semua produk dalam beberapa size | Cari produk dalam size spesifik |

## ✅ Summary

**Before:**
- Multi-select: bisa pilih XS, XL, XXL sekaligus ❌
- User bingung kenapa bisa pilih banyak ❌

**After:**
- Single-select: hanya bisa pilih satu size ✅
- Klik size lain → otomatis ganti ✅
- Klik size yang sama → deselect ✅
- UX lebih jelas dan intuitif ✅

**Build Status:** ✅ Success

**Ready to Test!** 🚀
