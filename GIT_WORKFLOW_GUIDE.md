# Git Workflow Guide - ZAVERA Project

## 📋 Table of Contents
1. [Setup Awal](#setup-awal)
2. [Workflow Harian](#workflow-harian)
3. [Branch Naming](#branch-naming)
4. [Commit Messages](#commit-messages)
5. [Pull Request](#pull-request)
6. [Resolve Conflicts](#resolve-conflicts)
7. [Quick Commands](#quick-commands)

---

## 🚀 Setup Awal

### Owner (Temen yang buat repo):

```bash
# 1. Buat repo di GitHub dulu (via website)
# 2. Di folder project:
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/USERNAME/zavera-project.git
git push -u origin main

# 3. Invite collaborator di GitHub:
# Settings → Collaborators → Add people
```

### Collaborator (Kamu):

```bash
# 1. Clone repo
git clone https://github.com/USERNAME/zavera-project.git
cd zavera-project

# 2. Setup git config (sekali aja)
git config user.name "Nama Kamu"
git config user.email "email@example.com"

# 3. Cek remote
git remote -v
```

---

## 💼 Workflow Harian

### Sebelum Mulai Kerja:

```bash
# 1. Pastikan di branch main
git checkout main

# 2. Pull perubahan terbaru
git pull origin main

# 3. Buat branch baru untuk fitur kamu
git checkout -b mobile/logo-update
```

### Saat Kerja:

```bash
# 1. Edit files...
# 2. Cek perubahan
git status

# 3. Add files yang diubah
git add .
# atau specific file:
git add zavera_mobile/lib/main.dart

# 4. Commit dengan message yang jelas
git commit -m "feat: update app logo and name"

# 5. Push ke GitHub
git push origin mobile/logo-update
```

### Setelah Selesai:

```bash
# 1. Buat Pull Request di GitHub
# 2. Tunggu review & merge
# 3. Setelah di-merge, update local:
git checkout main
git pull origin main
git branch -d mobile/logo-update  # hapus branch lokal
```

---

## 🌿 Branch Naming Convention

```
mobile/feature-name       → Mobile app features
backend/feature-name      → Backend API features
frontend/feature-name     → Frontend website features
fix/bug-description       → Bug fixes
hotfix/critical-bug       → Urgent production fixes
docs/update-readme        → Documentation updates
refactor/code-cleanup     → Code refactoring
```

**Contoh:**
- `mobile/payment-integration`
- `backend/webhook-handler`
- `frontend/checkout-ui`
- `fix/cart-quantity-bug`
- `hotfix/payment-crash`

---

## 📝 Commit Message Convention

Format: `type: description`

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting, no code change
- `refactor`: Code restructure
- `test`: Add tests
- `chore`: Maintenance

**Contoh:**
```bash
git commit -m "feat: add product search functionality"
git commit -m "fix: resolve cart quantity bug"
git commit -m "docs: update API documentation"
git commit -m "refactor: optimize database queries"
```

---

## 🔄 Pull Request Process

### 1. Buat Pull Request di GitHub:
- Go to: https://github.com/USERNAME/zavera-project/pulls
- Click "New Pull Request"
- Base: `main` ← Compare: `your-branch`
- Add title & description
- Click "Create Pull Request"

### 2. Review Process:
- Temen review code kamu
- Diskusi kalau ada yang perlu diubah
- Kamu bisa push lagi ke branch yang sama untuk update

### 3. Merge:
- Setelah approved, klik "Merge Pull Request"
- Delete branch (optional)

### 4. Update Local:
```bash
git checkout main
git pull origin main
```

---

## ⚠️ Resolve Merge Conflicts

### Kalau ada conflict saat pull/merge:

```bash
# 1. Pull dari main
git pull origin main
# Output: CONFLICT in file.dart

# 2. Buka file yang conflict
# Cari marker:
<<<<<<< HEAD
your code
=======
their code
>>>>>>> branch-name

# 3. Edit file:
# - Pilih code yang mau dipakai
# - Atau gabungkan keduanya
# - Hapus semua marker (<<<, ===, >>>)

# 4. Save file, lalu:
git add .
git commit -m "fix: resolve merge conflict in file.dart"
git push
```

---

## ⚡ Quick Commands Cheat Sheet

### Status & Info:
```bash
git status                    # Lihat perubahan
git log --oneline            # Lihat history
git branch -a                # Lihat semua branch
git remote -v                # Lihat remote URL
```

### Branch Operations:
```bash
git checkout main            # Pindah ke main
git checkout -b new-branch   # Buat & pindah ke branch baru
git branch -d branch-name    # Hapus branch lokal
git push origin --delete branch-name  # Hapus branch remote
```

### Sync & Update:
```bash
git pull origin main         # Pull dari main
git fetch origin             # Fetch tanpa merge
git merge main               # Merge main ke branch current
```

### Undo Changes:
```bash
git checkout -- file.txt     # Batalkan perubahan file
git reset HEAD file.txt      # Unstage file
git reset --hard HEAD        # Batalkan semua perubahan (HATI-HATI!)
git revert commit-hash       # Revert commit tertentu
```

### Stash (Simpan sementara):
```bash
git stash                    # Simpan perubahan sementara
git stash list               # Lihat stash list
git stash pop                # Apply & hapus stash terakhir
git stash apply              # Apply tanpa hapus
```

---

## 🎯 Best Practices

1. **Selalu pull sebelum mulai kerja**
   ```bash
   git checkout main
   git pull origin main
   ```

2. **Commit sering dengan message yang jelas**
   - Jangan commit file besar (>100MB)
   - Jangan commit file sensitive (.env, passwords)

3. **Buat branch untuk setiap fitur**
   - Jangan kerja langsung di main
   - 1 branch = 1 fitur/fix

4. **Review code sebelum merge**
   - Cek perubahan di Pull Request
   - Test dulu sebelum merge

5. **Keep branch up-to-date**
   ```bash
   git checkout your-branch
   git merge main
   ```

---

## 🆘 Troubleshooting

### "Permission denied" saat push:
```bash
# Setup SSH key atau gunakan HTTPS dengan token
git remote set-url origin https://github.com/USERNAME/repo.git
```

### "Your branch is behind":
```bash
git pull origin main
```

### "Detached HEAD state":
```bash
git checkout main
```

### Lupa nama branch:
```bash
git branch -a
```

---

## 📞 Need Help?

- Tanya temen yang lebih expert
- Baca: https://git-scm.com/doc
- GitHub Docs: https://docs.github.com

**Remember:** Git itu tool, bukan musuh. Practice makes perfect! 🚀
