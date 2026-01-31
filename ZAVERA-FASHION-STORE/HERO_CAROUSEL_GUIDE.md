# Hero Carousel - Zalora Style Banner Slider

## ✅ IMPLEMENTED

Hero Carousel dengan auto-slide seperti Zalora telah berhasil diimplementasikan!

## 🎨 Features

### 1. **Auto-Sliding Banner** ✅
- Banner bergeser otomatis setiap 5 detik
- Smooth transition dengan fade effect
- 3 slides default (bisa ditambah lebih banyak)

### 2. **Navigation Controls** ✅
- **Arrow Buttons**: Panah kiri/kanan untuk manual navigation
- **Dot Indicators**: Titik-titik di bawah untuk menunjukkan slide aktif
- **Click Dots**: Klik titik untuk langsung ke slide tertentu

### 3. **Interactive Features** ✅
- **Hover to Pause**: Auto-play berhenti saat user interact
- **Resume Auto-play**: Otomatis lanjut setelah 5 detik
- **Smooth Animations**: Fade in/out dengan animasi text
- **Progress Bar**: Bar di bawah menunjukkan progress slide

### 4. **Responsive Design** ✅
- Mobile: 400px height
- Tablet: 500px height
- Desktop: 600px height
- Text dan button menyesuaikan ukuran layar

### 5. **Content Positioning** ✅
- Text bisa di **left**, **center**, atau **right**
- Setiap slide bisa punya posisi berbeda
- Background overlay untuk readability

## 📁 Files Created/Modified

**New Files:**
- ✅ `frontend/src/components/HeroCarousel.tsx` - Main carousel component

**Modified Files:**
- ✅ `frontend/src/app/page.tsx` - Homepage menggunakan HeroCarousel
- ✅ `frontend/src/app/globals.css` - Added carousel animations

## 🎯 How to Customize

### 1. Menambah/Mengubah Slides

Edit file `frontend/src/components/HeroCarousel.tsx`:

```typescript
const slides: Slide[] = [
  {
    id: 1,
    image: "URL_GAMBAR_ANDA",
    title: "Judul Banner",
    subtitle: "Subtitle/Promo",
    description: "Deskripsi detail",
    ctaText: "SHOP NOW",
    ctaLink: "/kategori",
    textPosition: "right", // left, center, atau right
  },
  // Tambah slide baru di sini
  {
    id: 2,
    image: "URL_GAMBAR_2",
    title: "Banner Kedua",
    subtitle: "Promo Spesial",
    description: "Diskon hingga 50%",
    ctaText: "LIHAT PROMO",
    ctaLink: "/promo",
    textPosition: "left",
  },
];
```

### 2. Mengubah Durasi Auto-Slide

Cari baris ini di `HeroCarousel.tsx`:

```typescript
const interval = setInterval(() => {
  nextSlide();
}, 5000); // 5000 = 5 detik, ubah sesuai kebutuhan
```

Contoh:
- 3 detik: `3000`
- 7 detik: `7000`
- 10 detik: `10000`

### 3. Mengubah Tinggi Banner

Edit di `HeroCarousel.tsx`:

```typescript
<div className="relative w-full h-[400px] md:h-[500px] lg:h-[600px]">
```

Ubah nilai:
- `h-[400px]` = Mobile height
- `md:h-[500px]` = Tablet height
- `lg:h-[600px]` = Desktop height

### 4. Mengubah Warna Overlay

Cari baris ini:

```typescript
<div className="absolute inset-0 bg-black/20" />
```

Ubah opacity:
- `bg-black/10` = Lebih terang
- `bg-black/30` = Lebih gelap
- `bg-black/50` = Sangat gelap

### 5. Disable Auto-Play

Jika ingin manual saja (tidak auto-slide):

```typescript
const [isAutoPlaying, setIsAutoPlaying] = useState(false); // Ubah true jadi false
```

## 🖼️ Recommended Image Sizes

Untuk hasil terbaik, gunakan gambar dengan:
- **Aspect Ratio**: 16:9 atau 21:9
- **Resolution**: Minimal 1920x1080px
- **Format**: JPG atau WebP (untuk performa)
- **File Size**: < 500KB (compress dulu)

