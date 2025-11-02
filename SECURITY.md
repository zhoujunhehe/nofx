# 前端防爬虫配置

本项目前端已配置防止机器人和爬虫扫描的措施：

## 1. robots.txt (已配置)

位置：`web/public/robots.txt`

功能：
- 阻止所有爬虫访问整站
- 特别阻止 AI 爬虫（GPTBot, ChatGPT-User, Claude-Web, Bytespider 等）
- 设置 Crawl-delay 限制

**自定义配置：**
如需允许部分路径被搜索引擎收录，可以修改 `robots.txt`：

```txt
# 允许搜索引擎，但阻止特定路径
User-agent: *
Disallow: /api/
Disallow: /admin/
Allow: /
```

## 2. HTTP 安全响应头 (已配置)

位置：`web/vite.config.ts`

已添加的安全响应头：
- `X-Robots-Tag: noindex, nofollow` - 防止页面被搜索引擎索引
- `X-Content-Type-Options: nosniff` - 防止 MIME 类型嗅探
- `X-Frame-Options: DENY` - 防止点击劫持
- `X-XSS-Protection: 1; mode=block` - XSS 过滤
- `Referrer-Policy: strict-origin-when-cross-origin` - 控制引用信息泄露

## 3. 生产环境部署建议

### 使用 CDN（如 Cloudflare）
可以额外启用：
- Bot Fight Mode（机器人对抗模式）
- Rate Limiting Rules（速率限制）
- WAF (Web Application Firewall)

### Nginx 配置示例
如果使用 Nginx 部署，可以添加：
```nginx
# 限制请求频率
limit_req_zone $binary_remote_addr zone=one:10m rate=10r/s;

server {
    location / {
        limit_req zone=one burst=20;
        # 其他配置...
    }
}
```

## 注意事项

⚠️ **重要提醒：**
1. `robots.txt` 只对遵守规则的爬虫有效，恶意爬虫会忽略
2. HTTP 响应头提供额外保护层，但不能完全阻止爬虫
3. 前端防护有限，真正有效的防护需要结合后端和 CDN

## 测试验证

部署后可以测试配置是否生效：

```bash
# 测试 robots.txt
curl https://your-domain.com/robots.txt

# 测试 HTTP 响应头
curl -I https://your-domain.com

# 应该能看到 X-Robots-Tag 和其他安全头
```
