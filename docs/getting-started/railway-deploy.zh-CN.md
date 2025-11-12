# 🚀 Railway 部署指南

本指南将帮助你在 Railway 平台上部署自己的 NOFX AI 交易系统。

## 📋 目录

- [准备工作](#准备工作)
- [步骤 1: 创建 Railway 项目](#步骤-1-创建-railway-项目)
- [步骤 2: 配置环境变量](#步骤-2-配置-环境变量)
- [步骤 3: 部署应用](#步骤-3-部署应用)
- [步骤 4: 初始化配置](#步骤-4-初始化配置)
- [常见问题](#常见问题)
- [安全建议](#安全建议)

---

## 准备工作

在开始之前,请确保你已经:

1. ✅ 注册 [Railway 账号](https://railway.app/)
2. ✅ Fork 本项目到你的 GitHub 账号
3. ✅ 准备好以下 API 密钥:
   - DeepSeek API Key (或 Qwen API Key)
   - Binance API Key + Secret (或 Hyperliquid/Aster 配置)

---

## 步骤 1: 创建 Railway 项目

### 1.1 连接 GitHub 仓库

1. 登录 [Railway Dashboard](https://railway.app/dashboard)
2. 点击 **"New Project"**
3. 选择 **"Deploy from GitHub repo"**
4. 授权 Railway 访问你的 GitHub 账号
5. 选择你 fork 的 `nofx` 仓库

### 1.2 选择服务类型

Railway 会自动检测到项目中的 `railway.toml` 和 `Dockerfile.railway`：
- ✅ **Docker** 部署方式（使用 `Dockerfile.railway`）
- ✅ 前端 + 后端合并在一个容器中
- ✅ Nginx 作为反向代理，自动处理静态文件和 API 请求

**注意：** 项目包含两个 Dockerfile：
- `docker/Dockerfile.backend` + `docker/Dockerfile.frontend` - 用于本地 docker-compose 部署
- `Dockerfile.railway` - 专门为 Railway 单容器部署优化

---

## 步骤 2: 配置环境变量

这是最关键的一步! Railway 需要以下环境变量才能正常运行。

### 2.1 必需的加密密钥

在 Railway 项目设置中,添加以下环境变量:

#### 方法 A: 使用自动生成脚本 (推荐)

在本地终端运行以下命令生成密钥:

```bash
# 克隆你 fork 的仓库
git clone https://github.com/YOUR_USERNAME/nofx.git
cd nofx

# 运行加密设置脚本
chmod +x scripts/setup_encryption.sh
./scripts/setup_encryption.sh
```

脚本会自动生成并保存到 `.env` 文件。然后复制以下值到 Railway:

```bash
# 查看生成的密钥
cat .env
```

你会看到类似这样的内容:

```env
DATA_ENCRYPTION_KEY=abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx1234yzab5678cdef==
JWT_SECRET="long_random_string_here_for_jwt_authentication_security"
```

#### 方法 B: 手动生成密钥

如果你不想运行脚本,也可以手动生成:

**在 macOS/Linux 终端:**

```bash
# 生成 DATA_ENCRYPTION_KEY (AES-256)
openssl rand -base64 32

# 生成 JWT_SECRET
openssl rand -base64 64
```

**在 Windows PowerShell:**

```powershell
# 生成 DATA_ENCRYPTION_KEY
$bytes = New-Object byte[] 32
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[Convert]::ToBase64String($bytes)

# 生成 JWT_SECRET
$bytes = New-Object byte[] 64
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[Convert]::ToBase64String($bytes)
```

### 2.2 在 Railway 中添加环境变量

1. 进入你的 Railway 项目
2. 点击项目名称 → **"Variables"** 标签
3. 添加以下变量:

| 变量名 | 值 | 说明 | 是否必需 |
|--------|-----|------|---------|
| `DATA_ENCRYPTION_KEY` | `(步骤2.1生成的密钥)` | AES-256 数据加密密钥 | ✅ **必需** |
| `JWT_SECRET` | `(步骤2.1生成的密钥)` | JWT 认证密钥 | ✅ **必需** |
| `NOFX_BACKEND_PORT` | `8080` | 后端 API 端口 | ❌ 可选 (默认 8080) |
| `NOFX_TIMEZONE` | `Asia/Shanghai` | 时区设置 | ❌ 可选 (默认 UTC) |

**⚠️ 重要提示:**
- `DATA_ENCRYPTION_KEY` 和 `JWT_SECRET` 是必需的,否则应用无法启动
- 密钥必须是 Base64 编码的随机字符串
- 请妥善保管这些密钥,丢失后无法恢复加密数据

### 2.3 管理员模式配置 (可选)

如果你想启用管理员模式 (单用户模式):

| 变量名 | 值 | 说明 |
|--------|-----|------|
| `NOFX_ADMIN_PASSWORD` | `your_secure_password` | 管理员登录密码 |

**关于管理员模式:**
- 启用后,所有 API 端点需要 JWT 认证
- 适合个人自用或单租户部署
- 登录后会获得 24 小时有效期的 token

---

## 步骤 3: 部署应用

### 3.1 触发部署

1. 环境变量配置完成后,Railway 会自动触发部署
2. 如果没有自动部署,点击 **"Deploy"** 按钮手动触发
3. 等待构建过程 (首次构建可能需要 5-10 分钟)

### 3.2 查看部署状态

在 Railway Dashboard 中:
- ✅ **Build Logs**: 查看 Docker 镜像构建日志
- ✅ **Deploy Logs**: 查看应用运行日志
- ✅ **Metrics**: 查看 CPU、内存使用情况

**成功部署的日志应该显示:**

```
╔════════════════════════════════════════════════════════════╗
║    🤖 AI多模型交易系统 - 支持 DeepSeek & Qwen            ║
╚════════════════════════════════════════════════════════════╝
✓ 已加载 4 个系统提示词模板
🔐 初始化加密服务...
✅ 加密服务初始化成功
📋 初始化配置数据库: config.db
🌐 API服务器启动在 http://localhost:8080
```

### 3.3 获取应用 URL

1. 在 Railway 项目设置中,点击 **"Settings"** → **"Networking"**
2. 点击 **"Generate Domain"** 生成公开访问域名
3. 你会得到类似这样的 URL: `https://your-app-name.up.railway.app`

---

## 步骤 4: 初始化配置

### 4.1 访问 Web 界面

在浏览器中打开你的 Railway 应用 URL:

```
https://your-app-name.up.railway.app
```

**首次访问:**
- 如果启用了管理员模式,你会看到登录页面
- 如果未启用管理员模式,直接进入主界面

### 4.2 配置 AI 模型

1. 点击 **"AI模型配置"** 按钮
2. 启用 DeepSeek 或 Qwen (或两者都启用)
3. 输入你的 API Key
4. 点击 **"保存配置"**

**示例配置:**

| AI 模型 | API Key | API URL (可选) |
|---------|---------|----------------|
| DeepSeek | `sk-xxxxxxxxxxxxx` | `https://api.deepseek.com` |
| Qwen | `sk-xxxxxxxxxxxxx` | `https://dashscope.aliyuncs.com/compatible-mode/v1` |

### 4.3 配置交易所

1. 点击 **"交易所配置"** 按钮
2. 选择交易所类型 (Binance/Hyperliquid/Aster)
3. 输入 API 凭证

**Binance 配置:**

```
API Key: your_binance_api_key
Secret Key: your_binance_secret_key
```

**Hyperliquid 配置:**

```
Private Key: your_ethereum_private_key (不带 0x 前缀)
Wallet Address: your_ethereum_wallet_address
Testnet: false (使用主网)
```

**Aster 配置:**

```
User Address: your_main_wallet_address
Signer Address: your_api_wallet_address
Private Key: your_api_wallet_private_key (不带 0x 前缀)
```

### 4.4 创建交易员

1. 点击 **"创建交易员"** 按钮
2. 填写以下信息:
   - **名称**: 例如 "DeepSeek Trader"
   - **AI 模型**: 选择已配置的模型
   - **交易所**: 选择已配置的交易所
   - **初始余额**: 用于盈亏计算的基准金额
   - **扫描间隔**: 建议 3-5 分钟
3. 点击 **"创建"**

### 4.5 启动交易

1. 在交易员列表中找到你创建的交易员
2. 点击 **"启动"** 按钮
3. 系统开始运行,你可以实时监控:
   - 账户余额
   - 持仓情况
   - AI 决策日志
   - 盈亏曲线

---

## 常见问题

### ❌ 错误: `DATA_ENCRYPTION_KEY not set`

**原因:** 缺少必需的环境变量

**解决方案:**
1. 返回 Railway Dashboard
2. 检查 **Variables** 标签中是否存在 `DATA_ENCRYPTION_KEY`
3. 如果不存在,按照 [步骤 2.1](#21-必需的加密密钥) 生成并添加
4. 重新部署应用

### ❌ 错误: `port already in use`

**原因:** 端口冲突 (通常不会在 Railway 上发生)

**解决方案:**
- 检查 `NOFX_BACKEND_PORT` 环境变量
- 确保没有重复的服务监听同一端口

### ❌ 错误: `database locked`

**原因:** SQLite 数据库并发访问问题

**解决方案:**
- Railway 的临时文件系统可能导致此问题
- 考虑使用 [Railway Volumes](https://docs.railway.app/reference/volumes) 持久化 `config.db`

### ❌ 应用无法访问

**原因:** 未生成公开域名或端口配置错误

**解决方案:**
1. 检查 **Settings** → **Networking** 是否生成了域名
2. 确保 Dockerfile 中 `EXPOSE 8080` 与 `NOFX_BACKEND_PORT` 一致
3. 检查防火墙或 Railway 的网络设置

### 💾 数据持久化问题

**问题:** Railway 重启后数据丢失

**原因:** Railway 的容器文件系统是临时的

**解决方案:**

1. **使用 Railway Volumes (推荐)**

在 Railway Dashboard 中:
```bash
Settings → Volumes → New Volume
Mount Path: /app/config.db
```

2. **或使用外部数据库**

未来版本可能支持 PostgreSQL/MySQL,届时可以连接 Railway 提供的数据库服务。

---

## 安全建议

### 🔒 密钥管理

1. **永远不要提交密钥到 Git**
   - `.env` 已在 `.gitignore` 中
   - 确保不要 commit `secrets/` 目录

2. **定期轮换密钥**
   - 建议每 3-6 个月更换一次 `JWT_SECRET`
   - 更换 `DATA_ENCRYPTION_KEY` 会导致现有加密数据无法解密

3. **使用强密码**
   - `NOFX_ADMIN_PASSWORD` 应至少 12 位,包含大小写、数字、特殊字符
   - 不要使用常见密码

### 🌐 网络安全

1. **启用 HTTPS**
   - Railway 自动提供 HTTPS,无需额外配置

2. **限制访问**
   - 如果只有自己使用,考虑使用 Railway 的 [Private Networking](https://docs.railway.app/reference/private-networking)
   - 或配置 IP 白名单 (需要额外配置)

3. **监控日志**
   - 定期检查 Deploy Logs 中的异常活动
   - 关注 API 调用频率和错误日志

### 💰 API 密钥安全

1. **Binance API Key 设置**
   - ✅ 只启用必需的权限 (Futures Trading)
   - ✅ 绑定 IP 白名单 (Railway 的出口 IP)
   - ❌ 不要启用提币权限

2. **Hyperliquid Private Key**
   - 使用专用交易钱包,不要用主钱包
   - 只存放必要的资金

3. **AI API Key**
   - 设置消费限额
   - 定期检查使用量

### 📊 资金安全

1. **测试阶段**
   - 初期只投入小额资金 (100-500 USDT)
   - 使用测试网 (Hyperliquid Testnet) 验证策略

2. **风险控制**
   - 设置合理的杠杆 (建议 ≤5x)
   - 监控最大回撤
   - 设置止损规则

3. **定期检查**
   - 每天检查账户余额和持仓
   - 关注异常交易
   - 及时停止表现不佳的策略

---

## 🎯 下一步

部署成功后,你可以:

1. 📖 阅读 [提示词编写指南](../../prompt-guide.zh-CN.md) 优化 AI 策略
2. 🔧 查看 [FAQ 文档](../guides/faq.zh-CN.md) 了解更多配置选项
3. 🐛 遇到问题? 查看 [故障排除指南](../guides/TROUBLESHOOTING.zh-CN.md)
4. 💬 加入 [Telegram 开发者社区](https://t.me/nofx_dev_community) 交流经验

---

## 📞 获取帮助

如果遇到问题:

1. 🔍 检查 Railway Deploy Logs
2. 📖 查看 [故障排除指南](../guides/TROUBLESHOOTING.zh-CN.md)
3. 🐛 提交 [GitHub Issue](https://github.com/tinkle-community/nofx/issues)
4. 💬 在 [Telegram 社区](https://t.me/nofx_dev_community) 求助

---

**⚠️ 风险提示:**

- 本系统用于学习和研究目的
- AI 自动交易存在巨大风险
- 加密货币市场波动极大
- 强烈建议只使用你可以承受损失的资金进行测试
- 作者不对任何交易损失负责

---

**最后更新:** 2025-01-12
**Railway 平台版本:** Railway V2
**NOFX 版本:** v3.0.0+

🚀 祝你部署顺利!
