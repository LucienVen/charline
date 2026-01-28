package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/LucienVen/charline/server/internal/errors"
)

// ============================================
// 统一响应结构
// ============================================

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ============================================
// HTTP 辅助方法
// ============================================

// RespondSuccess 成功响应
func RespondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Code:    errors.ErrCodeSuccess,
		Message: errors.GetMsg(errors.ErrCodeSuccess),
		Data:    data,
	})
}

// RespondError 业务错误响应
func RespondError(w http.ResponseWriter, bizErr *errors.BizError) {
	httpStatus := getHTTPStatus(bizErr.Code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := Response{
		Code:    bizErr.Code,
		Message: bizErr.Message,
	}

	// 有详细信息时放入 data 字段
	if bizErr.Details != nil && len(bizErr.Details) > 0 {
		response.Data = bizErr.Details
	}

	json.NewEncoder(w).Encode(response)
}

// RespondWithError 使用 HTTP 状态码 + 业务错误码
func RespondWithError(w http.ResponseWriter, httpStatus int, bizErr *errors.BizError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := Response{
		Code:    bizErr.Code,
		Message: bizErr.Message,
	}

	// 有详细信息时放入 data 字段
	if bizErr.Details != nil && len(bizErr.Details) > 0 {
		response.Data = bizErr.Details
	}

	json.NewEncoder(w).Encode(response)
}

// getHTTPStatus 根据业务错误码获取 HTTP 状态码
func getHTTPStatus(code int) int {
	switch {
	case code >= 5000:
		return http.StatusInternalServerError
	case code >= 3000:
		return http.StatusUnauthorized
	case code >= 2000:
		return http.StatusBadRequest
	case code >= 1000:
		return http.StatusBadRequest
	default:
		return http.StatusOK
	}
}
