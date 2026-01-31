# 🚀 ZAVERA Frontend - Ready for Cloudflare Pages Deployment

## ✅ Pre-Deployment Checklist

- [x] Biteship auto-resi implementation complete
- [x] All code pushed to GitHub
- [x] Next.js build tested successfully ✅
- [x] Environment variables documented
- [x] Backend CORS needs update (see below)

---

## 📋 Quick Deploy Steps

### Step 1: Login to Cloudflare Dashboard

1. Go to: https://dash.cloudflare.com
2. Login with your account
3. Click **"Workers & Pages"** in the left sidebar

### Step 2: Create New Pages Project

1. Click **"Create application"**
2. Select **"Pages"** tab
3. Click **"Connect to Git"**
4. Authorize Cloudflare to access your GitHub
5. Select your repository: **ZAVERA-FASHION-STORE**
6. Click **"Begin setup"**

### Step 3: Configure Build Settings

**Project name:** `zavera-fashion-store` (or your preferred name)

**Production branch:** `main`

**Framework preset:** Select **"Next.js"**

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

**Node version:** `18` or `20` (recommended)

### Step 4: Add Environment Variables

Click **"Add variable"** and add these **EXACTLY**:

```
NEXT_PUBLIC_API_URL=https://your-backend-domain.com/api
NEXT_PUBLIC_GOOGLE_CLIENT_ID=822822832882-lp2qrgqm8v3rebts11p2n9uaq7qjogj2.apps.googleusercontent.com
NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=Mid-client-Ytj4WRtkbsTrLe2y
NEXT_PUBLIC_ADMIN_EMAIL=pemberani073@gmail.com
```

**⚠️ IMPORTANT:** Replace `https://your-backend-domain.com/api` with your actual backend URL!

### Step 5: Deploy!

1. Click **"Save and Deploy"**
2. Wait 5-10 minutes for build to complete
3. Your site will be live at: `https://zavera-fashion-store.pages.dev`

---

## 🔧 Backend CORS Configuration

After deployment, you MUST update your backend CORS to allow the Cloudflare domain.

Edit `backend/main.go` and update the CORS configuration:

```go
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",
        "https://zavera-fashion-store.pages.dev",  // Add this!
        // If you add custom domain later:
        // "https://www.zavera.com",
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

**After updating CORS:**
1. Rebuild backend: `go build -o zavera.exe`
2. Restart backend server
3. Test API calls from Cloudflare domain

---

## 🧪 Testing After Deployment

### 1. Test Homepage
- Visit: `https://zavera-fashion-store.pages.dev`
- Check if products load
- Check if images display

### 2. Test Authentication
- Try Google OAuth login
- Check if admin login works

### 3. Test Shopping Flow
- Add product to cart
- Go to checkout
- Test payment (use Midtrans sandbox)

### 4. Test Admin Panel
- Login as admin: `pemberani073@gmail.com`
- Check orders list
- Test "Generate dari Biteship" button
- Test order shipping flow

---

## 🌐 Custom Domain (Optional)

If you want to use your own domain (e.g., `www.zavera.com`):

### Step 1: Add Custom Domain in Cloudflare

1. Go to your Pages project
2. Click **"Custom domains"** tab
3. Click **"Set up a custom domain"**
4. Enter your domain: `www.zavera.com`
5. Cloudflare will auto-configure DNS if domain is on Cloudflare

### Step 2: Update Backend CORS

Add your custom domain to CORS:

```go
AllowOrigins: []string{
    "http://localhost:3000",
    "https://zavera-fashion-store.pages.dev",
    "https://www.zavera.com",  // Add custom domain
},
```

### Step 3: Update Environment Variables

In Cloudflare Pages settings, you might want to update:
```
NEXT_PUBLIC_API_URL=https://api.zavera.com/api
```

---

## 🔄 Auto-Deploy on Git Push

Cloudflare Pages automatically deploys when you push to GitHub:

- ✅ Push to `main` branch → Deploy to production
- ✅ Push to other branches → Deploy to preview URL
- ✅ Pull requests → Deploy to preview URL

**Preview URL format:**
```
https://[commit-hash].zavera-fashion-store.pages.dev
```

---

## 📊 What's Included in Free Tier

Cloudflare Pages Free Tier includes:

