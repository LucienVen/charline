package httputil

// ============================================
// Invite 端点请求 DTO
// ============================================

// ActivateInviteRequest POST /api/v1/invite/activate
type ActivateInviteRequest struct {
	Code      string `json:"code"`
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

// ============================================
// Invite 端点响应 DTO
// ============================================

// GenerateInviteCodeResponse POST /api/v1/invite/generate
type GenerateInviteCodeResponse struct {
	Code string `json:"code"`
}

// ActivateInviteResponse POST /api/v1/invite/activate
type ActivateInviteResponse struct {
	Token   string `json:"token"`
	Version int    `json:"version"`
}

// ValidateTokenResponse GET /api/v1/validate
type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
	Version  int    `json:"version,omitempty"`
}

// ============================================
// Auth 端点请求 DTO（Phase 2.1）
// ============================================

// LoginRequest POST /api/v1/auth/login
type LoginRequest struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

// ============================================
// Auth 端点响应 DTO（Phase 2.1）
// ============================================

// ChallengeResponse GET /api/v1/auth/challenge
type ChallengeResponse struct {
	Nonce     string `json:"nonce"`
	ExpiresIn int    `json:"expires_in"`
}

// LoginResponse POST /api/v1/auth/login
type LoginResponse struct {
	Token   string `json:"token"`
	Version int    `json:"version"`
}
