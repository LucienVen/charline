package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/LucienVen/charline/server/internal/errors"
)

// DecodeJSON 统一处理请求体 JSON 解析
//
// 参数:
//   - w: ResponseWriter
//   - r: Request
//   - dest: 目标结构体指针
//
// 返回:
//   - true: 解析成功
//   - false: 解析失败（已自动响应错误）
//
// 使用示例:
//   var req httputil.ActivateInviteRequest
//   if !httputil.DecodeJSON(w, r, &req) {
//       return  // 错误已响应，直接返回
//   }
func DecodeJSON(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		RespondWithError(w, http.StatusBadRequest,
			errors.ErrInvalidParam.
				WithDetail("reason", "参数解析失败").
				WithDetail("error", err.Error()))
		return false
	}
	return true
}