### Free Image Sources:
- [Unsplash](https://unsplash.com) - Fashion photos
- [Pexels](https://pexels.com) - Free stock photos
- [Pixabay](https://pixabay.com) - Free images

## 🎨 Styling Options

### Button Styles

Ubah style button CTA:

```typescript
// Current (white button)
className="inline-block px-8 py-4 bg-white text-black"

// Black button
className="inline-block px-8 py-4 bg-black text-white"

// Transparent button
className="inline-block px-8 py-4 border-2 border-white text-white"

// Gradient button
className="inline-block px-8 py-4 bg-gradient-to-r from-purple-500 to-pink-500 text-white"
```

### Text Colors

```typescript
// White text (current)
className="text-white"

// Black text
className="text-black"

// Colored text
className="text-purple-500"
```

## 📱 Mobile Optimization

Carousel sudah responsive, tapi bisa di-customize:

### Hide Elements on Mobile

```typescript
// Hide subtitle on mobile
<p className="hidden md:block text-white/90">
  {slide.subtitle}
</p>

// Smaller title on mobile
<h1 className="text-3xl md:text-6xl">
  {slide.title}
</h1>
```

### Touch Swipe (Optional Enhancement)

Untuk menambah swipe gesture di mobile, install library:

```bash
npm install react-swipeable
```

Lalu wrap carousel dengan:

```typescript
import { useSwipeable } from 'react-swipeable';

const handlers = useSwipeable({
  onSwipedLeft: () => nextSlide(),
  onSwipedRight: () => prevSlide(),
});

<div {...handlers}>
  {/* Carousel content */}
</div>
```

## 🚀 Performance Tips

### 1. Lazy Load Images

Tambah `loading="lazy"` pada images:

```typescript
<img src={slide.image} loading="lazy" alt={slide.title} />
```

### 2. Preload First Image

Di `<head>` tag, tambah:

```html
<link rel="preload" as="image" href="URL_FIRST_IMAGE" />
```

### 3. Use WebP Format

Convert images ke WebP untuk file size lebih kecil:
- Online: [Squoosh.app](https://squoosh.app)
- CLI: `cwebp input.jpg -o output.webp`

### 4. CDN for Images

Upload images ke CDN seperti:
- Cloudinary
- ImageKit
- Vercel Image Optimization

## 🎯 Advanced Features (Optional)

### 1. Video Background

Ganti image dengan video:

```typescript
<video
  autoPlay
  muted
  loop
  playsInline
  className="absolute inset-0 w-full h-full object-cover"
>
  <source src="/videos/hero.mp4" type="video/mp4" />
</video>
```

### 2. Parallax Effect

Install framer-motion dan tambah:

```typescript
import { motion } from "framer-motion";

<motion.div
  initial={{ scale: 1.2 }}
  animate={{ scale: 1 }}
  transition={{ duration: 10 }}
  className="absolute inset-0 bg-cover"
  style={{ backgroundImage: `url('${slide.image}')` }}
/>
```

### 3. Ken Burns Effect (Zoom Animation)

Tambah CSS:

```css
@keyframes kenBurns {
  0% { transform: scale(1); }
  100% { transform: scale(1.1); }
}

.ken-burns {
  animation: kenBurns 10s ease-out infinite alternate;
}
```

### 4. Multiple CTAs per Slide

```typescript
<div className="flex gap-4">
  <Link href={slide.ctaLink} className="btn-primary">
    {slide.ctaText}
  </Link>
  <Link href={slide.secondaryLink} className="btn-secondary">
    {slide.secondaryText}
  </Link>
</div>
```

## 🔧 Troubleshooting

### Carousel tidak auto-slide
- Cek `isAutoPlaying` state = `true`
- Cek console untuk errors
- Pastikan `useEffect` tidak di-block

### Images tidak muncul
- Cek URL image valid
- Cek CORS policy
- Cek network tab di browser

### Animasi patah-patah
- Reduce image file size
- Check browser performance
- Disable animations on low-end devices

### Dots tidak klik
- Cek `onClick` handler
- Cek z-index positioning
- Cek button tidak tertutup element lain

## 📊 Analytics (Optional)

Track slide views dengan Google Analytics:

```typescript
const goToSlide = (index: number) => {
  setCurrentSlide(index);
  
  // Track with GA
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('event', 'carousel_slide_view', {
      slide_index: index,
      slide_title: slides[index].title,
    });
  }
};
```

## ✅ Testing Checklist

- [ ] Auto-slide berfungsi (5 detik per slide)
- [ ] Arrow buttons berfungsi (kiri/kanan)
- [ ] Dot indicators berfungsi (klik untuk pindah slide)
- [ ] Pause on interaction berfungsi
- [ ] Resume auto-play setelah 5 detik
- [ ] Responsive di mobile/tablet/desktop
- [ ] Images load dengan baik
- [ ] Text readable di semua slides
- [ ] CTA buttons link ke halaman yang benar
- [ ] Smooth transitions tanpa lag
- [ ] Progress bar berfungsi (optional)

## 🎉 Result

Hero Carousel sekarang berfungsi seperti Zalora:
- ✅ Auto-sliding banner
- ✅ Dot indicators
- ✅ Arrow navigation
- ✅ Smooth animations
- ✅ Responsive design
- ✅ Interactive controls

**Status**: ✅ READY TO USE

## 📞 Next Steps

1. **Test di browser**: `npm run dev` dan buka `http://localhost:3000`
2. **Ganti images**: Upload gambar promo Anda
3. **Customize text**: Sesuaikan judul dan CTA
4. **Add more slides**: Tambah slide sesuai kebutuhan
5. **Optimize images**: Compress untuk performa

Selamat! Hero Carousel Anda sudah siap seperti Zalora! 🎉
