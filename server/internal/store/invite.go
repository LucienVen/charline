package store

import (
	"database/sql"
	"math/rand"
	"time"

	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/pkg/sqlite"
	"go.uber.org/zap"
)

// InviteStore 邀请码存储
type InviteStore struct {
	db     *sqlite.DB
	logger *logger.Logger
}

// NewInviteStore 创建邀请码存储实例
func NewInviteStore(db *sqlite.DB, log *logger.Logger) *InviteStore {
	return &InviteStore{
		db:     db,
		logger: log,
	}
}

// Generate 生成新邀请码
// 返回格式: INV-XXXXXXXX (8位大写字母+数字，排除易混淆字符 0OI1)
func (s *InviteStore) Generate() (string, *apperrors.BizError) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code := s.generateCode()

		// 检查是否已存在
		exists, err := s.Exists(code)
		if err != nil {
			s.logger.Error("检查邀请码失败",
				zap.String("code", code),
				zap.String("error", err.Error()))
			return "", apperrors.ErrSystemError
		}

		if !exists {
			// 插入新邀请码
			_, err := s.db.Exec(
				"INSERT INTO invite_codes (code, created_at) VALUES (?, ?)",
				code, time.Now(),
			)
			if err != nil {
				s.logger.Error("插入邀请码失败",
					zap.String("code", code),
					zap.String("error", err.Error()))
				return "", apperrors.ErrSystemError
			}

			s.logger.Info("生成邀请码",
				zap.String("code", code),
				zap.Int("attempt", attempt+1))
			return code, nil
		}
	}

	return "", apperrors.ErrSystemError
}

// Activate 激活邀请码
// 参数: code - 邀请码, username - 用户名
// 返回: 业务错误
func (s *InviteStore) Activate(code, username string) *apperrors.BizError {
	if !s.isValidFormat(code) {
		return apperrors.ErrInviteInvalid
	}

	// 检查邀请码是否存在且未使用
	valid, bizErr := s.Validate(code)
	if bizErr != nil {
		return bizErr
	}
	if !valid {
		return apperrors.ErrInviteUsed
	}

	// 激活邀请码
	result, err := s.db.Exec(
		"UPDATE invite_codes SET used_at = ?, username = ? WHERE code = ?",
		time.Now(), username, code,
	)
	if err != nil {
		s.logger.Error("激活邀请码失败",
			zap.String("code", code),
			zap.String("error", err.Error()))
		return apperrors.ErrSystemError
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrInviteNotFound
	}

	s.logger.Info("激活邀请码",
		zap.String("code", code),
		zap.String("username", username))

	return nil
}

// Validate 验证邀请码是否有效
// 返回: (是否有效, 业务错误)
func (s *InviteStore) Validate(code string) (bool, *apperrors.BizError) {
	if !s.isValidFormat(code) {
		return false, apperrors.ErrInviteInvalid
	}

	var usedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT used_at FROM invite_codes WHERE code = ?",
		code,
	).Scan(&usedAt)

	if err == sql.ErrNoRows {
		return false, apperrors.ErrInviteNotFound
	}
	if err != nil {
		s.logger.Error("查询邀请码失败",
			zap.String("code", code),
			zap.String("error", err.Error()))
		return false, apperrors.ErrSystemError
	}

	// 如果 used_at 为 NULL，说明未使用
	return !usedAt.Valid, nil
}

// Exists 检查邀请码是否存在
func (s *InviteStore) Exists(code string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM invite_codes WHERE code = ?",
		code,
	).Scan(&count)
	return count > 0, err
}

// IsUsed 检查邀请码是否已使用
func (s *InviteStore) IsUsed(code string) (bool, *apperrors.BizError) {
	var usedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT used_at FROM invite_codes WHERE code = ?",
		code,
	).Scan(&usedAt)

	if err == sql.ErrNoRows {
		return false, apperrors.ErrInviteNotFound
	}
	if err != nil {
		s.logger.Error("查询邀请码失败",
			zap.String("code", code),
			zap.String("error", err.Error()))
		return false, apperrors.ErrSystemError
	}

	return usedAt.Valid, nil
}

// GetUserByCode 获取使用该邀请码的用户名
func (s *InviteStore) GetUserByCode(code string) (string, *apperrors.BizError) {
	var username sql.NullString
	err := s.db.QueryRow(
		"SELECT username FROM invite_codes WHERE code = ?",
		code,
	).Scan(&username)

	if err == sql.ErrNoRows {
		return "", apperrors.ErrInviteNotFound
	}
	if err != nil {
		s.logger.Error("查询用户名失败",
			zap.String("code", code),
			zap.String("error", err.Error()))
		return "", apperrors.ErrSystemError
	}

	return username.String, nil
}

// generateCode 生成随机邀请码
// 格式: INV-XXXXXXXX (8位字符)
func (s *InviteStore) generateCode() string {
	const chars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ" // 排除 0OI1
	const codeLength = 8

	b := make([]byte, codeLength)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return "INV-" + string(b)
}

// isValidFormat 验证邀请码格式
// 格式: INV-XXXXXXXX (前缀 INV- + 8位大写字母/数字)
func (s *InviteStore) isValidFormat(code string) bool {
	if len(code) != 12 {
		return false
	}
	if code[:4] != "INV-" {
		return false
	}
	for _, c := range code[4:] {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}
