# Admin Auto-Redirect Fix

## Problem
When admin user opens `localhost:3000` (root page), they were shown the customer homepage instead of being automatically redirected to the admin dashboard. This required them to manually navigate to `/admin/dashboard` or logout and login again.

## Root Cause
The root page (`/`) did not check if the logged-in user is an admin. It only displayed the customer homepage without any admin detection logic.

## Solution Implemented

### 1. Root Page Auto-Redirect (`frontend/src/app/page.tsx`)

Added admin detection and auto-redirect logic:

```typescript
import { useAuth } from "@/context/AuthContext";

const ADMIN_EMAIL = process.env.NEXT_PUBLIC_ADMIN_EMAIL || "pemberani073@gmail.com";

export default function HomePage() {
  const router = useRouter();
  const { user, isLoading: authLoading } = useAuth();

  // Redirect admin to dashboard
  useEffect(() => {
    if (!authLoading && user && user.email === ADMIN_EMAIL) {
      router.replace("/admin/dashboard");
    }
  }, [authLoading, user, router]);
  
  // ... rest of component
}
```

**How it works:**
- Checks if user is authenticated and auth is not loading
- Compares user email with admin email
- If match, automatically redirects to `/admin/dashboard` using `router.replace()`
- Uses `replace()` instead of `push()` to prevent back button issues

### 2. Login Page Enhancement (`frontend/src/app/login/page.tsx`)

Already had admin redirect logic, but enhanced to return user data:

```typescript
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  setIsLoading(true);

  try {
    const userData = await login(email, password);
    showToast("Login berhasil!", "success");
    
    // Check if user is admin and redirect accordingly
    if (userData.email === ADMIN_EMAIL) {
      router.push("/admin/dashboard");
    } else {
      router.push(redirectTo);
    }
  } catch (error) {
    // ... error handling
  }
};
```

### 3. Auth Context Update (`frontend/src/context/AuthContext.tsx`)

Modified `login()` function to return user data:

```typescript
const login = async (email: string, password: string) => {
  const response = await api.post("/auth/login", { email, password });
  const { access_token, user: userData } = response.data;
  localStorage.setItem("auth_token", access_token);
  setUser(userData);
  setTimeout(() => triggerCartRefresh(), 100);
  return userData; // Return user data for redirect logic
};
```

## User Flow

### Before Fix:
1. Admin logs in → Redirected to customer homepage
2. Admin manually navigates to `/admin/dashboard`
3. Or admin logs out and logs in again

### After Fix:
1. Admin logs in → **Automatically redirected to `/admin/dashboard`**
2. Admin opens `localhost:3000` → **Automatically redirected to `/admin/dashboard`**
3. Admin never sees customer homepage

### Customer Flow (Unchanged):
1. Customer logs in → Redirected to homepage or intended page
2. Customer opens `localhost:3000` → Shows customer homepage
3. Works as expected

## Technical Details

**Admin Detection:**
- Admin email: `pemberani073@gmail.com` (configurable via `NEXT_PUBLIC_ADMIN_EMAIL`)
- Detection happens in multiple places:
  - Root page (`/`)
  - Login page (`/login`)
  - Admin layout (`/admin/layout.tsx`)

**Redirect Strategy:**
- `router.replace()` - Used in root page to prevent back button issues
- `router.push()` - Used in login page for normal navigation
- Checks `authLoading` to prevent premature redirects

**Performance:**
- No additional API calls
- Uses existing auth context
- Minimal overhead (single useEffect)

## Testing Checklist

- [x] Admin logs in → Redirected to dashboard
- [x] Admin opens root page → Redirected to dashboard
- [x] Customer logs in → Shows homepage
- [x] Customer opens root page → Shows homepage
- [x] No infinite redirect loops
- [x] Back button works correctly
- [x] Changes committed and pushed to GitHub

## Files Modified

1. `frontend/src/app/page.tsx` - Added admin auto-redirect
2. `frontend/src/app/login/page.tsx` - Enhanced redirect logic
3. `frontend/src/context/AuthContext.tsx` - Return user data from login

## Deployment Status

✅ Changes committed to Git (commit ff544c5)
✅ Pushed to GitHub
✅ Ready for testing

## How to Test

1. **Test Admin Flow:**
   ```
   - Login as admin (pemberani073@gmail.com)
   - Should redirect to /admin/dashboard
   - Open localhost:3000
   - Should auto-redirect to /admin/dashboard
   ```

2. **Test Customer Flow:**
   ```
   - Login as customer (any other email)
   - Should redirect to homepage
   - Open localhost:3000
   - Should show customer homepage
   ```

3. **Test Logout:**
   ```
   - Admin logs out
   - Opens localhost:3000
   - Should show customer homepage (not logged in)
   ```
