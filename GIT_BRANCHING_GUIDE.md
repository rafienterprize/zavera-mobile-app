# Git Branching untuk Tim - Panduan Lengkap

## 🌳 Konsep Branch

Branch itu seperti "cabang" development yang terpisah. Setiap orang bisa kerja di branch sendiri tanpa ganggu orang lain.

```
main (production)
├── feature/login (kamu)
├── feature/payment (teman 1)
└── feature/cart (teman 2)
```

## 📋 Struktur Branch yang Umum Dipakai

### 1. Branch Utama
- **main** atau **master** → Production code (yang sudah jadi)
- **develop** → Development code (untuk testing)

### 2. Branch Fitur (per orang/per fitur)
- **feature/nama-fitur** → Untuk fitur baru
- **bugfix/nama-bug** → Untuk fix bug
- **hotfix/nama-urgent** → Untuk fix urgent di production

## 🚀 Workflow untuk Tim

### Setup Awal (Sekali aja)

```bash
# 1. Clone repo
git clone https://github.com/username/repo-name.git
cd repo-name

# 2. Cek branch yang ada
git branch -a

# 3. Bikin branch develop (kalau belum ada)
git checkout -b develop
git push -u origin develop
```

### Workflow Harian (Setiap Kerja)

#### Kamu kerja di fitur baru:

```bash
# 1. Pastikan di branch develop dulu
git checkout develop

# 2. Pull update terbaru dari tim
git pull origin develop

# 3. Bikin branch baru untuk fitur kamu
git checkout -b feature/login-screen

# 4. Kerja di fitur kamu...
# Edit file, coding, dll

# 5. Commit perubahan
git add .
git commit -m "feat: add login screen UI"

# 6. Push ke GitHub
git push -u origin feature/login-screen
```

#### Teman kamu kerja di fitur lain:

```bash
# Teman 1
git checkout -b feature/payment-integration
# coding...
git push -u origin feature/payment-integration

# Teman 2
git checkout -b feature/cart-system
# coding...
git push -u origin feature/cart-system
```

## 🔄 Merge Branch (Gabungin Kode)

### Cara 1: Merge Langsung (Simple)

```bash
# 1. Pindah ke branch develop
git checkout develop

# 2. Pull update terbaru
git pull origin develop

# 3. Merge branch fitur kamu
git merge feature/login-screen

# 4. Push hasil merge
git push origin develop

# 5. Hapus branch fitur (opsional)
git branch -d feature/login-screen
git push origin --delete feature/login-screen
```

### Cara 2: Pull Request (Recommended untuk Tim)

1. Push branch kamu ke GitHub
2. Buka GitHub → Tab "Pull Requests"
3. Klik "New Pull Request"
4. Pilih: `develop` ← `feature/login-screen`
5. Tulis deskripsi perubahan
6. Klik "Create Pull Request"
7. Minta teman review
8. Setelah approved, klik "Merge Pull Request"

## 🛠️ Command Penting

### Cek Status
```bash
# Cek branch sekarang
git branch

# Cek semua branch (termasuk remote)
git branch -a

# Cek status file
git status
```

### Pindah Branch
```bash
# Pindah ke branch yang sudah ada
git checkout develop

# Bikin branch baru dan langsung pindah
git checkout -b feature/new-feature
```

### Update Branch
```bash
# Pull update dari GitHub
git pull origin develop

# Atau fetch dulu, baru merge
git fetch origin
git merge origin/develop
```

### Hapus Branch
```bash
# Hapus branch lokal
git branch -d feature/login-screen

# Hapus branch di GitHub
git push origin --delete feature/login-screen
```

## 🎯 Naming Convention Branch

### Format: `type/nama-fitur`

**Type:**
- `feature/` → Fitur baru
- `bugfix/` → Fix bug
- `hotfix/` → Fix urgent
- `refactor/` → Refactor code
- `docs/` → Update dokumentasi

**Contoh:**
```
feature/login-screen
feature/payment-integration
bugfix/cart-total-calculation
hotfix/crash-on-startup
refactor/api-service
docs/update-readme
```

