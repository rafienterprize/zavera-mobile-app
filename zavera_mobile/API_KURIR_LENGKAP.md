# 📦 API Kurir & Tracking - LENGKAP!

## ✅ Semua API Kurir Sudah Ada di `api_service.dart`

Total: **16 endpoints** untuk shipping, tracking, dan address management.

---

## 🚚 SHIPPING APIs (13 endpoints)

### 1. Get Shipping Providers
```dart
Future<List<dynamic>> getShippingProviders()
```
**Endpoint:** `GET /shipping/providers`  
**Return:** List kurir yang tersedia (JNE, J&T, SiCepat, dll)

---

### 2. Get Shipping Rates
```dart
Future<List<dynamic>> getShippingRates(Map<String, dynamic> shippingData)
```
**Endpoint:** `POST /shipping/rates`  
**Payload:**
```json
{
  "origin_area_id": "IDNP6IDNC148IDND1649IDZ10677",
  "destination_area_id": "IDNP6IDNC148IDND1649IDZ10677",
  "weight": 1000
}
```
**Return:** List rates dari berbagai kurir dengan harga & estimasi

---

### 3. Search Areas (Biteship)
```dart
Future<List<dynamic>> searchAreas(String query)
```
**Endpoint:** `GET /shipping/areas?q={query}`  
**Example:** `searchAreas("Semarang")` atau `searchAreas("50191")`  
**Return:** List areas dengan `area_id`, `name`, `postal_code`

**Use case:** Autocomplete untuk input alamat

---

### 4. Get Provinces
```dart
Future<List<dynamic>> getProvinces()
```
**Endpoint:** `GET /shipping/provinces`  
**Return:** List provinsi Indonesia

---

### 5. Get Cities
```dart
Future<List<dynamic>> getCities(int provinceId)
```
**Endpoint:** `GET /shipping/cities?province_id={provinceId}`  
**Return:** List kota/kabupaten dalam provinsi

---

### 6. Get Districts
```dart
Future<List<dynamic>> getDistricts(int cityId)
```
**Endpoint:** `GET /shipping/districts?city_id={cityId}`  
**Return:** List kecamatan dalam kota

---

### 7. Get Subdistricts
```dart
Future<List<dynamic>> getSubdistricts(int districtId)
```
**Endpoint:** `GET /shipping/subdistricts?district_id={districtId}`  
**Return:** List kelurahan dalam kecamatan

---

### 8. Get Cart Shipping Preview
```dart
Future<Map<String, dynamic>?> getCartShippingPreview()
```
**Endpoint:** `GET /shipping/preview`  
**Return:** Preview ongkir untuk cart saat ini

---

### 9-13. Address Management (5 endpoints)

#### Get User Addresses
```dart
Future<List<dynamic>> getUserAddresses()
```
**Endpoint:** `GET /user/addresses`

#### Create Address
```dart
Future<Map<String, dynamic>?> createAddress(Map<String, dynamic> addressData)
```
**Endpoint:** `POST /user/addresses`  
**Payload:**
```json
{
  "label": "Rumah",
  "recipient_name": "John Doe",
  "phone": "081234567890",
  "address": "Jl. Sudirman No. 123",
  "province_id": 6,
  "city_id": 148,
  "district_id": 1649,
  "subdistrict_id": 10677,
  "postal_code": "50191",
  "is_default": true
}
```

#### Get Address by ID
```dart
Future<Map<String, dynamic>?> getAddress(int addressId)
```
**Endpoint:** `GET /user/addresses/{addressId}`

#### Update Address
```dart
Future<bool> updateAddress(int addressId, Map<String, dynamic> addressData)
```
**Endpoint:** `PUT /user/addresses/{addressId}`

#### Delete Address
```dart
Future<bool> deleteAddress(int addressId)
```
**Endpoint:** `DELETE /user/addresses/{addressId}`

#### Set Default Address
```dart
Future<bool> setDefaultAddress(int addressId)
```
**Endpoint:** `POST /user/addresses/{addressId}/default`

---

## 📍 TRACKING APIs (3 endpoints)

### 1. Track by Resi Number
```dart
Future<Map<String, dynamic>?> getTrackingByResi(String resi)
```
**Endpoint:** `GET /tracking/{resi}`  
**Example:** `getTrackingByResi("JP1234567890")`  
**Return:**
```json
{
  "resi": "JP1234567890",
  "courier": "jne",
  "status": "delivered",
  "history": [
    {
      "date": "2024-01-15 10:30:00",
      "description": "Paket telah diterima",
      "location": "Jakarta"
    }
  ]
}
```

---

### 2. Get Shipment Details
```dart
Future<Map<String, dynamic>?> getShipment(int shipmentId)
```
**Endpoint:** `GET /shipments/{shipmentId}`  
**Return:** Detail lengkap shipment termasuk tracking history

---

### 3. Refresh Tracking
```dart
Future<bool> refreshTracking(int shipmentId)
```
**Endpoint:** `POST /shipments/{shipmentId}/refresh`  
**Use case:** Force update tracking dari Biteship API

---

## 🎯 FLOW PENGGUNAAN

### Flow 1: Checkout dengan Shipping
```dart
// 1. User pilih/buat alamat
final addresses = await apiService.getUserAddresses();
// atau
final newAddress = await apiService.createAddress(addressData);

// 2. Get shipping rates
final rates = await apiService.getShippingRates({
  'origin_area_id': 'ORIGIN_ID',
  'destination_area_id': address['area_id'],
  'weight': 1000, // gram
});

// 3. User pilih kurir & service
// rates = [
//   { courier: 'jne', service: 'REG', price: 15000, etd: '2-3' },
//   { courier: 'jne', service: 'YES', price: 25000, etd: '1-2' },
// ]

// 4. Checkout
final order = await apiService.checkout({
  'shipping_address_id': addressId,
  'courier_code': 'jne',
  'courier_service': 'REG',
});
```

