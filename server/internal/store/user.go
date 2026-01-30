package store

import (
	"database/sql"

	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/pkg/sqlite"
	"go.uber.org/zap"
)

// User 用户模型
type User struct {
	ID           int64
	Username     string
	PublicKey    string
	TokenVersion int
	CreatedAt    string
	LastLogin    sql.NullString
}

// UserStore 用户存储
type UserStore struct {
	db     *sqlite.DB
	logger *logger.Logger
}

// NewUserStore 创建用户存储实例
func NewUserStore(db *sqlite.DB, log *logger.Logger) *UserStore {
	return &UserStore{
		db:     db,
		logger: log,
	}
}

// GetByID 根据用户 ID 获取用户
func (s *UserStore) GetByID(userID int64) (*User, *apperrors.BizError) {
	var user User
	err := s.db.QueryRow(
		"SELECT id, username, public_key, token_version, created_at, last_login FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.PublicKey, &user.TokenVersion, &user.CreatedAt, &user.LastLogin)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		s.logger.Error("查询用户失败",
			zap.Int64("user_id", userID),
			zap.String("error", err.Error()))
		return nil, apperrors.ErrSystemError
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *UserStore) GetByUsername(username string) (*User, *apperrors.BizError) {
	var user User
	err := s.db.QueryRow(
		"SELECT id, username, public_key, token_version, created_at, last_login FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PublicKey, &user.TokenVersion, &user.CreatedAt, &user.LastLogin)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		s.logger.Error("查询用户失败",
			zap.String("username", username),
			zap.String("error", err.Error()))
		return nil, apperrors.ErrSystemError
	}

	return &user, nil
}

// UpdateTokenVersion 更新用户 Token 版本号
func (s *UserStore) UpdateTokenVersion(userID int64, newVersion int) *apperrors.BizError {
	_, err := s.db.Exec(
		"UPDATE users SET token_version = ? WHERE id = ?",
		newVersion, userID,
	)
	if err != nil {
		s.logger.Error("更新 Token 版本失败",
			zap.Int64("user_id", userID),
			zap.Int("new_version", newVersion),
			zap.String("error", err.Error()))
		return apperrors.ErrSystemError
	}

	s.logger.Info("更新 Token 版本成功",
		zap.Int64("user_id", userID),
		zap.Int("new_version", newVersion))
	return nil
}
// GetByPublicKey 根据公钥获取用户
func (s *UserStore) GetByPublicKey(publicKey string) (*User, error) {
	var user User
	err := s.db.QueryRow(
		"SELECT id, username, public_key, token_version, created_at, last_login FROM users WHERE public_key = ?",
		publicKey,
	).Scan(&user.ID, &user.Username, &user.PublicKey, &user.TokenVersion, &user.CreatedAt, &user.LastLogin)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		s.logger.Error("查询用户失败",
			zap.String("public_key", publicKey),
			zap.String("error", err.Error()))
		return nil, err
	}

	return &user, nil
}
