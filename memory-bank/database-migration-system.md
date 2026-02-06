# 数据库迁移系统（Database Migration System）

**作者**: claude code  
**创建时间**: 2026-02-03  
**用途**: 记录 SQLite 数据库迁移系统的工作原理、使用方法和最佳实践

---

## 一、系统概述

### 1.1 什么是数据库迁移？

数据库迁移（Migration）是一种**版本化管理数据库结构变更**的机制，类似于 Git 管理代码变更。

**核心特性**：
- ✅ 版本化：每个变更有唯一版本号
- ✅ 可追溯：记录所有变更历史
- ✅ 幂等性：重复执行不会破坏数据
- ✅ 自动化：启动时自动检查并执行

### 1.2 为什么需要迁移系统？

| 场景 | 传统方式 | 迁移系统 |
|------|---------|---------|
| 添加新字段 | 手动执行 SQL | 自动检测并执行 |
| 多环境同步 | 手动同步结构 | 代码部署自动同步 |
| 回滚变更 | 手动恢复 | 版本化回滚 |
| 团队协作 | 口头通知 | 代码即文档 |

---

## 二、系统架构

### 2.1 核心组件

```
pkg/sqlite/
├── sqlite.go              # 数据库连接与迁移执行
├── migrations.go          # 迁移注册表
└── migrations/            # 迁移 SQL 文件
    ├── 001_init.sql
    ├── 002_add_public_key.sql
    └── 003_xxx.sql
```

### 2.2 迁移记录表（schema_migrations）

```sql
CREATE TABLE schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version INTEGER NOT NULL UNIQUE,      -- 迁移版本号（唯一）
    name TEXT NOT NULL,                   -- 迁移名称
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- 执行时间
);
```

**示例数据**：
```
id | version | name            | applied_at
1  | 1       | init            | 2026-01-26 06:41:12
2  | 2       | add_public_key  | 2026-02-03 07:15:55
```

---

## 三、工作原理

### 3.1 执行流程

```
Server 启动
    ↓
初始化数据库连接 (sqlite.New)
    ↓
执行迁移 (db.Migrate)
    ↓
┌─────────────────────────────────────┐
│ 遍历所有注册的迁移                    │
│   ↓                                  │
│ 检查 schema_migrations 表             │
│   ↓                                  │
│ 查询: SELECT COUNT(*) WHERE version=? │
│   ↓                                  │
│ count > 0 ?                          │
│   ├─ 是 → 跳过（已执行）              │
│   └─ 否 → 执行 SQL                   │
│       ↓                              │
│     记录到 schema_migrations          │
└─────────────────────────────────────┘
    ↓
完成（继续启动流程）
```

### 3.2 关键代码解析

#### 迁移检查与执行（pkg/sqlite/sqlite.go:79-123）

```go
func (db *DB) Migrate() error {
    migrations := GetMigrations()  // 获取所有迁移
    
    for _, m := range migrations {
        // 1. 检查迁移是否已执行
        var count int
        err := db.conn.QueryRow(
            "SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
            m.Version,
        ).Scan(&count)
        
        // 2. 如果已执行，跳过
        if count > 0 {
            continue  // ← 关键：防止重复执行
        }
        
        // 3. 执行迁移 SQL
        if _, err := db.conn.Exec(m.SQL); err != nil {
            return fmt.Errorf("执行迁移 v%d %s 失败: %w", m.Version, m.Name, err)
        }
        
        // 4. 记录迁移
        _, err = db.conn.Exec(
            "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
            m.Version, m.Name,
        )
    }
    
    return nil
}
```

#### 迁移注册（pkg/sqlite/migrations.go）

```go
//go:embed migrations/001_init.sql
var migration001Init string

//go:embed migrations/002_add_public_key.sql
var migration002AddPublicKey string

func GetMigrations() []Migration {
    return []Migration{
        {Version: 1, Name: "init", SQL: migration001Init},
        {Version: 2, Name: "add_public_key", SQL: migration002AddPublicKey},
    }
}
```

---

## 四、使用方法

### 4.1 添加新迁移

#### 步骤 1：创建 SQL 文件

