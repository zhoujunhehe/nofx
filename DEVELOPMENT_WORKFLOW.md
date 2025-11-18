# NoFxAiOS/nofxos 私庫開發流程

## 📂 專案定位
- **路徑**: `~/Documents/GitHub/repos-thedevz/nofxos-private/`
- **身份**: `the-dev-z` (自動配置)
- **權限**: TRIAGE (無法直接 push)
- **協作模式**: Pull Model (本地開發 + Patch 提交)
- **倉庫限制**: `allow_forking: false` (維護者禁用 fork 功能)

---

## 🔄 開發流程

### 1. 同步上游最新代碼
```bash
cd ~/Documents/GitHub/repos-thedevz/nofxos-private

# 拉取最新代碼
git fetch upstream

# 更新本地 dev 分支（主開發分支）
git checkout dev
git merge upstream/dev

# 或者硬重置（如果本地沒有重要修改）
git reset --hard upstream/dev
```

### 2. 創建功能分支
```bash
# 基於 dev 分支創建新功能分支
git checkout -b feature/your-feature-name

# 或基於 main 分支（根據維護者要求）
git checkout -b fix/bug-name upstream/main
```

### 3. 開發與提交
```bash
# 修改代碼...

# 查看變更
git status
git diff

# 提交變更
git add .
git commit -m "feat: implement your feature"

# 多次提交（保持每個 commit 原子化）
git commit -m "fix: resolve edge case"
git commit -m "docs: update README"
```

### 4. 生成 Patch 文件
```bash
# 方式 A：生成單個文件包含所有提交（推薦）
git format-patch upstream/dev --stdout > ~/Desktop/your-feature-name.patch

# 方式 B：為每個 commit 生成獨立 patch
git format-patch upstream/dev
# 輸出：0001-feat-implement-your-feature.patch
#      0002-fix-resolve-edge-case.patch
#      ...

# 方式 C：只生成最近一次提交的 patch
git format-patch -1 HEAD
```

### 5. 提交 Patch 給維護者（推薦：Issue + Gist）

**步驟 5.1：生成 Patch**
```bash
git patch-dev my-feature
# 輸出：~/Desktop/my-feature.patch
```

**步驟 5.2：上傳到 GitHub Gist**
```bash
# 使用 gh CLI 創建 gist
gh gist create ~/Desktop/my-feature.patch \
  --desc "feat: [your feature name]" \
  --public
# 輸出 gist URL，例如：https://gist.github.com/the-dev-z/abc123...
```

**步驟 5.3：創建 Issue 並附上 Patch**
```bash
gh issue create --repo NoFxAiOS/nofxos \
  --title "feat: [your feature name]" \
  --body "## 📋 功能描述
[簡短描述這個功能的目的和實現]

## 🔄 如何應用此 Patch
\`\`\`bash
# 下載並應用 patch
curl -L https://gist.github.com/the-dev-z/abc123.../raw | git am

# 或手動下載後應用
gh gist view abc123 --raw > feature.patch
git am < feature.patch
\`\`\`

## 📊 變更統計
- Commits: X 個
- Files changed: Y 個
- Lines added: +Z

## 🧪 測試
- [ ] 本地測試通過
- [ ] 符合代碼規範

**Patch Gist**: https://gist.github.com/the-dev-z/abc123..."
```

**其他選項**（不推薦）：
- 選項 A：Discord/Telegram 私訊（無公開記錄）
- 選項 B：臨時私庫（需額外配置）

---

## 📦 Patch 文件範例

**生成：**
```bash
git format-patch upstream/dev --stdout > agent-wallet-enhancement.patch
```

**維護者應用：**
```bash
# 選項 1：直接應用到當前分支
git am < agent-wallet-enhancement.patch

# 選項 2：創建審查分支
git checkout -b review/sotadic-agent-wallet
git am < agent-wallet-enhancement.patch
git push origin review/sotadic-agent-wallet
# 在 GitHub 創建 PR
```

---

## 🔍 常用指令

### 查看分支狀態
```bash
# 查看所有分支
git branch -a

# 查看上游遠程分支
git remote show upstream

# 查看當前分支的 commit 歷史
git log --oneline --graph --decorate
```

### 清理本地分支
```bash
# 刪除已合併的功能分支
git branch -d feature/old-feature

# 強制刪除未合併的分支
git branch -D feature/abandoned-feature
```

### 重置到上游狀態（危險！）
```bash
# 放棄本地所有修改，完全重置到上游
git fetch upstream
git reset --hard upstream/dev
git clean -fd
```

---

## 🎯 分支策略

根據你的分析，這個專案可能使用：
- `main`: 穩定版本（生產環境）
- `dev`: 主開發分支（推薦基於此創建功能分支）
- `beta`: 測試版本
- `hotfix/*`: 緊急修復
- `feature/*`: 新功能開發
- `fix/*`: Bug 修復

**建議：** 開發新功能時基於 `dev` 分支創建。

---

## ⚠️ 注意事項

1. **不要嘗試 push**：你只有 TRIAGE 權限，push 會失敗
2. **無法 fork**：倉庫設定 `allow_forking: false`，GitHub UI 不會顯示 fork 按鈕
3. **保持 commit 整潔**：使用 `git rebase -i` 整理 commit 歷史
4. **敏感信息**：絕對不要在 patch 中包含 API keys、私鑰等
5. **測試通過後再提交**：確保代碼能通過測試
6. **遵循代碼規範**：參考專案的 CONTRIBUTING.md（如果有）

---

## 🆘 如果需要 WRITE 權限

聯絡維護者：
```
嗨，目前我在 NoFxAiOS/nofxos 只有 TRIAGE 權限，
無法直接 push 分支。能否升級權限到 WRITE？
這樣我可以：
- 直接在主庫創建功能分支
- 提交 PR（無需 patch 文件）
- 參與 Code Review
```

---

## 📚 相關資源

- **開源版 nofx**: `~/Documents/GitHub/nofx/`
- **私庫版 nofxos**: `~/Documents/GitHub/repos-thedevz/nofxos-private/`
- **多帳號配置**: `~/.gitconfig` (Conditional Includes)
- **SSH 配置**: `~/.ssh/config` (Host aliases)

---

## 🔍 為什麼用 Patch 工作流？

**技術限制**：
```bash
$ gh api repos/NoFxAiOS/nofxos | jq '.allow_forking, .permissions'
false
{"admin": false, "push": false, "triage": true, "pull": true}
```

**Patch 工作流的優勢**（在此環境下）：
- ✅ 無需額外權限即可貢獻代碼
- ✅ 提交歷史保留完整的 author 信息
- ✅ 維護者可以自由選擇審查和合併的方式
- ✅ 適合偶爾性的貢獻（無需維護 fork）

**如需更頻繁協作**：建議向維護者申請 WRITE 權限，轉用標準 PR 工作流。

---

*最後更新：2025-11-13*
