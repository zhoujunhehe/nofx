# 🚀 Railway 部署完全指南（小白版）

> 这是一份面向完全零基础用户的部署教程。跟着步骤一步步做，10-15 分钟内即可完成部署。

## 📋 你需要准备什么？

在开始之前，请准备好：

1. ✅ 一个 GitHub 账号（[免费注册](https://github.com/signup)）
2. ✅ 一个 Railway 账号（[免费注册](https://railway.app/)）
3. ✅ 一台能上网的电脑（Windows/Mac/Linux 都可以）
4. ✅ 15 分钟的时间

**费用说明：**
- Railway 提供 $5 免费额度（每月）
- 本项目预计消耗：约 $3-5/月
- 首月可以免费试用

---

## 第一步：Fork 项目到你的 GitHub

### 1.1 访问项目仓库

打开浏览器，访问：
```
https://github.com/NoFxAiOS/nofx
```

### 1.2 Fork 项目

1. 点击右上角的 **"Fork"** 按钮
2. 选择你的账号
3. 点击 **"Create fork"**
4. 等待几秒钟，Fork 完成

✅ **完成标志：** 你会看到类似 `你的用户名/nofx` 的新仓库

---

## 第二步：生成加密密钥

你需要生成两个密钥：`DATA_ENCRYPTION_KEY` 和 `JWT_SECRET`

### 2.1 打开在线密钥生成器

访问：[https://generate-secret.vercel.app/32](https://generate-secret.vercel.app/32)

或者使用这个备用网站：[https://www.random.org/strings/](https://www.random.org/strings/)

### 2.2 生成第一个密钥 (DATA_ENCRYPTION_KEY)

**方法 A：使用在线工具**

1. 访问 https://generate-secret.vercel.app/32
2. 点击 **"Generate"** 按钮
3. 复制生成的字符串（类似：`abcd1234efgh5678...`）
4. 保存到记事本，标记为 `DATA_ENCRYPTION_KEY`

**方法 B：使用系统终端（推荐）**

**macOS/Linux 用户：**
```bash
# 打开终端（Terminal），运行：
openssl rand -base64 32
```

**Windows 用户：**
```powershell
# 打开 PowerShell，运行：
$bytes = New-Object byte[] 32
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[Convert]::ToBase64String($bytes)
```

复制输出的字符串，保存到记事本。

### 2.3 生成第二个密钥 (JWT_SECRET)

**使用相同的方法，但长度改为 64：**

**macOS/Linux：**
```bash
openssl rand -base64 64
```

**Windows PowerShell：**
```powershell
$bytes = New-Object byte[] 64
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[Convert]::ToBase64String($bytes)
```

复制输出的字符串，保存到记事本。

### 2.4 检查你的记事本

现在你的记事本应该有两行：

```
DATA_ENCRYPTION_KEY=abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx1234yzab5678cdef==
JWT_SECRET=long_random_string_here_with_many_characters_about_88_chars_long==
```

✅ **完成标志：** 你有两个长长的随机字符串

---

## 第三步：在 Railway 上创建项目

### 3.1 登录 Railway

1. 访问 [https://railway.app/](https://railway.app/)
2. 点击右上角 **"Login"**
3. 选择 **"Login with GitHub"**（用 GitHub 账号登录）
4. 授权 Railway 访问你的 GitHub

### 3.2 创建新项目

1. 登录后，点击 **"New Project"**
2. 选择 **"Deploy from GitHub repo"**
3. 如果提示需要授权，点击 **"Configure GitHub App"**
4. 选择你刚才 Fork 的仓库：`你的用户名/nofx`
5. 点击该仓库

### 3.3 等待初始化

Railway 会自动：
- ✅ 检测到 `railway.toml` 配置文件
- ✅ 准备使用 `Dockerfile.railway` 构建

**注意：** 此时**不要点击 Deploy**，我们还需要配置环境变量！

✅ **完成标志：** 你看到了项目设置页面

---

## 第四步：配置环境变量（最重要！）

### 4.1 进入环境变量设置

1. 在项目页面，点击 **"Variables"** 标签
2. 你会看到一个空白的变量列表

### 4.2 添加必需的环境变量

依次添加以下 4 个变量（点击 **"New Variable"** 按钮）：

#### 变量 1: DATA_ENCRYPTION_KEY
```
Variable Name:  DATA_ENCRYPTION_KEY
Value:          (粘贴你在第二步生成的第一个密钥)
```

#### 变量 2: JWT_SECRET
```
Variable Name:  JWT_SECRET
Value:          (粘贴你在第二步生成的第二个密钥)
```

#### 变量 3: NOFX_ADMIN_PASSWORD
```
Variable Name:  NOFX_ADMIN_PASSWORD
Value:          (设置一个强密码，至少 12 位)
```
**示例密码：** `MySecure@Pass2024!`

#### 变量 4: NOFX_TIMEZONE
```
Variable Name:  NOFX_TIMEZONE
Value:          Asia/Shanghai
```
（如果你在其他地区，可以改为 `America/New_York`、`Europe/London` 等）

### 4.3 检查配置

确保你添加了这 4 个变量：
- ✅ DATA_ENCRYPTION_KEY
- ✅ JWT_SECRET
- ✅ NOFX_ADMIN_PASSWORD
- ✅ NOFX_TIMEZONE

✅ **完成标志：** Variables 标签下显示 4 个变量

---

## 第五步：部署应用

### 5.1 触发部署

1. Railway 会自动检测到变量更新
2. 如果没有自动部署，点击右上角 **"Deploy"** 按钮
3. 等待构建开始

### 5.2 查看构建日志

1. 点击 **"Deployments"** 标签
2. 点击最新的部署记录
3. 点击 **"View Logs"**

**预期日志（需要等待 5-10 分钟）：**

```
Building...
[Stage 1/4] Building TA-Lib...           (约 2 分钟)
[Stage 2/4] Building Frontend...         (约 2 分钟)
[Stage 3/4] Building Backend...          (约 3 分钟)
[Stage 4/4] Creating Runtime Image...    (约 1 分钟)

Deployment successful!
```

**如果看到错误：**
- 检查第四步的环境变量是否正确
- 确保密钥已正确复制（没有多余空格）
- 查看本文末尾的"常见问题"部分

### 5.3 等待启动

部署成功后，查看 **"Deploy Logs"**：

```
🚀 Starting NOFX AI Trading System on Railway...
📡 Frontend Port: 3000
🔧 Backend Port: 8080
🤖 Starting Backend in background...
✅ Backend is ready!
🌐 Starting Nginx on port 3000...
```

✅ **完成标志：** 看到 "Starting Nginx on port 3000..."

---

## 第六步：生成访问域名

### 6.1 配置网络

1. 点击 **"Settings"** 标签
2. 滚动到 **"Networking"** 部分
3. 点击 **"Generate Domain"** 按钮

### 6.2 填写端口

在弹出的对话框中：
```
Port: 3000
```
输入 `3000`，然后点击 **"Generate Domain"**

### 6.3 获取域名

Railway 会生成一个域名，类似：
```
https://nofx-production.up.railway.app
```

复制这个域名！

✅ **完成标志：** 你有了一个 `.up.railway.app` 结尾的域名

---

## 第七步：访问和配置系统

### 7.1 首次访问

1. 在浏览器中打开你的域名（例如：`https://nofx-production.up.railway.app`）
2. 你会看到登录页面

**如果看到 "Application failed to respond"：**
- 等待 1-2 分钟（应用可能还在启动）
- 刷新页面
- 检查 Deploy Logs 确认应用已启动

### 7.2 登录系统

输入你在第四步设置的管理员密码：
```
Password: (你设置的 NOFX_ADMIN_PASSWORD)
```

点击 **"登录"**

### 7.3 配置 AI 模型

登录后，点击 **"AI模型配置"**：

#### 选项 1: 配置 DeepSeek（推荐新手）

1. 访问 [https://platform.deepseek.com](https://platform.deepseek.com)
2. 注册账号并登录
3. 充值（最低 $5）
4. 创建 API Key
5. 复制 API Key（类似 `sk-xxxxxxxxxxxxx`）

回到 NOFX：
- 启用 **DeepSeek**
- 粘贴 API Key
- 点击 **"保存配置"**

#### 选项 2: 配置 Qwen

1. 访问 [https://dashscope.console.aliyun.com](https://dashscope.console.aliyun.com)
2. 注册阿里云账号
3. 开通 DashScope 服务
4. 创建 API Key
5. 复制 API Key

回到 NOFX：
- 启用 **Qwen**
- 粘贴 API Key
- 点击 **"保存配置"**

### 7.4 配置交易所

点击 **"交易所配置"**：

#### 选项 1: 配置 Binance

1. 注册 Binance 账号：[https://www.binance.com/join?ref=TINKLEVIP](https://www.binance.com/join?ref=TINKLEVIP)
2. 完成 KYC 认证
3. 开通合约账户（USD-M Futures）
4. 创建 API Key：
   - 进入 API 管理
   - 创建新 API Key
   - **启用 "Futures" 权限**
   - **绑定 IP 白名单**（可选，更安全）
5. 复制 API Key 和 Secret Key

回到 NOFX：
- 选择 **Binance**
- 粘贴 API Key 和 Secret Key
- 点击 **"保存配置"**

#### 选项 2: 配置 Hyperliquid

1. 准备一个以太坊钱包（MetaMask）
2. 导出私钥（**去掉 0x 前缀**）
3. 复制钱包地址

回到 NOFX：
- 选择 **Hyperliquid**
- 粘贴私钥（不带 0x）
- 粘贴钱包地址
- 选择主网或测试网
- 点击 **"保存配置"**

### 7.5 创建交易员

点击 **"创建交易员"**：

填写以下信息：
```
名称:         我的第一个交易员
AI 模型:      DeepSeek (或 Qwen)
交易所:       Binance (或 Hyperliquid)
初始余额:     1000 (用于计算盈亏的基准)
扫描间隔:     3 分钟
```

点击 **"创建"**

### 7.6 启动交易

1. 在交易员列表中找到刚创建的交易员
2. 点击 **"启动"** 按钮
3. 系统开始运行！

**你可以实时查看：**
- 💰 账户余额
- 📊 持仓情况
- 🤖 AI 决策日志
- 📈 盈亏曲线

---

## 🎉 完成！你成功部署了 NOFX！

现在你可以：
- ✅ 通过 Railway 域名随时访问系统
- ✅ 监控 AI 交易决策
- ✅ 查看实时盈亏
- ✅ 随时启动/停止交易员

---

## 📊 重要提示

### ⚠️ 风险警告

1. **从小资金开始**
   - 建议初期投入 100-500 USDT 测试
   - 不要投入全部资金

2. **AI 不保证盈利**
   - 加密货币市场波动极大
   - AI 决策仅供参考
   - 可能亏损本金

3. **定期监控**
   - 每天检查账户状态
   - 关注异常交易
   - 及时止损

### 🔒 安全建议

1. **保护你的密钥**
   - 不要分享给任何人
   - 定期更换密码
   - 使用强密码

2. **API 安全**
   - Binance API 只开启必要权限
   - 绑定 IP 白名单
   - 不要开启提币权限

3. **备份重要信息**
   - 记录你的域名
   - 保存环境变量
   - 定期备份配置

---

## ❓ 常见问题

### Q1: 部署失败，显示 "Build failed"

**解决方案：**
1. 检查环境变量是否正确配置
2. 确保 `DATA_ENCRYPTION_KEY` 和 `JWT_SECRET` 都已添加
3. 查看完整的构建日志找到具体错误
4. 尝试 **"Redeploy"** 重新部署

### Q2: 访问域名显示 "Application failed to respond"

**解决方案：**
1. 等待 1-2 分钟（首次启动较慢）
2. 检查 Deploy Logs 确认应用已启动
3. 确认在 Networking 中设置了端口 `3000`
4. 尝试清除浏览器缓存

### Q3: 登录后看不到数据

**解决方案：**
1. 确认已配置 AI 模型和交易所
2. 确认已创建交易员
3. 确认交易员已启动
4. 等待 3-5 分钟让 AI 做第一次决策

### Q4: Railway 说我超出免费额度了

**解决方案：**
1. Railway 提供 $5/月免费额度
2. 如需继续使用，需要绑定信用卡
3. 预计费用：$3-5/月
4. 可以随时在 Railway 中查看费用

### Q5: 如何更新代码？

**解决方案：**
1. 在你的 GitHub Fork 仓库中点击 **"Sync fork"**
2. 点击 **"Update branch"**
3. Railway 会自动检测更新并重新部署

### Q6: 如何停止服务？

**解决方案：**
1. 在 Railway 项目页面
2. 点击 **"Settings"**
3. 滚动到底部
4. 点击 **"Delete Service"**（永久删除）
或
- 在系统中停止所有交易员（暂时停止交易）

### Q7: 数据会丢失吗？

**注意：**
- Railway 容器重启后，`config.db` 会被重置
- 要持久化数据，需要使用 Railway Volumes
- 建议定期备份重要配置

**设置 Railway Volume：**
1. 进入 **"Settings"** → **"Volumes"**
2. 创建新 Volume
3. 挂载路径：`/app/config.db`

---

## 📚 下一步学习

部署成功后，你可以：

1. 📖 **学习提示词优化**
   - 阅读 [提示词编写指南](../../prompt-guide.zh-CN.md)
   - 优化 AI 决策策略

2. 🔧 **了解系统配置**
   - 查看 [FAQ 文档](../guides/faq.zh-CN.md)
   - 学习高级配置选项

3. 💬 **加入社区**
   - [Telegram 开发者社区](https://t.me/nofx_dev_community)
   - 与其他用户交流经验

4. 🐛 **遇到问题？**
   - 查看 [故障排除指南](../guides/TROUBLESHOOTING.zh-CN.md)
   - 提交 [GitHub Issue](https://github.com/NoFxAiOS/nofx/issues)

---

## 📞 获取帮助

如果遇到问题：

1. 🔍 **检查日志**
   - Railway Deploy Logs
   - 系统内的决策日志

2. 📖 **查看文档**
   - [完整部署指南](railway-deploy.zh-CN.md)
   - [故障排除](../guides/TROUBLESHOOTING.zh-CN.md)

3. 💬 **寻求帮助**
   - [Telegram 社区](https://t.me/nofx_dev_community)
   - [GitHub Issues](https://github.com/NoFxAiOS/nofx/issues)

---

**最后更新：** 2025-01-12
**适用版本：** NOFX v3.0.0+
**预计部署时间：** 10-15 分钟

🚀 **祝你部署顺利！开始你的 AI 交易之旅吧！**

---

**⚠️ 最终风险提示：**

本系统仅供学习和研究使用。加密货币交易存在极高风险，可能导致全部资金损失。请：
- ✅ 只使用你能承受损失的资金
- ✅ 充分了解风险后再开始
- ✅ 定期监控系统运行状态
- ❌ 不要投入全部资金
- ❌ 不要借钱投资
- ❌ 不要盲目相信 AI 决策

**作者不对任何交易损失负责。**