---

### Flow 2: Input Alamat dengan Autocomplete
```dart
// User ketik "Semarang"
final areas = await apiService.searchAreas("Semarang");

// areas = [
//   {
//     area_id: "IDNP6IDNC148IDND1649IDZ10677",
//     name: "Semarang Tengah, Semarang, Jawa Tengah",
//     postal_code: "50132"
//   },
//   ...
// ]

// User pilih area
final selectedArea = areas[0];

// Save address dengan area_id
await apiService.createAddress({
  'label': 'Rumah',
  'recipient_name': 'John Doe',
  'phone': '081234567890',
  'address': 'Jl. Sudirman No. 123',
  'area_id': selectedArea['area_id'], // IMPORTANT!
  'postal_code': selectedArea['postal_code'],
});
```

---

### Flow 3: Track Order
```dart
// Dari order detail, ambil resi number
final order = await apiService.getOrder(orderCode);
final resi = order['shipment']['resi'];

// Track by resi
final tracking = await apiService.getTrackingByResi(resi);

// tracking = {
//   resi: "JP1234567890",
//   courier: "jne",
//   status: "on_delivery",
//   history: [
//     { date: "...", description: "...", location: "..." },
//     ...
//   ]
// }

// Refresh tracking (force update)
await apiService.refreshTracking(order['shipment']['id']);
```

---

### Flow 4: Manage Addresses
```dart
// List addresses
final addresses = await apiService.getUserAddresses();

// Create new
final newAddress = await apiService.createAddress({...});

// Update
await apiService.updateAddress(addressId, {...});

// Set as default
await apiService.setDefaultAddress(addressId);

// Delete
await apiService.deleteAddress(addressId);
```

---

## 🚨 IMPORTANT NOTES

### 1. Area ID vs Province/City/District
Ada 2 cara input alamat:

**Cara 1: Biteship Area Search (RECOMMENDED)**
```dart
// User search area
final areas = await searchAreas("Semarang");
// Save dengan area_id
address['area_id'] = selectedArea['area_id'];
```

**Cara 2: Manual Province/City/District**
```dart
// User pilih province → city → district → subdistrict
final provinces = await getProvinces();
final cities = await getCities(provinceId);
final districts = await getDistricts(cityId);
final subdistricts = await getSubdistricts(districtId);

// Save dengan IDs
address['province_id'] = provinceId;
address['city_id'] = cityId;
address['district_id'] = districtId;
address['subdistrict_id'] = subdistrictId;
```

**Website pakai Cara 1 (Biteship)** - lebih cepat & akurat!

---

### 2. Weight Calculation
```dart
// Total weight dari cart items
int totalWeight = 0;
for (var item in cart) {
  totalWeight += item.weight * item.quantity;
}

// Get rates dengan total weight
final rates = await getShippingRates({
  'origin_area_id': 'ORIGIN',
  'destination_area_id': 'DESTINATION',
  'weight': totalWeight, // dalam gram
});
```

---

### 3. Courier Codes
```
jne     - JNE
jnt     - J&T Express
sicepat - SiCepat
tiki    - TIKI
pos     - POS Indonesia
ninja   - Ninja Xpress
anteraja - AnterAja
lion    - Lion Parcel
```

---

### 4. Tracking Status
```
pending       - Menunggu pickup
picked_up     - Sudah dipickup
on_delivery   - Dalam pengiriman
delivered     - Sudah diterima
cancelled     - Dibatalkan
returned      - Dikembalikan
```

---

## 📱 UI COMPONENTS YANG PERLU DIBUAT

### 1. Address Form Screen
- Input fields (nama, phone, alamat)
- **Autocomplete search** untuk area (pakai `searchAreas`)
- Checkbox "Set as default"
- Save button

### 2. Address List Screen
- List addresses dengan label (Rumah, Kantor, dll)
- Badge "Default" untuk default address
- Edit & Delete buttons
- Add new address button

### 3. Shipping Options Screen (Checkout)
- List kurir dengan logo
- Service name & price
- Estimasi pengiriman (ETD)
- Radio button untuk pilih

### 4. Tracking Screen
- Resi number (copyable)
- Current status dengan icon
- Timeline tracking history
- Refresh button

---

## ✅ CHECKLIST IMPLEMENTASI

### Shipping
- [ ] Area search dengan autocomplete
- [ ] Get shipping rates
- [ ] Display shipping options
- [ ] Calculate total weight dari cart

### Address Management
- [ ] List addresses
- [ ] Create address form
- [ ] Edit address
- [ ] Delete address
- [ ] Set default address

### Tracking
- [ ] Track by resi number
- [ ] Display tracking history
- [ ] Refresh tracking
- [ ] Show status dengan icon/color

---

## 🎊 CONCLUSION

**SEMUA API KURIR SUDAH ADA!** Tinggal buat UI dan connect ke API yang sudah ready.

**Total APIs:**
- ✅ 13 Shipping endpoints
- ✅ 3 Tracking endpoints
- ✅ 5 Address management endpoints

**Integration dengan Biteship:**
- ✅ Real shipping rates
- ✅ Multiple couriers (JNE, J&T, SiCepat, dll)
- ✅ Real-time tracking
- ✅ Area search untuk input alamat

**Siap production! 🚀**