- ✅ **Unlimited requests** - No bandwidth limits!
- ✅ **Unlimited bandwidth** - Serve as much traffic as you want
- ✅ **500 builds per month** - More than enough for most projects
- ✅ **Free SSL certificate** - HTTPS automatically enabled
- ✅ **Free DDoS protection** - Enterprise-grade security
- ✅ **Global CDN** - Fast loading worldwide
- ✅ **Automatic deployments** - Deploy on every git push

---

## ⚠️ Important Notes

### 1. Static Export Limitations

Since we're using static export (`output: 'export'`), these Next.js features are NOT available:

- ❌ Server-Side Rendering (SSR)
- ❌ API Routes (`/api/*`)
- ❌ `getServerSideProps`
- ❌ Incremental Static Regeneration (ISR)

**But these ARE available:**
- ✅ Static Site Generation (SSG)
- ✅ Client-Side Rendering (CSR)
- ✅ `getStaticProps` + `getStaticPaths`
- ✅ All client-side features (React hooks, state, etc.)

**Your app uses client-side rendering, so this is PERFECT!** ✅

### 2. Environment Variables

- All environment variables MUST start with `NEXT_PUBLIC_` to be accessible in the browser
- Environment variables are embedded at BUILD time, not runtime
- If you change environment variables, you must rebuild and redeploy

### 3. API Calls

- All API calls go to your backend server (not Cloudflare Pages)
- Make sure backend is accessible from the internet
- Backend must have CORS configured for Cloudflare domain

---

## 🐛 Troubleshooting

### Issue: Build Failed

**Solution:**
```bash
# Test build locally first
cd frontend
npm install
npm run build

# Check for errors
# Fix any errors before deploying
```

### Issue: API Calls Fail (CORS Error)

**Solution:**
1. Check browser console for CORS error
2. Verify backend CORS includes Cloudflare domain
3. Restart backend after CORS update
4. Clear browser cache and try again

### Issue: Images Not Loading

**Solution:**
- Static export requires `unoptimized: true` (already configured ✅)
- Check if image URLs are correct
- Verify Cloudinary/external image hosts allow hotlinking

### Issue: Google OAuth Not Working

**Solution:**
1. Go to Google Cloud Console
2. Add Cloudflare domain to authorized origins:
   - `https://zavera-fashion-store.pages.dev`
3. Add to authorized redirect URIs:
   - `https://zavera-fashion-store.pages.dev/login`

### Issue: Midtrans Payment Not Working

**Solution:**
1. Check Midtrans dashboard settings
2. Verify client key is correct in environment variables
3. Add Cloudflare domain to Midtrans allowed origins (if required)

---

## 📝 Deployment Checklist

Before deploying, make sure:

- [ ] Latest code pushed to GitHub
- [ ] `frontend/next.config.js` configured for static export ✅
- [ ] Environment variables documented ✅
- [ ] Backend CORS will be updated after deployment
- [ ] Google OAuth redirect URIs will be updated
- [ ] Midtrans settings will be verified

After deploying:

- [ ] Update backend CORS with Cloudflare domain
- [ ] Rebuild and restart backend
- [ ] Update Google OAuth authorized origins
- [ ] Test homepage loads
- [ ] Test authentication works
- [ ] Test shopping cart works
- [ ] Test checkout works
- [ ] Test admin panel works
- [ ] Test Biteship resi generation works

---

## 🎯 Expected Results

After successful deployment:

1. **Frontend URL:** `https://zavera-fashion-store.pages.dev`
2. **Build time:** 5-10 minutes
3. **Deploy time:** Instant (after build)
4. **SSL:** Automatically enabled
5. **CDN:** Global distribution
6. **Performance:** Fast loading worldwide

---

## 🚀 Ready to Deploy?

**Quick Start:**

1. Open: https://dash.cloudflare.com
2. Workers & Pages → Create application → Pages
3. Connect GitHub → Select ZAVERA-FASHION-STORE
4. Configure build settings (see Step 3 above)
5. Add environment variables (see Step 4 above)
6. Click "Save and Deploy"
7. Wait for build to complete
8. Update backend CORS
9. Test your site!

**Estimated total time:** 15-20 minutes

---

## 📞 Need Help?

If you encounter any issues:

1. Check Cloudflare Pages deployment logs
2. Check browser console for errors
3. Verify backend is accessible
4. Check CORS configuration
5. Test API endpoints manually

---

**Good luck with your deployment! 🎉**

Your Biteship auto-resi feature is ready to go live! 🚀