```bash
# 文件名格式：<版本号>_<描述>.sql
touch pkg/sqlite/migrations/003_add_avatar_field.sql
```

```sql
-- 003_add_avatar_field.sql
-- 为 users 表添加头像字段

ALTER TABLE users ADD COLUMN avatar TEXT DEFAULT '';

-- 创建索引（可选）
CREATE INDEX idx_users_avatar ON users(avatar);
```

#### 步骤 2：注册迁移

编辑 `pkg/sqlite/migrations.go`：

```go
//go:embed migrations/003_add_avatar_field.sql
var migration003AddAvatarField string

func GetMigrations() []Migration {
    return []Migration{
        {Version: 1, Name: "init", SQL: migration001Init},
        {Version: 2, Name: "add_public_key", SQL: migration002AddPublicKey},
        {Version: 3, Name: "add_avatar_field", SQL: migration003AddAvatarField},  // ← 新增
    }
}
```

#### 步骤 3：重新编译并启动

```bash
make server
./bin/server
```

**日志输出**：
```
+0800 2026-02-03 15:30:00  INFO  数据库连接成功  {"path": "data/charline.db"}
```

迁移会自动执行，无需手动操作。

### 4.2 验证迁移

```bash
# 查看迁移记录
sqlite3 data/charline.db "SELECT * FROM schema_migrations ORDER BY version;"

# 查看表结构
sqlite3 data/charline.db "PRAGMA table_info(users);"
```

---

## 五、常见场景

### 5.1 添加新列

```sql
-- 004_add_email.sql
ALTER TABLE users ADD COLUMN email TEXT DEFAULT '';
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

### 5.2 修改列属性（SQLite 限制）

SQLite 不支持 `ALTER COLUMN`，需要重建表：

```sql
-- 005_modify_username_length.sql
BEGIN TRANSACTION;

-- 1. 创建新表
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE CHECK(length(username) <= 50),  -- 新约束
    -- ... 其他列
);

-- 2. 复制数据
INSERT INTO users_new SELECT * FROM users;

-- 3. 删除旧表
DROP TABLE users;

-- 4. 重命名新表
ALTER TABLE users_new RENAME TO users;

-- 5. 重建索引
CREATE UNIQUE INDEX idx_username ON users(username);

COMMIT;
```

### 5.3 创建新表

```sql
-- 006_create_messages.sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id)
);

CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_receiver ON messages(receiver_id);
```

---

## 六、安全机制

### 6.1 防止重复执行

```go
// 检查迁移是否已执行
if count > 0 {
    continue  // 跳过已执行的迁移
}
```

### 6.2 版本唯一性

```sql
version INTEGER NOT NULL UNIQUE  -- 确保版本号唯一
```

### 6.3 原子性保护（建议改进）

**当前实现**：无事务保护

**改进方案**：
```go
func (db *DB) Migrate() error {
    for _, m := range migrations {
        if count > 0 { continue }
        
        // 使用事务
        tx, _ := db.conn.Begin()
        
        // 执行迁移
        if _, err := tx.Exec(m.SQL); err != nil {
            tx.Rollback()
            return err
        }
        
        // 记录迁移
        if _, err := tx.Exec("INSERT INTO schema_migrations ..."); err != nil {
            tx.Rollback()
            return err
        }
        
        tx.Commit()  // 要么全成功，要么全失败
    }
}
```

---

## 七、最佳实践

### 7.1 命名规范

```
<版本号>_<动作>_<对象>.sql

示例：
001_init.sql                    # 初始化
002_add_public_key.sql          # 添加字段
003_create_messages_table.sql   # 创建表
004_drop_old_sessions.sql       # 删除表
005_modify_user_constraints.sql # 修改约束
```

### 7.2 版本号管理

- ✅ 使用连续整数：1, 2, 3, 4...
- ✅ 不要跳号：避免 1, 2, 5, 10
- ✅ 不要重用版本号
- ❌ 不要修改已执行的迁移

### 7.3 SQL 编写规范

```sql
-- ✅ 好的实践
-- 1. 添加注释说明变更原因
-- 2. 使用 IF NOT EXISTS（幂等性）
-- 3. 添加默认值（兼容现有数据）

ALTER TABLE users ADD COLUMN avatar TEXT DEFAULT '';

