# 🚀 Git Setup untuk ZAVERA Project

## 📦 Files yang Dibuat

1. **`.gitignore`** - File yang tidak akan di-commit ke Git
2. **`GIT_WORKFLOW_GUIDE.md`** - Panduan lengkap Git workflow
3. **`setup-git-repo.bat`** - Script untuk setup Git repository
4. **`git-daily-workflow.bat`** - Helper untuk workflow harian
5. **`README_GIT_SETUP.md`** - File ini

---

## 🎯 Quick Start

### Untuk Owner (Yang Buat Repo):

1. **Buat repository di GitHub:**
   - Buka https://github.com/new
   - Nama: `zavera-fashion-store` (atau terserah)
   - Visibility: Private (recommended)
   - Jangan centang "Initialize with README"
   - Klik "Create repository"

2. **Setup local repository:**
   ```bash
   # Double-click file ini:
   setup-git-repo.bat
   
   # Atau manual:
   git init
   git add .
   git commit -m "Initial commit"
   git branch -M main
   git remote add origin https://github.com/USERNAME/zavera-fashion-store.git
   git push -u origin main
   ```

3. **Invite collaborator:**
   - GitHub → Settings → Collaborators
   - Add people → Masukkan username/email temen
   - Temen akan dapat email invitation

### Untuk Collaborator (Yang Di-invite):

1. **Accept invitation** (cek email)

2. **Clone repository:**
   ```bash
   git clone https://github.com/USERNAME/zavera-fashion-store.git
   cd zavera-fashion-store
   ```

3. **Setup Git config:**
   ```bash
   git config user.name "Nama Kamu"
   git config user.email "email@example.com"
   ```

---

## 💼 Daily Workflow

### Cara Gampang (Pakai Script):

```bash
# Double-click file ini:
git-daily-workflow.bat

# Pilih menu:
# 1 = Buat branch baru
# 2 = Commit changes
# 3 = Push ke GitHub
# 4 = Pull latest
# 5 = Buat Pull Request
```

### Cara Manual:

```bash
# 1. Pull latest
git checkout main
git pull origin main

# 2. Buat branch baru
git checkout -b mobile/new-feature

# 3. Edit files...

# 4. Commit
git add .
git commit -m "feat: add new feature"

# 5. Push
git push origin mobile/new-feature

# 6. Buat Pull Request di GitHub
```

---

## 🌿 Branch Strategy

```
main (production)
├── mobile/logo-update
├── mobile/api-integration
├── backend/payment-webhook
├── frontend/checkout-ui
└── fix/cart-bug
```

**Rules:**
- `main` = production code (always stable)
- Jangan push langsung ke `main`
- Buat branch untuk setiap fitur
- Merge via Pull Request

---

## 📝 Commit Message Examples

```bash
# Good ✅
git commit -m "feat: add product search"
git commit -m "fix: resolve cart quantity bug"
git commit -m "docs: update API documentation"

# Bad ❌
git commit -m "update"
git commit -m "fix bug"
git commit -m "changes"
```

---

## 🔄 Pull Request Flow

1. **Push branch ke GitHub**
   ```bash
   git push origin your-branch
   ```

2. **Buat PR di GitHub:**
   - Go to repository
   - Click "Pull requests" → "New pull request"
   - Base: `main` ← Compare: `your-branch`
   - Add title & description
   - Click "Create pull request"

3. **Review & Discuss:**
   - Temen review code
   - Diskusi kalau ada yang perlu diubah
   - Push lagi kalau ada update

4. **Merge:**
   - Klik "Merge pull request"
   - Delete branch (optional)

5. **Update local:**
   ```bash
   git checkout main
   git pull origin main
   git branch -d your-branch
   ```

---

## ⚠️ Common Issues & Solutions

### Issue: "Permission denied"
```bash
# Solution: Check remote URL
git remote -v
git remote set-url origin https://github.com/USERNAME/repo.git
```

### Issue: "Your branch is behind"
```bash
# Solution: Pull first
git pull origin main
```

### Issue: Merge conflict
```bash
# Solution:
# 1. Open conflicted file
# 2. Find markers: <<<<<<< HEAD
# 3. Choose code to keep
# 4. Remove markers
# 5. Save file
git add .
git commit -m "fix: resolve merge conflict"
```

### Issue: Forgot branch name
```bash
# Solution:
git branch -a
```

---

## 📚 Resources

- **Full Guide:** Baca `GIT_WORKFLOW_GUIDE.md`
- **Git Docs:** https://git-scm.com/doc
- **GitHub Docs:** https://docs.github.com
- **Interactive Tutorial:** https://learngitbranching.js.org/

---

## 🎓 Tips & Best Practices

1. **Commit sering** - Jangan tunggu sampai banyak perubahan
2. **Pull sebelum push** - Hindari conflict
3. **Branch naming** - Gunakan format yang jelas
4. **Commit message** - Jelas dan deskriptif
5. **Review code** - Sebelum merge, review dulu
6. **Test dulu** - Pastikan code jalan sebelum push
7. **Jangan commit secrets** - .env, passwords, API keys
8. **Keep branch updated** - Merge main ke branch kamu secara berkala

---

## 🆘 Need Help?

1. Baca `GIT_WORKFLOW_GUIDE.md`
2. Tanya temen yang lebih expert
3. Google: "git [problem] stackoverflow"
4. GitHub Discussions

---

## ✅ Checklist Setup

- [ ] Git installed
- [ ] GitHub account created
- [ ] Repository created (owner)
- [ ] Collaborator invited (owner)
- [ ] Repository cloned (collaborator)
- [ ] Git config setup (name & email)
- [ ] Read `GIT_WORKFLOW_GUIDE.md`
- [ ] Try first commit & push
- [ ] Try create Pull Request

---

**Happy Coding! 🚀**

*Last updated: 2026-02-01*
