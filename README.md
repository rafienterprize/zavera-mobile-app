# ZAVERA Mobile App - Flutter

Mobile application untuk ZAVERA Fashion Store yang dibangun dengan Flutter & Dart.

## 📱 Features

- **Home Screen** dengan hero carousel dan kategori produk
- **Product Listing** dengan grid layout
- **Product Detail** dengan image carousel, size selection, dan quantity picker
- **Shopping Cart** dengan persistent storage
- **Wishlist** untuk menyimpan produk favorit
- **User Authentication** (Login & Register)
- **Checkout** dengan integrasi Midtrans payment
- **Profile Management**
- **Responsive UI** dengan Material Design 3

## 🛠 Tech Stack

- **Flutter** - UI Framework
- **Dart** - Programming Language
- **Provider** - State Management
- **HTTP/Dio** - API Communication
- **SharedPreferences** - Local Storage
- **CachedNetworkImage** - Image Caching
- **CarouselSlider** - Image Carousel
- **GoogleFonts** - Typography

## 📋 Prerequisites

- Flutter SDK (3.0.0 atau lebih tinggi)
- Dart SDK
- Android Studio / VS Code dengan Flutter extension
- Android Emulator atau iOS Simulator
- Backend API ZAVERA yang sudah running

## 🚀 Installation

### 1. Clone Repository

```bash
git clone https://github.com/ynzphyz/ZAVERA-FASHION-STORE.git
cd ZAVERA-FASHION-STORE/zavera_mobile
```

### 2. Install Dependencies

```bash
flutter pub get
```

### 3. Konfigurasi API

Edit file `lib/services/api_service.dart` dan sesuaikan base URL:

```dart
static const String baseUrl = 'http://YOUR_IP:8080/api';
```

**Catatan:** 
- Untuk Android Emulator gunakan: `http://10.0.2.2:8080/api`
- Untuk iOS Simulator gunakan: `http://localhost:8080/api`
- Untuk Real Device gunakan: `http://YOUR_LOCAL_IP:8080/api`

### 4. Run Application

```bash
# Run on Android
flutter run

# Run on iOS
flutter run

# Run on specific device
flutter devices
flutter run -d <device_id>
```

## 📁 Project Structure

```
zavera_mobile/
├── lib/
│   ├── main.dart                 # Entry point
│   ├── models/                   # Data models
│   │   ├── product.dart
│   │   ├── cart_item.dart
│   │   └── user.dart
│   ├── providers/                # State management
│   │   ├── auth_provider.dart
│   │   ├── cart_provider.dart
│   │   └── wishlist_provider.dart
│   ├── screens/                  # UI Screens
│   │   ├── splash_screen.dart
│   │   ├── home_screen.dart
│   │   ├── categories_screen.dart
│   │   ├── cart_screen.dart
│   │   ├── profile_screen.dart
│   │   ├── product_detail_screen.dart
│   │   ├── login_screen.dart
│   │   ├── register_screen.dart
│   │   └── checkout_screen.dart
│   ├── services/                 # API services
│   │   └── api_service.dart
│   └── widgets/                  # Reusable widgets
│       ├── product_card.dart
│       └── category_card.dart
├── assets/
│   ├── images/
│   └── icons/
├── pubspec.yaml                  # Dependencies
└── README.md
```

## 🎨 UI Components

### Screens

1. **Splash Screen** - Loading screen dengan logo ZAVERA
2. **Home Screen** - Hero carousel, categories, new arrivals, trending products
3. **Categories Screen** - Grid view semua kategori
4. **Product Detail** - Image carousel, size selection, add to cart
5. **Cart Screen** - List cart items dengan quantity control
6. **Profile Screen** - User info dan menu navigasi
7. **Login/Register** - Authentication forms
8. **Checkout** - Order summary dan payment

### Widgets

- **ProductCard** - Card untuk menampilkan produk
- **CategoryCard** - Card untuk kategori dengan gradient overlay
- Custom bottom navigation bar dengan badge counter

## 🔌 API Integration

App ini terhubung dengan backend ZAVERA:

### Endpoints yang digunakan:

- `GET /api/products` - Fetch all products
- `GET /api/products/:id` - Fetch product detail
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `GET /api/auth/me` - Get current user
- `POST /api/checkout` - Create order

## 💾 Local Storage

Menggunakan **SharedPreferences** untuk menyimpan:

- Auth token
- Cart items
- Wishlist product IDs

## 🎯 State Management

Menggunakan **Provider** pattern dengan 3 providers:

1. **AuthProvider** - Manage authentication state
2. **CartProvider** - Manage shopping cart
3. **WishlistProvider** - Manage wishlist

## 🔧 Build for Production

### Android

```bash
flutter build apk --release
# Output: build/app/outputs/flutter-apk/app-release.apk

# Build App Bundle (untuk Google Play)
flutter build appbundle --release
```

### iOS

```bash
flutter build ios --release
```

## 📱 Testing

```bash
# Run tests
flutter test

# Run with coverage
flutter test --coverage
```

## 🐛 Troubleshooting

### Network Error

Jika mengalami error koneksi:
1. Pastikan backend sudah running
2. Cek IP address di `api_service.dart`
3. Untuk Android Emulator, gunakan `10.0.2.2` bukan `localhost`

### Build Error

```bash
# Clean build
flutter clean
flutter pub get
flutter run
```

## 📝 TODO / Future Improvements

- [ ] Add search functionality
- [ ] Add filter & sort products
- [ ] Implement order tracking
- [ ] Add push notifications
- [ ] Implement Midtrans WebView payment
- [ ] Add product reviews & ratings
- [ ] Implement address management
- [ ] Add dark mode support
- [ ] Implement internationalization (i18n)

## 👥 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is for educational and commercial use.

## 📧 Contact

For issues or questions, please open an issue in the repository.

---

Built with ❤️ using Flutter & Dart
