package sqlite

import (
	_ "embed"
)

//go:embed migrations/001_init.sql
var migration001Init string

//go:embed migrations/002_add_public_key.sql
var migration002AddPublicKey string

//go:embed migrations/003_create_user_invite_relations.sql
var migration003CreateUserInviteRelations string

// Migration 迁移定义
type Migration struct {
	Version int    // 迁移版本号
	Name    string // 迁移名称
	SQL     string // SQL 语句
}

// GetMigrations 获取所有迁移
func GetMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "init",
			SQL:     migration001Init,
		},
		{
			Version: 2,
			Name:    "add_public_key",
			SQL:     migration002AddPublicKey,
		},
		{
			Version: 3,
			Name:    "create_user_invite_relations",
			SQL:     migration003CreateUserInviteRelations,
		},
	}
}
