#!/bin/bash

# NOFX Railway 部署密钥生成脚本
# 一键生成 Railway 部署所需的两个密钥

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${PURPLE}╔════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                          NOFX AI 交易系统                             ║${NC}"
echo -e "${PURPLE}║                     Railway 部署密钥生成工具                          ║${NC}"
echo -e "${PURPLE}║                                                                        ║${NC}"
echo -e "${PURPLE}║  用途: 生成 Railway 部署所需的加密密钥                                ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════════════════════════════════════╝${NC}"
echo

# 检查 OpenSSL
echo -e "${CYAN}🔍 检查系统依赖...${NC}"

if ! command -v openssl &> /dev/null; then
    echo -e "${RED}❌ 错误: 系统中未安装 OpenSSL${NC}"
    echo
    echo -e "${YELLOW}请安装 OpenSSL:${NC}"
    echo -e "  macOS:          ${GREEN}brew install openssl${NC}"
    echo -e "  Ubuntu/Debian:  ${GREEN}sudo apt-get install openssl${NC}"
    echo -e "  CentOS/RHEL:    ${GREEN}sudo yum install openssl${NC}"
    echo
    exit 1
fi

echo -e "${GREEN}✓ OpenSSL: $(openssl version)${NC}"
echo

# 生成密钥
echo -e "${BLUE}🔐 正在生成密钥...${NC}"
echo

# 生成 DATA_ENCRYPTION_KEY (AES-256, 32 bytes -> base64)
echo -e "${CYAN}[1/2] 生成数据加密密钥 (DATA_ENCRYPTION_KEY)...${NC}"
DATA_KEY=$(openssl rand -base64 32)
echo -e "${GREEN}✓ 生成完成${NC}"
echo

# 生成 JWT_SECRET (64 bytes -> base64)
echo -e "${CYAN}[2/2] 生成 JWT 认证密钥 (JWT_SECRET)...${NC}"
JWT_KEY=$(openssl rand -base64 64)
echo -e "${GREEN}✓ 生成完成${NC}"
echo

# 保存到 .env 文件
echo -e "${BLUE}💾 保存密钥到 .env 文件...${NC}"

if [ -f ".env" ]; then
    echo -e "${YELLOW}⚠️  .env 文件已存在${NC}"
    read -p "是否覆盖现有密钥？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}ℹ️  保持现有密钥，退出${NC}"
        echo
        echo -e "${YELLOW}生成的密钥（未保存）：${NC}"
        echo -e "${GREEN}DATA_ENCRYPTION_KEY=${DATA_KEY}${NC}"
        echo -e "${GREEN}JWT_SECRET=\"${JWT_KEY}\"${NC}"
        exit 0
    fi

    # 备份现有 .env
    BACKUP_FILE=".env.backup.$(date +%Y%m%d_%H%M%S)"
    cp .env "$BACKUP_FILE"
    echo -e "${GREEN}✓ 已备份到: $BACKUP_FILE${NC}"
fi

# 更新或创建 .env 文件
if [ -f ".env" ]; then
    # 更新现有密钥
    if grep -q "^DATA_ENCRYPTION_KEY=" .env; then
        # macOS 和 Linux 的 sed 语法不同
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/^DATA_ENCRYPTION_KEY=.*/DATA_ENCRYPTION_KEY=$DATA_KEY/" .env
        else
            sed -i "s/^DATA_ENCRYPTION_KEY=.*/DATA_ENCRYPTION_KEY=$DATA_KEY/" .env
        fi
    else
        echo "DATA_ENCRYPTION_KEY=$DATA_KEY" >> .env
    fi

    if grep -q "^JWT_SECRET=" .env; then
        # 使用替代分隔符避免 / 字符冲突
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|^JWT_SECRET=.*|JWT_SECRET=\"$JWT_KEY\"|" .env
        else
            sed -i "s|^JWT_SECRET=.*|JWT_SECRET=\"$JWT_KEY\"|" .env
        fi
    else
        printf "JWT_SECRET=\"%s\"\n" "$JWT_KEY" >> .env
    fi
else
    # 创建新 .env 文件
    cat > .env << EOF
# NOFX Railway 部署环境变量
# 由 scripts/generate_railway_keys.sh 自动生成
# 生成时间: $(date)

# 数据加密密钥 (AES-256)
DATA_ENCRYPTION_KEY=$DATA_KEY

# JWT 认证密钥
JWT_SECRET="$JWT_KEY"

# 可选配置（根据需要取消注释）
# NOFX_ADMIN_PASSWORD=your_secure_password
# NOFX_BACKEND_PORT=8080
# NOFX_TIMEZONE=Asia/Shanghai
EOF
fi

# 设置权限
chmod 600 .env
echo -e "${GREEN}✓ 密钥已保存到 .env (权限: 600)${NC}"
echo

# 显示结果
echo -e "${PURPLE}╔════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                            密钥生成完成                               ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════════════════════════════════════╝${NC}"
echo

echo -e "${GREEN}✅ 以下密钥已成功生成并保存到 .env 文件：${NC}"
echo
echo -e "${CYAN}DATA_ENCRYPTION_KEY:${NC}"
echo -e "${GREEN}$DATA_KEY${NC}"
echo
echo -e "${CYAN}JWT_SECRET:${NC}"
echo -e "${GREEN}$JWT_KEY${NC}"
echo

# Railway 部署指南
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📋 下一步：在 Railway 中配置环境变量${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo -e "${CYAN}1. 登录 Railway Dashboard:${NC}"
echo -e "   https://railway.app/dashboard"
echo

echo -e "${CYAN}2. 选择你的项目，进入 Variables 标签${NC}"
echo

echo -e "${CYAN}3. 添加以下环境变量：${NC}"
echo

echo -e "${GREEN}   Variable 1:${NC}"
echo -e "   Name:  ${YELLOW}DATA_ENCRYPTION_KEY${NC}"
echo -e "   Value: ${GREEN}$DATA_KEY${NC}"
echo

echo -e "${GREEN}   Variable 2:${NC}"
echo -e "   Name:  ${YELLOW}JWT_SECRET${NC}"
echo -e "   Value: ${GREEN}$JWT_KEY${NC}"
echo

echo -e "${GREEN}   Variable 3 (可选):${NC}"
echo -e "   Name:  ${YELLOW}NOFX_ADMIN_PASSWORD${NC}"
echo -e "   Value: ${GREEN}your_secure_password${NC} ${CYAN}(设置一个强密码)${NC}"
echo

echo -e "${GREEN}   Variable 4 (可选):${NC}"
echo -e "   Name:  ${YELLOW}NOFX_TIMEZONE${NC}"
echo -e "   Value: ${GREEN}Asia/Shanghai${NC} ${CYAN}(或你的时区)${NC}"
echo

echo -e "${CYAN}4. 点击 Deploy 按钮部署应用${NC}"
echo
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# 安全提醒
echo -e "${RED}🔒 安全提醒：${NC}"
echo -e "  • ${YELLOW}.env 文件已设置为只读权限 (600)${NC}"
echo -e "  • ${YELLOW}请勿将 .env 文件提交到 Git${NC}"
echo -e "  • ${YELLOW}请妥善保管生成的密钥${NC}"
echo -e "  • ${YELLOW}定期更换密钥以提高安全性${NC}"
echo

# 快速复制命令
echo -e "${BLUE}💡 快速查看密钥：${NC}"
echo -e "   ${GREEN}cat .env${NC}"
echo

echo -e "${GREEN}✅ 密钥生成完成！现在可以在 Railway 上部署了${NC}"
echo
