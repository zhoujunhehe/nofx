#!/bin/bash

set -euo pipefail

# 保证从仓库根目录运行
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR"

# PostgreSQL数据恢复工具
echo "📥 PostgreSQL 数据恢复工具"
echo "========================="

# 检查参数
if [ $# -eq 0 ]; then
    echo "用法: $0 <备份文件路径> [选项]"
    echo ""
    echo "选项:"
    echo "  --force     强制恢复，删除现有数据"
    echo "  --dry-run   仅显示恢复信息，不执行实际恢复"
    echo ""
    echo "示例:"
    echo "  $0 backups/nofx_backup_20231107_143022.sql.gz"
    echo "  $0 backups/nofx_backup_20231107_143022.dump --force"
    echo ""
    echo "可用的备份文件:"
    ls -la backups/nofx_backup_* 2>/dev/null | tail -5 || echo "  (无备份文件)"
    exit 1
fi

BACKUP_FILE="$1"
FORCE_RESTORE=false
DRY_RUN=false

# 解析参数
shift
while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE_RESTORE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "❌ 未知参数: $1"
            exit 1
            ;;
    esac
done

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ 备份文件不存在: $BACKUP_FILE"
    exit 1
fi

echo "📁 备份文件: $BACKUP_FILE"
echo "📏 文件大小: $(du -h "$BACKUP_FILE" | cut -f1)"

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
        echo "❌ PostgreSQL 容器仍未就绪，恢复失败"
        exit 1
    fi
fi

echo "📋 数据库容器: $POSTGRES_CONTAINER"
echo "📋 连接参数: $POSTGRES_HOST:${POSTGRES_PORT}/$POSTGRES_DB (user: $POSTGRES_USER)"

# 设置环境变量
PG_ENV_ARGS=()
if [ -n "$POSTGRES_PASSWORD" ]; then
    PG_ENV_ARGS=(--env "PGPASSWORD=$POSTGRES_PASSWORD")
fi

# 检查当前数据库状态
echo ""
echo "🔍 检查当前数据库状态..."
CURRENT_TABLES=$(docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
    psql -t -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ' || echo "0")

echo "📊 当前表数量: $CURRENT_TABLES"

if [ "$CURRENT_TABLES" -gt 0 ] && [ "$FORCE_RESTORE" = false ]; then
    echo "⚠️  数据库包含现有数据!"
    echo "   如要继续恢复，请使用 --force 参数"
    echo "   这将删除所有现有数据!"
    exit 1
fi

if [ "$DRY_RUN" = true ]; then
    echo ""
    echo "🔍 DRY-RUN 模式 - 仅显示恢复信息"
    echo "================================"
    
    # 检查备份文件类型
    if [[ "$BACKUP_FILE" =~ \.dump$ ]]; then
        echo "📄 备份类型: PostgreSQL 二进制格式 (.dump)"
        echo "📄 恢复命令: pg_restore"
    elif [[ "$BACKUP_FILE" =~ \.sql(\.gz)?$ ]]; then
        echo "📄 备份类型: SQL 文本格式"
        echo "📄 恢复命令: psql"
    else
        echo "❓ 未知备份格式: $BACKUP_FILE"
    fi
    
    echo "📄 目标数据库: $POSTGRES_DB"
    echo "📄 将删除现有表数量: $CURRENT_TABLES"
    echo ""
    echo "✅ DRY-RUN 完成。使用不带 --dry-run 参数执行实际恢复。"
    exit 0
fi

# 确认恢复操作
if [ "$FORCE_RESTORE" = true ] && [ "$CURRENT_TABLES" -gt 0 ]; then
    echo ""
    echo "⚠️  即将删除现有数据并恢复备份!"
    read -p "确认继续? (输入 'yes' 确认): " CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
        echo "❌ 恢复已取消"
        exit 1
    fi
fi

# 执行恢复
echo ""
echo "🔄 开始恢复数据库..."

# 如果强制恢复，先清理现有数据
if [ "$FORCE_RESTORE" = true ] && [ "$CURRENT_TABLES" -gt 0 ]; then
    echo "🧹 清理现有数据..."
    docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
        psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
        "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" > /dev/null 2>&1 || {
        echo "❌ 清理现有数据失败"
        exit 1
    }
    echo "✅ 现有数据已清理"
fi

# 根据文件类型选择恢复方式
if [[ "$BACKUP_FILE" =~ \.dump$ ]]; then
    # 二进制格式恢复
    echo "📥 使用 pg_restore 恢复二进制备份..."
    if docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
        pg_restore -v -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-privileges --no-owner \
        < "$BACKUP_FILE" 2>/dev/null; then
        echo "✅ 二进制备份恢复成功"
    else
        echo "❌ 二进制备份恢复失败"
        exit 1
    fi
    
elif [[ "$BACKUP_FILE" =~ \.sql\.gz$ ]]; then
    # 压缩SQL格式恢复
    echo "📥 使用 psql 恢复压缩SQL备份..."
    if gunzip -c "$BACKUP_FILE" | docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
        psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" > /dev/null 2>&1; then
        echo "✅ 压缩SQL备份恢复成功"
    else
        echo "❌ 压缩SQL备份恢复失败"
        exit 1
    fi
    
elif [[ "$BACKUP_FILE" =~ \.sql$ ]]; then
    # SQL格式恢复
    echo "📥 使用 psql 恢复SQL备份..."
    if docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
        psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        < "$BACKUP_FILE" > /dev/null 2>&1; then
        echo "✅ SQL备份恢复成功"
    else
        echo "❌ SQL备份恢复失败"
        exit 1
    fi
else
    echo "❌ 不支持的备份文件格式: $BACKUP_FILE"
    exit 1
fi

# 验证恢复结果
echo ""
echo "🔍 验证恢复结果..."
RESTORED_TABLES=$(docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
    psql -t -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ' || echo "0")

echo "📊 恢复后表数量: $RESTORED_TABLES"

if [ "$RESTORED_TABLES" -gt 0 ]; then
    # 显示表统计
    echo ""
    echo "📊 恢复数据统计:"
    echo "================"
    docker exec -i "${PG_ENV_ARGS[@]}" "$POSTGRES_CONTAINER" \
        psql -v ON_ERROR_STOP=1 --pset pager=off -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
        "SELECT 
            relname AS \"表名\", 
            n_live_tup AS \"记录数\"
         FROM pg_stat_user_tables 
         WHERE n_live_tup > 0 
         ORDER BY relname;" 2>/dev/null || echo "⚠️  无法获取表统计信息"
    
    echo ""
    echo "✅ 数据库恢复完成!"
    echo "📊 共恢复 $RESTORED_TABLES 个表"
    echo "⏰ 恢复时间: $(date)"
else
    echo "❌ 恢复失败：未发现任何表"
    exit 1
fi