-- 003_create_user_invite_relations.sql
-- 创建用户-邀请码关联表

CREATE TABLE IF NOT EXISTS user_invite_relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    invite_code_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (invite_code_id) REFERENCES invite_codes(id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_invite_user_id ON user_invite_relations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_invite_code_id ON user_invite_relations(invite_code_id);
