# ✅ Railway 部署检查清单

快速检查你的 Railway 部署配置是否正确。

## 📋 部署前检查

### 1. 仓库文件确认

- [ ] `Dockerfile.railway` 存在
- [ ] `railway.toml` 存在
- [ ] `nginx/nginx-railway.conf` 存在
- [ ] `.env` 文件包含必要的密钥（不要提交到 Git）

### 2. 环境变量准备

从你的本地 `.env` 文件复制以下值：

```bash
# 查看你的密钥
cat .env
```

你需要在 Railway 中配置：

| 变量名 | 必需 | 说明 |
|--------|------|------|
| `DATA_ENCRYPTION_KEY` | ✅ | 数据加密密钥（Base64） |
| `JWT_SECRET` | ✅ | JWT 认证密钥（Base64） |
| `NOFX_ADMIN_PASSWORD` | ❌ | 管理员密码（可选） |
| `NOFX_TIMEZONE` | ❌ | 时区（默认 Asia/Shanghai） |

---

## 🚀 Railway 部署步骤

### 步骤 1: 创建项目

1. 访问 [Railway Dashboard](https://railway.app/dashboard)
2. 点击 **"New Project"**
3. 选择 **"Deploy from GitHub repo"**
4. 选择你的 `nofx` 仓库

### 步骤 2: 配置环境变量

在 Railway 项目中：

1. 进入 **Variables** 标签
2. 添加环境变量（从上面的表格复制）
3. 点击 **"Save"**

**示例：**
```
DATA_ENCRYPTION_KEY=odO5tWson72hprAaPKgpKNZG+spZSG0XMxEYdDN38SU=
JWT_SECRET=sqGUIkbP6gYtjHT4x3nVflUrVIId6TMhG26GkpZNAvq8tehM+MgXZmEbJQ/bz3+en5cAkOP+tqm/ht2GDBBpYQ==
NOFX_ADMIN_PASSWORD=your_secure_password
NOFX_TIMEZONE=Asia/Shanghai
```

### 步骤 3: 触发部署

1. 保存环境变量后，Railway 会自动部署
2. 查看 **Deploy Logs** 确认构建进度

**预期日志：**
```
Building with Dockerfile.railway...
[Stage 1/4] Building TA-Lib...
[Stage 2/4] Building Frontend...
[Stage 3/4] Building Backend...
[Stage 4/4] Creating Runtime Image...
Deployment successful!
```

### 步骤 4: 生成域名

1. 进入 **Settings** → **Networking**
2. 点击 **"Generate Domain"**
3. 获得类似：`https://your-app.up.railway.app`

### 步骤 5: 验证部署

打开浏览器访问你的域名：

- ✅ 首页加载正常
- ✅ API 健康检查：`https://your-app.up.railway.app/api/health`
- ✅ 前端静态资源加载正常

---

## 🔍 故障排查

### ❌ 构建失败

**检查项：**
1. Deploy Logs 中的错误信息
2. 确认 `Dockerfile.railway` 存在
3. 确认 `railway.toml` 配置正确

### ❌ 启动失败：`DATA_ENCRYPTION_KEY not set`

**解决方案：**
1. 检查 **Variables** 标签
2. 确认 `DATA_ENCRYPTION_KEY` 已添加
3. 重新部署

### ❌ 前端无法访问

**检查项：**
1. 访问 `/api/health` 是否返回 `{"status":"ok"}`
2. 查看 Deploy Logs 中 Nginx 是否启动成功
3. 确认 Railway 生成的域名正确

### ❌ API 请求失败 (404)

**可能原因：**
- Nginx 反向代理配置错误
- 后端未正常启动

**解决方案：**
1. 查看 Deploy Logs
2. 确认后端日志：`🤖 Starting Backend on port 8080...`
3. 测试健康检查：`curl https://your-app.up.railway.app/api/health`

### ❌ 数据持久化问题

**问题：** Railway 重启后数据丢失

**解决方案：**

Railway 的容器文件系统是临时的。要持久化数据：

1. **使用 Railway Volumes**
   - 进入 **Settings** → **Volumes**
   - 创建新 Volume
   - 挂载路径：`/app/config.db`

2. **或手动备份数据库**
   - 定期下载 `config.db` 文件
   - 使用外部存储服务（如 S3）

---

## 📊 监控和维护

### 查看日志

在 Railway Dashboard：
- **Deploy Logs**: 实时查看应用日志
- **Metrics**: 监控 CPU、内存使用

### 重新部署

当你更新代码后：
1. Push 到 GitHub
2. Railway 自动检测并重新部署
3. 或手动点击 **"Redeploy"**

### 环境变量更新

修改环境变量后：
1. 点击 **"Redeploy"**
2. 或等待自动重启

---

## 🎯 部署成功标志

- ✅ Deploy Logs 显示：`🤖 Starting Backend on port 8080...`
- ✅ 访问域名能看到前端界面
- ✅ `/api/health` 返回：`{"status":"ok"}`
- ✅ 可以正常登录和配置交易员

---

## 📞 获取帮助

- 📖 完整文档：[docs/getting-started/railway-deploy.zh-CN.md](./docs/getting-started/railway-deploy.zh-CN.md)
- 🐛 提交问题：[GitHub Issues](https://github.com/tinkle-community/nofx/issues)
- 💬 社区讨论：[Telegram 开发者社区](https://t.me/nofx_dev_community)

---

**最后更新：** 2025-01-12
