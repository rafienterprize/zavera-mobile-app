# 🚀 Deploy ZAVERA Frontend ke Cloudflare Pages

## 📋 Prerequisites

1. ✅ Akun Cloudflare (gratis)
2. ✅ GitHub repository sudah push
3. ✅ Frontend Next.js ready

---

## 🎯 Metode 1: Deploy via Cloudflare Dashboard (RECOMMENDED)

### Step 1: Login ke Cloudflare

1. Buka: https://dash.cloudflare.com
2. Login dengan akun Anda
3. Pilih "Workers & Pages" di sidebar

### Step 2: Create New Project

1. Klik **"Create application"**
2. Pilih tab **"Pages"**
3. Klik **"Connect to Git"**

### Step 3: Connect GitHub Repository

1. Klik **"Connect GitHub"**
2. Authorize Cloudflare untuk access GitHub
3. Pilih repository: **ZAVERA-FASHION-STORE**
4. Klik **"Begin setup"**

### Step 4: Configure Build Settings

**Project name:** `zavera-fashion-store` (atau nama lain)

**Production branch:** `main`

**Framework preset:** `Next.js`

**Build command:**
```bash
cd frontend && npm install && npm run build
```

**Build output directory:**
```
frontend/.next
```

**Root directory (advanced):**
```
frontend
```

### Step 5: Environment Variables

Klik **"Add variable"** dan tambahkan:

```
NEXT_PUBLIC_API_URL=https://your-backend-url.com/api
NEXT_PUBLIC_GOOGLE_CLIENT_ID=your-google-client-id
NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=your-midtrans-client-key
```

**PENTING:** Ganti dengan nilai yang sesuai!

### Step 6: Deploy

1. Klik **"Save and Deploy"**
2. Tunggu proses build (5-10 menit)
3. ✅ Selesai! Frontend akan live di: `https://zavera-fashion-store.pages.dev`

---

## 🎯 Metode 2: Deploy via Wrangler CLI

### Step 1: Install Wrangler

```bash
npm install -g wrangler
```

### Step 2: Login ke Cloudflare

```bash
wrangler login
```

Browser akan terbuka, authorize Wrangler.

### Step 3: Build Frontend

```bash
cd frontend
npm run build
```

### Step 4: Deploy

```bash
wrangler pages deploy .next --project-name=zavera-fashion-store
```

---

## ⚙️ Konfigurasi Next.js untuk Cloudflare

### Option A: Static Export (Recommended untuk Cloudflare)

Edit `frontend/next.config.js`:

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export', // Enable static export
  images: {
    unoptimized: true, // Required for static export
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
      },
      {
        protocol: 'https',
        hostname: 'res.cloudinary.com',
      },
    ],
  },
  // Disable features not supported in static export
  trailingSlash: true,
};

module.exports = nextConfig;
```

**Build command untuk static export:**
```bash
cd frontend && npm install && npm run build
```

**Output directory:**
```
frontend/out
```

### Option B: Next.js with Edge Runtime

Jika butuh SSR, upgrade Next.js dulu:

```bash
cd frontend
npm install next@latest react@latest react-dom@latest
npm install --save-dev @cloudflare/next-on-pages
```

Edit `package.json`:
```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "pages:build": "npx @cloudflare/next-on-pages",
    "preview": "npm run pages:build && wrangler pages dev",
    "deploy": "npm run pages:build && wrangler pages deploy"
  }
}
```

---

## 🔧 Troubleshooting

### Issue 1: Build Failed - "Module not found"

**Solution:**
```bash
cd frontend
npm install
npm run build
```

Pastikan semua dependencies terinstall.

### Issue 2: Environment Variables Not Working

**Solution:**
- Tambahkan prefix `NEXT_PUBLIC_` untuk client-side variables
- Restart build setelah add environment variables
- Check di Cloudflare Dashboard → Pages → Settings → Environment variables

### Issue 3: API Calls Failed (CORS)

**Solution:**

Backend perlu allow Cloudflare domain:

```go
// backend/main.go
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",
        "https://zavera-fashion-store.pages.dev", // Add this!
        "https://your-custom-domain.com",
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

