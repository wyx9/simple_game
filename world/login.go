// world/login.go
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

// loginHandler 处理 POST /login 请求。
type loginHandler struct {
	db             *sqlx.DB
	rdb            *redis.Client
	agentAddr      string
	gameAddr       string
	tokenSecret    []byte
	tokenExpire    time.Duration
}

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	AgentAddr string `json:"agent_addr"`
}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Password == "" {
		http.Error(w, "name and password required", http.StatusBadRequest)
		return
	}

	// 验证账号密码（简化：首次登录即允许，实际应查 DB 验证）
	// TODO: 接入 DB 验证密码

	// 写 Redis 注册信息
	if err := registerPlayer(h.rdb, req.Name, h.agentAddr, h.gameAddr, h.tokenExpire); err != nil {
		http.Error(w, "register player failed", http.StatusInternalServerError)
		return
	}

	// 签发 token
	token, err := generateToken(req.Name, h.tokenSecret, h.tokenExpire)
	if err != nil {
		http.Error(w, "generate token failed", http.StatusInternalServerError)
		return
	}

	resp := loginResponse{
		Token:     token,
		AgentAddr: h.agentAddr,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