-- ❌ 避免的做法
-- 1. 删除生产数据
-- 2. 无默认值的 NOT NULL 列（现有数据会报错）
-- 3. 复杂的数据转换（应该分步执行）

ALTER TABLE users ADD COLUMN email TEXT NOT NULL;  -- ❌ 现有数据会报错
```

### 7.4 测试流程

```bash
# 1. 在开发环境测试
rm data/charline.db  # 删除数据库
./bin/server         # 重新初始化

# 2. 验证迁移
sqlite3 data/charline.db "SELECT * FROM schema_migrations;"

# 3. 验证表结构
sqlite3 data/charline.db "PRAGMA table_info(users);"

# 4. 测试数据兼容性
# 插入测试数据，确保新字段不影响现有功能
```

---

## 八、故障排查

### 8.1 迁移失败

**现象**：
```
ERROR: 执行迁移 v3 add_avatar_field 失败: duplicate column name: avatar
```

**原因**：列已存在（可能手动执行过）

**解决**：
```sql
-- 方案 1：修改迁移 SQL（添加 IF NOT EXISTS）
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT DEFAULT '';

-- 方案 2：手动标记迁移已完成
INSERT INTO schema_migrations (version, name) VALUES (3, 'add_avatar_field');
```

### 8.2 版本冲突

**现象**：
```
ERROR: UNIQUE constraint failed: schema_migrations.version
```

**原因**：版本号重复

**解决**：
```bash
# 查看已有版本
sqlite3 data/charline.db "SELECT version FROM schema_migrations ORDER BY version;"

# 使用下一个可用版本号
```

### 8.3 数据库锁定

**现象**：
```
ERROR: database is locked
```

**原因**：其他进程正在使用数据库

**解决**：
```bash
# 查找占用进程
lsof | grep charline.db

# 关闭进程
pkill -f "./bin/server"
```

---

## 九、与其他系统对比

| 特性 | 本项目 | Rails Migrations | Flyway | Liquibase |
|------|--------|------------------|--------|-----------|
| 语言 | Go | Ruby | Java | Java/XML |
| 版本管理 | 整数 | 时间戳 | 版本号 | Changeset ID |
| 回滚支持 | ❌ | ✅ | ✅ | ✅ |
| 事务保护 | ❌ | ✅ | ✅ | ✅ |
| 自动执行 | ✅ | ✅ | ✅ | ✅ |
| 复杂度 | 简单 | 中等 | 中等 | 复杂 |

---

## 十、未来改进方向

### 10.1 事务保护

```go
// 为每个迁移添加事务
tx, _ := db.conn.Begin()
defer tx.Rollback()

tx.Exec(m.SQL)
tx.Exec("INSERT INTO schema_migrations ...")

tx.Commit()
```

### 10.2 回滚支持

```go
type Migration struct {
    Version int
    Name    string
    Up      string  // 升级 SQL
    Down    string  // 回滚 SQL
}

func (db *DB) Rollback(version int) error {
    // 执行 Down SQL
    // 删除 schema_migrations 记录
}
```

### 10.3 迁移验证

```go
func (db *DB) ValidateMigrations() error {
    // 检查版本号连续性
    // 检查 SQL 语法
    // 检查依赖关系
}
```

### 10.4 迁移日志

```go
// 记录详细的迁移日志
log.Info("执行迁移", 
    zap.Int("version", m.Version),
    zap.String("name", m.Name),
    zap.Duration("duration", elapsed))
```

---

## 十一、参考资料

- **代码位置**：
  - `pkg/sqlite/sqlite.go:78-137` - 迁移执行逻辑
  - `pkg/sqlite/migrations.go` - 迁移注册表
  - `pkg/sqlite/migrations/` - SQL 文件目录

- **相关文档**：
  - `architecture.md` - 数据库架构设计
  - `progress.md` - 迁移历史记录

- **外部参考**：
  - [Rails Migrations](https://guides.rubyonrails.org/active_record_migrations.html)
  - [Flyway Documentation](https://flywaydb.org/documentation/)
  - [SQLite ALTER TABLE](https://www.sqlite.org/lang_altertable.html)

---

**最后更新时间**: 2026-02-03
