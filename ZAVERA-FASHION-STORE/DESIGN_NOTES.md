# 🎨 ZAVERA - Redesigned Premium Fashion E-Commerce

## Design Updates

### Visual Design

- **Premium Aesthetic**: Clean, minimal, high-end look inspired by Zara, COS, and SSENSE
- **Typography**: Playfair Display serif for headings, Inter for body text
- **Color Palette**:
  - Primary: `#0a0a0a` (Deep black)
  - Secondary: `#fafafa` (Off-white)
  - Accent: `#e5e5e5` (Light gray)
  - Muted: `#737373` (Medium gray)

### New Components

#### 1. **Hero Section** (`Hero.tsx`)

- Full viewport height
- Full-width background image with dark overlay
- Bold, large typography
- Two CTA buttons (primary + secondary)
- Animated scroll indicator
- Smooth fade-in animations

#### 2. **Brand Values** (`BrandValues.tsx`)

- Three core values with icons
- Centered layout with hover effects
- Stagger animations on scroll

#### 3. **Product Card** (`ProductCard.tsx`)

- 3:4 aspect ratio
- Image zoom on hover
- Quick view button fade-in
- Low stock / sold out badges
- Skeleton loading states

### Enhanced Components

#### Header

- Fixed/sticky with transparency on top
- Backdrop blur when scrolled
- Animated cart badge
- Minimal navigation
- Responsive design

#### Footer

- Extended grid layout (4 columns)
- Social media icons
- Improved typography
- Better spacing

### Animations (Framer Motion)

- Hero fade-in and slide-up
- Product cards stagger on scroll
- Hover effects (scale, zoom)
- Cart badge spring animation
- Scroll-triggered animations with `whileInView`

### Responsive Design

- Mobile-first approach
- Breakpoints: sm, md, lg, xl
- Touch-friendly buttons
- Optimized spacing

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Styling**: Tailwind CSS
- **Animations**: Framer Motion
- **Fonts**: Google Fonts (Inter + Playfair Display)

## Running the Project

```bash
# Install dependencies
npm install

# Run development server
npm run dev
```

Visit: `http://localhost:3000`

## Features

✅ Full-screen hero with CTA
✅ Brand philosophy section
✅ Product grid with hover effects
✅ Skeleton loading states
✅ Newsletter subscription
✅ Sticky header with scroll effects
✅ Smooth animations throughout
✅ Premium typography scale
✅ Fully responsive

## Notes

- All existing functionality preserved (cart, checkout, etc.)
- Backend integration maintained
- Production-ready code structure
- No inline styles (Tailwind only)
