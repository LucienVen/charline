-- ============================================
-- CharLine IM 数据库初始化脚本
-- Version: 001
-- Description: 创建邀请码、用户、消息表
-- ============================================

-- 邀请码表（服务端）
CREATE TABLE IF NOT EXISTS invite_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 主键 ID
    code TEXT UNIQUE NOT NULL,             -- 邀请码（格式: INV-XXXXXXXX）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    used_at DATETIME,                      -- 使用时间（NULL 表示未使用）
    username TEXT,                         -- 使用者用户名（激活时填充）
    CHECK (code != '')                     -- 确保邀请码非空
);

-- 邀请码索引：按码查询
CREATE INDEX IF NOT EXISTS idx_invite_code ON invite_codes(code);

-- 邀请码索引：按使用时间查询（查找未使用的）
CREATE INDEX IF NOT EXISTS idx_used_at ON invite_codes(used_at);

-- 用户表（客户端/服务端共用）
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 主键 ID
    username TEXT UNIQUE NOT NULL,         -- 用户名（唯一）
    token_version INTEGER DEFAULT 1,       -- Token 版本号（用于作废旧 Token）
    server_url TEXT,                       -- 服务器地址（客户端存储）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- 注册时间
    last_login DATETIME,                   -- 最后登录时间
    CHECK (username != '')                 -- 确保用户名非空
);

-- 用户索引：按用户名查询
CREATE INDEX IF NOT EXISTS idx_username ON users(username);

-- 聊天消息表（客户端，Phase 3）
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 主键 ID
    from_user TEXT NOT NULL,               -- 发送者用户名
    to_user TEXT NOT NULL,                 -- 接收者用户名
    content TEXT NOT NULL,                 -- 消息内容
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- 消息时间
);

-- 消息索引：按时间查询（历史记录）
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at);
