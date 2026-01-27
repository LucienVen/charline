-- 002_add_public_key.sql
-- 添加 public_key 列以支持 Ed25519 公钥存储

-- 添加 public_key 列（TEXT 类型，NOT NULL）
ALTER TABLE users ADD COLUMN public_key TEXT NOT NULL DEFAULT '';

-- 创建唯一索引以确保公钥唯一性和查询性能
CREATE UNIQUE INDEX idx_users_public_key ON users(public_key);

-- 注意：
-- 1. public_key 存储 Base64 编码的 Ed25519 公钥（32 字节 -> ~44 字符）
-- 2. 唯一索引确保一个公钥只能绑定一个用户
-- 3. 索引提升基于公钥的查询性能（登录时验证签名需要查询公钥）
