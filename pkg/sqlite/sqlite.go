package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB 数据库实例
type DB struct {
	conn *sql.DB
	path string
}

// Config 数据库配置
type Config struct {
	DataDir string // 数据目录
	Name    string // 数据库文件名（如: charline.db）
}

// New 创建数据库实例
func New(cfg Config) (*DB, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 构建数据库文件路径
	dbPath := filepath.Join(cfg.DataDir, cfg.Name)

	// 打开数据库连接
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// 执行迁移
	if err := db.Migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn 获取底层连接
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Path 获取数据库文件路径
func (db *DB) Path() string {
	return db.path
}

// Migrate 执行数据库迁移
func (db *DB) Migrate() error {
	migrations := GetMigrations()

	for _, m := range migrations {
		// 检查迁移是否已执行
		var count int
		err := db.conn.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			m.Version,
		).Scan(&count)

		if err != nil {
			// schema_migrations 表不存在，创建并初始化
			db.initSchemaMigrations()
			// 重新检查
			err = db.conn.QueryRow(
				"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
				m.Version,
			).Scan(&count)
			if err != nil {
				return fmt.Errorf("查询迁移状态失败: %w", err)
			}
		}

		// 如果迁移已执行，跳过
		if count > 0 {
			continue
		}

		// 执行迁移
		if _, err := db.conn.Exec(m.SQL); err != nil {
			return fmt.Errorf("执行迁移 v%d %s 失败: %w", m.Version, m.Name, err)
		}

		// 记录迁移
		if _, err := db.conn.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
			m.Version, m.Name,
		); err != nil {
			return fmt.Errorf("记录迁移失败: %w", err)
		}
	}

	return nil
}

// initSchemaMigrations 初始化迁移记录表
func (db *DB) initSchemaMigrations() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_migration_version ON schema_migrations(version);
	`)
	return err
}

// Exec 执行 SQL 语句（无返回结果）
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query 执行查询
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow 执行查询并返回单行
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}
