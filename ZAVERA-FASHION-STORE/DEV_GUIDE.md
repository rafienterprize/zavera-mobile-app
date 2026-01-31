# ZAVERA Fashion Store - Development Guide

## 🚀 Quick Start Commands

### Start Frontend (Next.js)

```powershell
cd "c:\Users\ASUS\Desktop\ZAVERA FASHION STORE\frontend"
npm run dev
```

**Access:** http://localhost:3000

### Start Backend (Go)

```powershell
cd "c:\Users\ASUS\Desktop\ZAVERA FASHION STORE\backend"
go run main.go
```

**Access:** http://localhost:8080

### Or Use Batch File for Backend

```powershell
cd "c:\Users\ASUS\Desktop\ZAVERA FASHION STORE"
.\start-backend.bat
```

---

## 📁 Project Structure

```
ZAVERA FASHION STORE/
├── frontend/          # Next.js 14 App Router
│   ├── src/
│   │   ├── app/      # Pages & layouts
│   │   ├── components/  # Reusable components
│   │   └── context/  # React context (Cart, etc)
│   └── package.json
│
├── backend/          # Go + Gin Framework
│   ├── handler/     # HTTP handlers
│   ├── service/     # Business logic
│   ├── repository/  # Database access
│   ├── models/      # Domain models
│   ├── dto/         # Request/Response DTOs
│   ├── config/      # Database config
│   ├── routes/      # Route setup
│   └── main.go      # Entry point
│
└── database/        # PostgreSQL schemas
    ├── schema.sql   # Production schema
    └── migrate.sql  # Migration script
```

---

## 🗄️ Database Setup

### Create Database

```powershell
$env:PGPASSWORD='Yan2692009'
createdb -U postgres zavera_db
```

### Apply Schema

```powershell
cd "c:\Users\ASUS\Desktop\ZAVERA FASHION STORE\database"
$env:PGPASSWORD='Yan2692009'
psql -U postgres -d zavera_db -f migrate.sql
```

---

## 🔧 Environment Variables

### Backend (.env)

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=zavera_db
DB_USER=postgres
DB_PASSWORD=Yan2692009

MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_CLIENT_KEY=your-client-key
MIDTRANS_ENV=sandbox
```

### Frontend (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=your-client-key
```

---

## 📡 API Endpoints

### Products

- `GET /api/products` - Get all products
- `GET /api/products/:id` - Get product by ID

### Cart

- `GET /api/cart` - Get cart
- `POST /api/cart/items` - Add to cart
- `PUT /api/cart/items/:id` - Update quantity
- `DELETE /api/cart/items/:id` - Remove item
- `DELETE /api/cart` - Clear cart

### Orders

- `POST /api/checkout` - Create order & payment
- `GET /api/orders/:code` - Get order details
- `POST /api/payment/callback` - Midtrans webhook

---

## 🧪 Testing

### Test Backend Health

```powershell
curl http://localhost:8080/health
```

### Test Get Products

```powershell
curl http://localhost:8080/api/products
```

### Test Add to Cart

```powershell
curl -X POST http://localhost:8080/api/cart/items `
  -H "Content-Type: application/json" `
  -d '{"product_id": 1, "quantity": 2}' `
  --cookie-jar cookies.txt --cookie cookies.txt
```

---

## 📦 Installation

### Frontend Dependencies

```powershell
cd frontend
npm install
```

### Backend Dependencies

```powershell
cd backend
go mod download
```

---

## 🔄 Git Workflow

### After Making Changes

```powershell
cd "c:\Users\ASUS\Desktop\ZAVERA FASHION STORE"
git add .
git commit -m "feat: your commit message"
git push origin main
```

---

## ⚠️ Troubleshooting

### VS Code Shows Errors for Deleted Files

- Close and reopen VS Code
- Or run: `Ctrl+Shift+P` → "Developer: Reload Window"

### Port Already in Use

```powershell
# Kill process on port 8080
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Database Connection Error

- Check PostgreSQL is running
- Verify credentials in .env file
- Test connection: `psql -U postgres -d zavera_db`

---

## 🎯 Tech Stack

**Frontend:**

- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS
- Framer Motion
- Axios

**Backend:**

- Go 1.21+
- Gin Framework
- PostgreSQL
- database/sql
- Midtrans SDK

**Database:**

- PostgreSQL 12+
- 8 tables with proper indexes
- Order status lifecycle
- Payment integration

---

**Made with ❤️ for ZAVERA Fashion Store**