### Issue 4: Images Not Loading

**Solution:**

Untuk static export, set `unoptimized: true`:

```javascript
// next.config.js
images: {
  unoptimized: true,
}
```

### Issue 5: Dynamic Routes Not Working

**Solution:**

Static export tidak support dynamic routes dengan `getServerSideProps`. Gunakan:
- `getStaticProps` + `getStaticPaths` untuk dynamic routes
- Atau upgrade ke Next.js 14.3+ dan gunakan `@cloudflare/next-on-pages`

---

## 🌐 Custom Domain

### Step 1: Add Custom Domain

1. Cloudflare Dashboard → Pages → Your Project
2. Tab **"Custom domains"**
3. Klik **"Set up a custom domain"**
4. Enter domain: `zavera.com` atau `www.zavera.com`

### Step 2: Update DNS

Cloudflare akan auto-configure DNS jika domain sudah di Cloudflare.

Jika domain di registrar lain:
1. Add CNAME record:
   ```
   Name: www
   Value: zavera-fashion-store.pages.dev
   ```

### Step 3: SSL/TLS

Cloudflare auto-provision SSL certificate (gratis).

---

## 📊 Monitoring & Analytics

### Enable Web Analytics

1. Cloudflare Dashboard → Pages → Your Project
2. Tab **"Analytics"**
3. Enable **"Web Analytics"**

### View Deployment Logs

1. Cloudflare Dashboard → Pages → Your Project
2. Tab **"Deployments"**
3. Click deployment → View logs

---

## 🔄 Auto-Deploy on Git Push

Cloudflare Pages auto-deploy ketika:
- ✅ Push ke branch `main` → Deploy to production
- ✅ Push ke branch lain → Deploy to preview URL
- ✅ Pull request → Deploy to preview URL

**Preview URL format:**
```
https://[commit-hash].zavera-fashion-store.pages.dev
```

---

## 💰 Pricing

**Cloudflare Pages Free Tier:**
- ✅ Unlimited requests
- ✅ Unlimited bandwidth
- ✅ 500 builds per month
- ✅ 1 build at a time
- ✅ Free SSL certificate
- ✅ Free DDoS protection

**Paid Plan ($20/month):**
- ✅ 5,000 builds per month
- ✅ 5 concurrent builds
- ✅ Advanced analytics

---

## 🎯 Recommended Setup

### For Production:

1. **Use Static Export** (fastest, most reliable)
   ```javascript
   // next.config.js
   output: 'export'
   ```

2. **Environment Variables:**
   ```
   NEXT_PUBLIC_API_URL=https://api.zavera.com
   NEXT_PUBLIC_GOOGLE_CLIENT_ID=...
   NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=...
   ```

3. **Custom Domain:**
   ```
   www.zavera.com → Cloudflare Pages
   ```

4. **Backend CORS:**
   ```go
   AllowOrigins: []string{
       "https://www.zavera.com",
       "https://zavera-fashion-store.pages.dev",
   }
   ```

---

## 📝 Quick Deploy Checklist

- [ ] Push latest code to GitHub
- [ ] Login ke Cloudflare Dashboard
- [ ] Create new Pages project
- [ ] Connect GitHub repository
- [ ] Configure build settings:
  - [ ] Build command: `cd frontend && npm install && npm run build`
  - [ ] Output directory: `frontend/out` (for static export)
  - [ ] Root directory: `frontend`
- [ ] Add environment variables
- [ ] Click "Save and Deploy"
- [ ] Wait for build to complete
- [ ] Test deployed site
- [ ] (Optional) Add custom domain
- [ ] Update backend CORS settings

---

## 🚀 Deploy Now!

**Cara Tercepat:**

1. Buka: https://dash.cloudflare.com
2. Workers & Pages → Create application → Pages
3. Connect GitHub → Select ZAVERA-FASHION-STORE
4. Configure:
   - Framework: Next.js
   - Build command: `cd frontend && npm install && npm run build`
   - Output: `frontend/.next`
   - Root: `frontend`
5. Add environment variables
6. Deploy!

**Estimated time:** 10-15 minutes

---

**Good luck with deployment! 🎉**