## 🔥 Skenario Umum

### Skenario 1: Kamu dan Teman Kerja Bersamaan

```bash
# Kamu
git checkout -b feature/login
# coding login...
git push origin feature/login

# Teman
git checkout -b feature/payment
# coding payment...
git push origin feature/payment

# Nanti merge satu-satu ke develop
```

### Skenario 2: Conflict (Bentrok)

Kalau kamu dan teman edit file yang sama:

```bash
# Saat merge, muncul conflict
git merge feature/login
# CONFLICT in lib/main.dart

# 1. Buka file yang conflict
# 2. Cari tanda <<<<<<< dan >>>>>>>
# 3. Edit manual, pilih mana yang mau dipakai
# 4. Hapus tanda conflict
# 5. Commit hasil fix

git add .
git commit -m "fix: resolve merge conflict"
git push
```

### Skenario 3: Update Branch Kamu dengan Develop Terbaru

```bash
# Kamu lagi di feature/login
# Tapi develop sudah update

# Cara 1: Merge
git checkout feature/login
git merge develop

# Cara 2: Rebase (lebih bersih)
git checkout feature/login
git rebase develop
```

## 📝 Commit Message Convention

Format: `type: deskripsi singkat`

**Type:**
- `feat:` → Fitur baru
- `fix:` → Fix bug
- `refactor:` → Refactor code
- `docs:` → Update docs
- `style:` → Format code
- `test:` → Tambah test
- `chore:` → Update config

**Contoh:**
```bash
git commit -m "feat: add login screen with email validation"
git commit -m "fix: resolve cart total calculation bug"
git commit -m "refactor: improve API service structure"
```

## 🎨 Visualisasi Workflow

```
main (production)
  │
  ├─ develop (testing)
  │   │
  │   ├─ feature/login (kamu)
  │   │   ├─ commit 1
  │   │   ├─ commit 2
  │   │   └─ merge → develop
  │   │
  │   ├─ feature/payment (teman 1)
  │   │   ├─ commit 1
  │   │   └─ merge → develop
  │   │
  │   └─ feature/cart (teman 2)
  │       ├─ commit 1
  │       └─ merge → develop
  │
  └─ merge develop → main (release)
```

## 🚨 Tips Penting

1. **Selalu pull sebelum mulai kerja**
   ```bash
   git pull origin develop
   ```

2. **Commit sering, push sering**
   - Jangan tunggu selesai semua
   - Commit setiap fitur kecil selesai

3. **Branch naming harus jelas**
   - ❌ `git checkout -b test`
   - ✅ `git checkout -b feature/login-screen`

4. **Jangan kerja langsung di main/develop**
   - Selalu bikin branch baru

5. **Komunikasi dengan tim**
   - Kasih tau kalau mau merge
   - Review code teman sebelum merge

## 🔧 Setup .gitignore

Pastikan file sensitif tidak ke-push:

```gitignore
# Environment
.env
.env.local

# Dependencies
node_modules/
vendor/

# Build
build/
dist/

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
```

## 📱 Contoh untuk Project Kamu

```bash
# Setup awal
git checkout -b develop
git push -u origin develop

# Kamu kerja mobile
git checkout -b feature/mobile-login
# coding...
git add .
git commit -m "feat: add mobile login screen"
git push -u origin feature/mobile-login

# Teman kerja backend
git checkout -b feature/backend-auth
# coding...
git push -u origin feature/backend-auth

# Merge via Pull Request di GitHub
# Atau merge manual:
git checkout develop
git merge feature/mobile-login
git merge feature/backend-auth
git push origin develop
```

## 🎓 Cheat Sheet

```bash
# Lihat branch
git branch -a

# Bikin branch baru
git checkout -b feature/nama-fitur

# Pindah branch
git checkout develop

# Update branch
git pull origin develop

# Commit
git add .
git commit -m "feat: deskripsi"

# Push
git push origin feature/nama-fitur

# Merge
git checkout develop
git merge feature/nama-fitur

# Hapus branch
git branch -d feature/nama-fitur
git push origin --delete feature/nama-fitur
```
