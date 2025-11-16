#!/bin/bash

set -euo pipefail

# 保证从仓库根目录运行
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR"

# PostgreSQL数据备份工具
echo "💾 PostgreSQL 数据备份工具"
echo "========================="

# 检测Docker Compose命令
DOCKER_COMPOSE_CMD=""
if command -v "docker-compose" &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif command -v "docker" &> /dev/null && docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    echo "❌ 错误：找不到 docker-compose 或 docker compose 命令"
    exit 1
fi

# 加载数据库配置
ENV_FILE=".env"
if [ -f "$ENV_FILE" ]; then
    echo "📁 加载 .env 配置..."
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
else
    echo "⚠️  未找到 .env 文件，使用默认数据库配置"
fi

POSTGRES_HOST=${POSTGRES_HOST:-postgres}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-nofx}
POSTGRES_USER=${POSTGRES_USER:-nofx}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-}
POSTGRES_SERVICE=${POSTGRES_SERVICE:-postgres}
POSTGRES_CONTAINER_NAME=${POSTGRES_CONTAINER_NAME:-nofx-postgres}

# 备份目录和文件名
BACKUP_DIR="${ROOT_DIR}/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILENAME="nofx_backup_${TIMESTAMP}.sql"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_FILENAME}"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 获取 PostgreSQL 容器 ID
POSTGRES_CONTAINER=$($DOCKER_COMPOSE_CMD ps -q "$POSTGRES_SERVICE" 2>/dev/null || true)
if [ -z "$POSTGRES_CONTAINER" ]; then
    POSTGRES_CONTAINER=$(docker ps -q --filter "name=$POSTGRES_CONTAINER_NAME" | head -n 1)
fi

if [ -z "$POSTGRES_CONTAINER" ]; then
    echo "❌ 找不到 PostgreSQL 容器 (${POSTGRES_SERVICE}/${POSTGRES_CONTAINER_NAME})"
    echo "💡 请确认数据库服务已启动"
    exit 1
fi

# 检查容器是否健康
if ! docker exec "$POSTGRES_CONTAINER" pg_isready -U "$POSTGRES_USER" > /dev/null 2>&1; then
    echo "❌ PostgreSQL 容器未就绪，等待启动..."
    sleep 5
    if ! docker exec "$POSTGRES_CONTAINER" pg_isready -U "$POSTGRES_USER" > /dev/null 2>&1; then
        echo "❌ PostgreSQL 容器仍未就绪，备份失败"
        exit 1
    fi
fi

echo "📋 数据库容器: $POSTGRES_CONTAINER"
echo "📋 连接参数: $POSTGRES_HOST:${POSTGRES_PORT}/$POSTGRES_DB (user: $POSTGRES_USER)"
echo "📋 备份文件: $BACKUP_PATH"

# 设置环境变量
PG_ENV_ARGS=()
if [ -n "$POSTGRES_PASSWORD" ]; then
    PG_ENV_ARGS=(--env "PGPASSWORD=$POSTGRES_PASSWORD")
fi

# 执行数据库备份
echo "🔄 开始备份数据库..."

# 使用 pg_dump 进行完整备份
if docker exec "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
    pg_dump -v --no-password -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    --format=custom --compress=9 --no-privileges --no-owner \
    > "${BACKUP_PATH%.sql}.dump" 2>/dev/null; then
    
    echo "✅ 二进制备份完成: ${BACKUP_PATH%.sql}.dump"
else
    echo "⚠️  二进制备份失败，尝试SQL格式备份..."
fi

# SQL格式备份（兼容性更好）
if docker exec "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
    pg_dump -v --no-password -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    --format=plain --no-privileges --no-owner \
    > "$BACKUP_PATH" 2>/dev/null; then
    
    echo "✅ SQL备份完成: $BACKUP_PATH"
    
    # 压缩SQL文件
    if command -v gzip &> /dev/null; then
        gzip "$BACKUP_PATH"
        echo "🗜️  备份文件已压缩: ${BACKUP_PATH}.gz"
        BACKUP_PATH="${BACKUP_PATH}.gz"
    fi
else
    echo "❌ SQL备份失败"
    exit 1
fi

# 获取备份文件大小
BACKUP_SIZE=$(du -h "$BACKUP_PATH" 2>/dev/null | cut -f1 || echo "未知")

# 显示备份统计信息
echo ""
echo "📊 备份统计信息:"
echo "=================="

# 统计各表记录数
docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
    psql -v ON_ERROR_STOP=1 --pset pager=off -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
    "SELECT 
        schemaname as \"模式\",
        relname AS \"表名\", 
        n_live_tup AS \"记录数\",
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) AS \"大小\"
     FROM pg_stat_user_tables 
     WHERE n_live_tup > 0 
     ORDER BY n_live_tup DESC;" 2>/dev/null || echo "⚠️  无法获取表统计信息"

echo ""
echo "✅ 备份完成!"
echo "📁 备份文件: $BACKUP_PATH"
echo "📏 文件大小: $BACKUP_SIZE"
echo "⏰ 备份时间: $(date)"

# 备份重要配置文件
echo ""
echo "🔧 备份配置文件..."
CONFIG_BACKUP_DIR="${BACKUP_DIR}/config_${TIMESTAMP}"
mkdir -p "$CONFIG_BACKUP_DIR"

# 备份环境配置（排除敏感信息）
if [ -f ".env" ]; then
    # 创建不包含敏感信息的环境配置备份
    grep -v -E "(PASSWORD|SECRET|KEY|TOKEN)" .env > "${CONFIG_BACKUP_DIR}/env_safe.txt" || true
    echo "✅ 环境配置已备份 (已排除敏感信息)"
fi

# 备份应用配置
if [ -f "config.json" ]; then
    cp "config.json" "${CONFIG_BACKUP_DIR}/"
    echo "✅ 应用配置已备份"
fi

# 备份Docker Compose配置
if [ -f "docker-compose.yml" ]; then
    cp "docker-compose.yml" "${CONFIG_BACKUP_DIR}/"
    echo "✅ Docker Compose配置已备份"
fi

echo ""
echo "📦 配置文件备份目录: $CONFIG_BACKUP_DIR"

# 清理旧备份（保留最近7天）
echo ""
echo "🧹 清理旧备份文件..."
find "$BACKUP_DIR" -name "nofx_backup_*.sql*" -mtime +7 -type f -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "nofx_backup_*.dump" -mtime +7 -type f -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "config_*" -mtime +7 -type d -exec rm -rf {} + 2>/dev/null || true

echo "✅ 备份流程完成!"
echo ""
echo "📝 恢复方式:"
echo "============"
echo "1. SQL格式恢复:"
echo "   gunzip -c $BACKUP_PATH | docker exec -i \$POSTGRES_CONTAINER psql -U $POSTGRES_USER -d $POSTGRES_DB"
echo ""
echo "2. 二进制格式恢复 (如果存在):"
echo "   docker exec -i \$POSTGRES_CONTAINER pg_restore -U $POSTGRES_USER -d $POSTGRES_DB < ${BACKUP_PATH%.sql.gz}.dump"
echo ""
echo "⚠️  恢复前请确保目标数据库为空，或先删除现有数据"