# Git Cheat Sheet - Quick Reference

## 🚀 Setup & Config

```bash
# Setup user
git config user.name "Your Name"
git config user.email "your@email.com"

# Check config
git config --list

# Initialize repo
git init

# Clone repo
git clone https://github.com/user/repo.git
```

---

## 📊 Status & Info

```bash
# Check status
git status

# View history
git log
git log --oneline
git log --graph --oneline --all

# View branches
git branch
git branch -a              # all branches
git branch -r              # remote branches

# View remotes
git remote -v

# View changes
git diff                   # unstaged changes
git diff --staged          # staged changes
git diff branch1 branch2   # compare branches
```

---

## 🌿 Branches

```bash
# Create branch
git branch branch-name

# Create & switch
git checkout -b branch-name

# Switch branch
git checkout branch-name

# Delete branch
git branch -d branch-name              # safe delete
git branch -D branch-name              # force delete
git push origin --delete branch-name   # delete remote

# Rename branch
git branch -m old-name new-name

# Merge branch
git checkout main
git merge branch-name
```

---

## 💾 Commit & Push

```bash
# Stage files
git add file.txt           # specific file
git add .                  # all files
git add *.js               # pattern

# Unstage files
git reset HEAD file.txt

# Commit
git commit -m "message"
git commit -am "message"   # add + commit

# Amend last commit
git commit --amend -m "new message"

# Push
git push origin branch-name
git push -u origin branch-name   # set upstream
git push --force                 # force push (CAREFUL!)

# Push all branches
git push --all origin
```

---

## 🔄 Pull & Fetch

```bash
# Pull (fetch + merge)
git pull origin main

# Fetch only
git fetch origin

# Pull with rebase
git pull --rebase origin main
```

---

## ↩️ Undo Changes

```bash
# Discard unstaged changes
git checkout -- file.txt
git restore file.txt

# Unstage file
git reset HEAD file.txt
git restore --staged file.txt

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1

# Revert commit (create new commit)
git revert commit-hash

# Reset to specific commit
git reset --hard commit-hash

# Clean untracked files
git clean -fd              # remove files & directories
git clean -n               # dry run (preview)
```

---

## 📦 Stash (Temporary Save)

```bash
# Save changes
git stash
git stash save "message"

# List stashes
git stash list

# Apply stash
git stash pop              # apply & remove
git stash apply            # apply & keep
git stash apply stash@{0}  # specific stash

# Drop stash
git stash drop stash@{0}
git stash clear            # remove all
```

---

## 🔀 Merge & Rebase

```bash
# Merge branch
git checkout main
git merge feature-branch

# Abort merge
git merge --abort

# Rebase
git checkout feature-branch
git rebase main

# Abort rebase
git rebase --abort

# Continue rebase (after resolving conflicts)
git rebase --continue
```

---

## 🏷️ Tags

```bash
# Create tag
git tag v1.0.0
git tag -a v1.0.0 -m "Version 1.0.0"

# List tags
git tag

# Push tags
git push origin v1.0.0
git push origin --tags     # all tags

# Delete tag
git tag -d v1.0.0
git push origin --delete v1.0.0
```

---

## 🔍 Search & Find

```bash
# Search in files
git grep "search term"

# Find commit by message
git log --grep="search term"

# Find who changed a line
git blame file.txt

# Show commit details
git show commit-hash
```

---

## 🌐 Remote

```bash
# Add remote
git remote add origin https://github.com/user/repo.git

# Change remote URL
git remote set-url origin https://github.com/user/new-repo.git

# Remove remote
git remote remove origin

# Rename remote
git remote rename old-name new-name

# Fetch all remotes
git fetch --all
```

---

## 🔧 Advanced

```bash
# Cherry-pick commit
git cherry-pick commit-hash

# Interactive rebase
git rebase -i HEAD~3

# Squash commits
git rebase -i HEAD~3
# Change 'pick' to 'squash' for commits to merge

# Show file at specific commit
git show commit-hash:path/to/file

# Restore file from another branch
git checkout other-branch -- path/to/file

# Create patch
git diff > changes.patch

# Apply patch
git apply changes.patch
```

---

## 🆘 Troubleshooting

```bash
# Detached HEAD - go back to branch
git checkout main

# Recover deleted branch
git reflog
git checkout -b branch-name commit-hash

# Undo git add (before commit)
git reset

# Undo git commit (keep changes)
git reset --soft HEAD~1

# Fix "Your branch is behind"
git pull origin main

# Fix "Your branch is ahead"
git push origin main

# Fix merge conflicts
# 1. Edit conflicted files
# 2. Remove conflict markers (<<<, ===, >>>)
# 3. git add .
# 4. git commit
```

---

## 📋 Commit Message Convention

```
feat: new feature
fix: bug fix
docs: documentation
style: formatting
refactor: code restructure
test: add tests
chore: maintenance
perf: performance improvement
ci: CI/CD changes
build: build system changes
revert: revert previous commit
```

**Examples:**
```bash
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login redirect issue"
git commit -m "docs: update API documentation"
git commit -m "refactor: optimize database queries"
```

---

## 🎯 Best Practices

1. **Commit often** - Small, focused commits
2. **Pull before push** - Avoid conflicts
3. **Use branches** - Don't work on main
4. **Write clear messages** - Explain what & why
5. **Review before merge** - Check changes
6. **Test before push** - Ensure code works
7. **Don't commit secrets** - Use .gitignore
8. **Keep history clean** - Use rebase when appropriate

---

## 🔗 Useful Aliases

Add to `~/.gitconfig`:

```ini
[alias]
    st = status
    co = checkout
    br = branch
    ci = commit
    unstage = reset HEAD --
    last = log -1 HEAD
    visual = log --graph --oneline --all
    undo = reset --soft HEAD~1
```

Usage:
```bash
git st              # instead of git status
git co main         # instead of git checkout main
git br              # instead of git branch
git visual          # pretty log
```

---

## 📚 Resources

- Git Docs: https://git-scm.com/doc
- GitHub Docs: https://docs.github.com
- Interactive Tutorial: https://learngitbranching.js.org/
- Git Cheat Sheet PDF: https://education.github.com/git-cheat-sheet-education.pdf

---

**Print this and keep it handy! 📄**
