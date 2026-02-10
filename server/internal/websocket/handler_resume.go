package websocket

// handleResumeRequest 处理 Resume 请求
func (h *WSMessageHandler) handleResumeRequest(conn *Connection, msg *Message) {
	// 解析 Resume 请求
	var resumeReq ResumeRequestPayload
	if err := msg.UnmarshalPayload(&resumeReq); err != nil {
		h.sendError(conn, "INVALID_PAYLOAD", "Failed to parse resume request")
		return
	}

	// 验证必填字段
	if resumeReq.ResumeToken == "" {
		h.sendError(conn, "INVALID_REQUEST", "Missing resume token")
		return
	}

	// 尝试恢复 Session
	sess, err := h.sessionManager.Resume(resumeReq.ResumeToken, conn.ID())
	if err != nil {
		// Resume 失败，返回失败响应
		resumeResp := ResumeResponsePayload{
			Success: false,
			Message: "Failed to resume session: " + err.Error(),
		}

		respMsg, err := NewMessage(MessageTypeResumeResponse, resumeResp)
		if err != nil {
			h.sendError(conn, "INTERNAL_ERROR", "Failed to create response")
			return
		}

		respData, err := respMsg.Marshal()
		if err != nil {
			h.sendError(conn, "INTERNAL_ERROR", "Failed to marshal response")
			return
		}

		conn.Send(respData)
		return
	}

	// Resume 成功，设置连接的用户 ID
	conn.SetUserID(sess.UserID)

	// 将连接添加到连接池
	h.pool.Add(conn)

	// 发送 Resume 成功响应
	resumeResp := ResumeResponsePayload{
		Success:   true,
		SessionID: sess.ID,
	}

	respMsg, err := NewMessage(MessageTypeResumeResponse, resumeResp)
	if err != nil {
		h.sendError(conn, "INTERNAL_ERROR", "Failed to create response")
		return
	}

	respData, err := respMsg.Marshal()
	if err != nil {
		h.sendError(conn, "INTERNAL_ERROR", "Failed to marshal response")
		return
	}

	conn.Send(respData)
}
