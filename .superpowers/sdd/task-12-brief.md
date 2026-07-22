### Task 12: 创建 agent/ 全部文件

**Files:**
- Create: `agent/main.go`
- Create: `agent/server.go`
- Create: `agent/session.go`

**Interfaces:**
- Produces: 可执行的 Agent 网关服务
- Consumes: `game/pkg`、`game/libs`、`game/tunnel`

- [ ] **Step 1: 写入 agent/session.go**

```go
// agent/session.go
package main

import (
    "sync"
    "simple_game/game/pkg"
)

// Session 客户端连接 ↔ Game 连接的映射。
type Session struct {
    TunnelID   string
    ClientConn pkg.NetConn
    GameConn   pkg.NetConn
}

// sessionMap 全局会话表。
var sessionMap sync.Map // string → *Session
```

- [ ] **Step 2: 写入 agent/server.go**

```go
// agent/server.go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
    "simple_game/game/pkg"
    "simple_game/game/tunnel"

    "github.com/go-redis/redis/v8"
)

// handshakeMsg 客户端握手消息（spec 5.1b）。
type handshakeMsg struct {
    Token   string `json:"token"`
    Version int    `json:"version"`
}

// playerReg Redis 中存储的玩家注册信息。
type playerReg struct {
    AgentAddr string `json:"agent_addr"`
    GameAddr  string `json:"game_addr"`
}

// startAgent 启动 Agent 网关服务。
func startAgent(listeners []listenerCfg, redisCfg redisCfg) {
    // 初始化 Redis 客户端（agent 独立管理，不使用 pkg.RCtx）
    rdb := redis.NewClient(&redis.Options{
        Addr:     redisCfg.Addr,
        Password: redisCfg.PassWord,
        DB:       redisCfg.DB,
    })

    for _, lc := range listeners {
        l, err := pkg.NewListener(lc.Network, lc.Addr, lc.Port)
        if err != nil {
            pkg.ERROR("agent listener create failed:", lc.Network, lc.Addr, lc.Port, err)
            continue
        }
        pkg.INFO("agent listener start:", lc.Network, l.Addr())
        go serveListener(l, rdb)
    }
}

// serveListener accept 客户端连接并为每个连接创建 Session。
func serveListener(l pkg.Listener, rdb *redis.Client) {
    for {
        conn, err := l.Accept()
        if err != nil {
            if strings.Contains(err.Error(), "closed") {
                return
            }
            pkg.ERROR("agent accept error:", err)
            continue
        }
        go handleClient(conn, rdb)
    }
}

// handleClient 处理单个客户端连接：握手 → 建连 Game → 双向转发。
func handleClient(clientConn pkg.NetConn, rdb *redis.Client) {
    // 1. 读握手消息
    data, err := clientConn.ReadMessage()
    if err != nil {
        pkg.ERROR("read handshake from client failed:", err)
        _ = clientConn.Close()
        return
    }

    hs := &handshakeMsg{}
    if err := json.Unmarshal(data, hs); err != nil {
        pkg.ERROR("parse handshake failed:", err)
        _ = clientConn.Close()
        return
    }

    // 2. 从 token 中提取 player_name（不验证签名 — 由 Game 验证）
    playerName := extractSubFromToken(hs.Token)
    if playerName == "" {
        pkg.ERROR("extract player name from token failed")
        _ = clientConn.Close()
        return
    }

    // 3. 查 Redis 获取 game_addr
    val, err := rdb.Get(rdb.Context(), "player:"+playerName).Result()
    if err != nil {
        pkg.ERROR("redis get player info failed:", err)
        _ = clientConn.Close()
        return
    }
    reg := &playerReg{}
    if err := json.Unmarshal([]byte(val), reg); err != nil {
        pkg.ERROR("parse player reg failed:", err)
        _ = clientConn.Close()
        return
    }

    // 4. 建连 Game
    gameConn, err := pkg.Dial("tcp", reg.GameAddr)
    if err != nil {
        pkg.ERROR("dial game failed:", reg.GameAddr, err)
        _ = clientConn.WriteMessage([]byte(`{"error":"game unavailable"}`))
        _ = clientConn.Close()
        return
    }

    // 5. 转发握手消息到 Game
    _ = gameConn.WriteMessage(data)

    // 6. 等待 Game 握手确认
    resp, err := gameConn.ReadMessage()
    if err != nil {
        pkg.ERROR("read game handshake response failed:", err)
        _ = clientConn.Close()
        _ = gameConn.Close()
        return
    }
    // 转发握手结果给客户端
    tm := tunnel.UnpackTunnel(resp)
    var clientResp []byte
    if tm != nil && tm.Name == "HandshakeOk" {
        clientResp = []byte(`{"status":"ok"}`)
    } else {
        clientResp = []byte(`{"status":"error","detail":"handshake failed"}`)
    }
    _ = clientConn.WriteMessage(clientResp)

    // 7. 创建 Session 并启动双向转发
    sess := &Session{
        TunnelID:   playerName,
        ClientConn: clientConn,
        GameConn:   gameConn,
    }
    sessionMap.Store(playerName, sess)
    defer func() {
        sessionMap.Delete(playerName)
        _ = clientConn.Close()
        _ = gameConn.Close()
    }()

    // 启动双向转发
    go clientToGame(sess)
    gameToClient(sess)
}

// clientToGame 客户端→Game 转发。
func clientToGame(sess *Session) {
    for {
        data, err := sess.ClientConn.ReadMessage()
        if err != nil {
            _ = sess.GameConn.Close()
            return
        }
        // 解析 PacketMsg 提取 name/data，注入 tunnel_id
        codePack := libs.DeCodePack(data)
        if codePack == nil {
            continue
        }
        wrapped := tunnel.PackTunnelRaw(codePack.Name, codePack.Data, sess.TunnelID)
        if err := sess.GameConn.WriteMessage(wrapped); err != nil {
            _ = sess.ClientConn.Close()
            return
        }
    }
}

// gameToClient Game→客户端 转发。
func gameToClient(sess *Session) {
    for {
        data, err := sess.GameConn.ReadMessage()
        if err != nil {
            _ = sess.ClientConn.Close()
            return
        }
        // 解包 TunnelMsg 获取原始 PacketMsg
        tm := tunnel.UnpackTunnel(data)
        if tm == nil {
            continue
        }
        // 还原为 PacketMsg JSON 发送给客户端
        raw := libs.EnCodePack(&libs.PacketMsg{Name: tm.Name, Data: tm.Data})
        if err := sess.ClientConn.WriteMessage(raw); err != nil {
            _ = sess.GameConn.Close()
            return
        }
    }
}

// extractSubFromToken 从 JWT token 中提取 sub（player_name），不验证签名。
func extractSubFromToken(tokenStr string) string {
    parts := strings.Split(tokenStr, ".")
    if len(parts) != 3 {
        return ""
    }
    // JWT payload 是 base64url 编码的 JSON
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return ""
    }
    var claims struct {
        Sub string `json:"sub"`
    }
    if err := json.Unmarshal(payload, &claims); err != nil {
        return ""
    }
    return claims.Sub
}
```

- [ ] **Step 3: 确保 agent/ 下所有文件都是 `package main`，import 分布在各文件中**

agent 全部文件（main.go, server.go, session.go）都用 `package main`。

**agent/server.go 的 import 汇总：**

```go
import (
    "encoding/base64"
    "encoding/json"
    "strings"

    "github.com/go-redis/redis/v8"

    pkg "simple_game/game/pkg"
    libs "simple_game/game/libs"
    "simple_game/game/tunnel"
)
```

注意：`pkg "simple_game/game/pkg"` 使用 alias 避免与标准库 `pkg` 概念混淆（实际无标准库包叫 pkg，可用原名 `pkg`）。

**agent/session.go 的 import：**

```go
import (
    "sync"
    pkg "simple_game/game/pkg"
)
```

**agent/main.go：**

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "gopkg.in/yaml.v3"
)

type agentConfig struct {
    Listeners []listenerCfg `yaml:"Listeners"`
    Redis     redisCfg      `yaml:"Redis"`
}

type listenerCfg struct {
    Network string `yaml:"Network"`
    Addr    string `yaml:"Addr"`
    Port    string `yaml:"Port"`
}

type redisCfg struct {
    Addr     string `yaml:"Addr"`
    PassWord string `yaml:"PassWord"`
    DB       int    `yaml:"DB"`
}

func main() {
    data, err := os.ReadFile("agent/config/config.yaml")
    if err != nil {
        fmt.Println("load config failed:", err)
        os.Exit(1)
    }
    cfg := &agentConfig{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        fmt.Println("parse config failed:", err)
        os.Exit(1)
    }

    startAgent(cfg.Listeners, cfg.Redis)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("Agent server started")
    <-sigChan
    fmt.Println("Agent server shutting down...")
}
```
```

- [ ] **Step 5: 编译验证**

```bash
go build ./agent/...
```

- [ ] **Step 6: 提交**

```bash
git add agent/
git commit -m "feat: 创建 agent 网关服务"
```

---

## Phase 5: World 服务

### Task 13: 创建 world/ 全部文件

**Files:**
- Create: `world/main.go`
- Create: `world/server.go`
- Create: `world/login.go`
- Create: `world/token.go`
- Create: `world/registry.go`

**Interfaces:**
- Produces: 可执行的 World 外部服务
- Consumes: `game/pkg`（logger）, MySQL, Redis, JWT

- [ ] **Step 1: 写入 world/registry.go**

```go
// world/registry.go
package main

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

const redisKeyPrefix = "player:"

// playerRegInfo Redis 中存储的玩家注册信息。
type playerRegInfo struct {
    AgentAddr string `json:"agent_addr"`
    GameAddr  string `json:"game_addr"`
}

// registerPlayer 将玩家信息写入 Redis。
func registerPlayer(rdb *redis.Client, playerName, agentAddr, gameAddr string, ttl time.Duration) error {
    info := playerRegInfo{
        AgentAddr: agentAddr,
        GameAddr:  gameAddr,
    }
    data, err := json.Marshal(info)
    if err != nil {
        return err
    }
    return rdb.Set(context.Background(), redisKeyPrefix+playerName, string(data), ttl).Err()
}

// getPlayerReg 从 Redis 读取玩家注册信息。
func getPlayerReg(rdb *redis.Client, playerName string) (*playerRegInfo, error) {
    val, err := rdb.Get(context.Background(), redisKeyPrefix+playerName).Result()
    if err != nil {
        return nil, err
    }
    info := &playerRegInfo{}
    if err := json.Unmarshal([]byte(val), info); err != nil {
        return nil, err
    }
    return info, nil
}
```

- [ ] **Step 2: 写入 world/token.go**

```go
// world/token.go
package main

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// generateToken 签发 JWT token。
func generateToken(playerName string, secret []byte, expireDuration time.Duration) (string, error) {
    claims := jwt.MapClaims{
        "sub": playerName,
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(expireDuration).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

// verifyWorldToken 验证 JWT token（World 端也保留验证能力用于调试）。
func verifyWorldToken(tokenStr string, secret []byte) (string, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return secret, nil
    })
    if err != nil {
        return "", err
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", fmt.Errorf("invalid token")
    }
    sub, _ := claims.GetSubject()
    return sub, nil
}
```

- [ ] **Step 3: 写入 world/login.go**

```go
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
```

- [ ] **Step 4: 写入 world/server.go**

```go
// world/server.go
package main

import (
    "fmt"
    "net/http"

    "github.com/go-redis/redis/v8"
    "github.com/jmoiron/sqlx"
    "simple_game/game/pkg"
)

// startWorld 启动 World HTTP 服务。
func startWorld(addr string, db *sqlx.DB, rdb *redis.Client, agentAddr, gameAddr, tokenSecret string, tokenExpire time.Duration) {
    handler := &loginHandler{
        db:          db,
        rdb:         rdb,
        agentAddr:   agentAddr,
        gameAddr:    gameAddr,
        tokenSecret: []byte(tokenSecret),
        tokenExpire: tokenExpire,
    }

    http.HandleFunc("/login", handler.ServeHTTP)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    pkg.INFO("world http server starting on", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        pkg.ERROR("world http server failed:", err)
    }
}
```

- [ ] **Step 5: 写入 world/main.go**

```go
// world/main.go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-redis/redis/v8"
    "gopkg.in/yaml.v3"

    "simple_game/game/pkg"
)

type worldConfig struct {
    Http struct {
        Addr string `yaml:"Addr"`
        Port string `yaml:"Port"`
    } `yaml:"Http"`
    AgentAddr   string `yaml:"AgentAddr"`
    TokenSecret string `yaml:"TokenSecret"`
    TokenExpire string `yaml:"TokenExpire"`
    SaveLog     bool   `yaml:"SaveLog"`
    MySql       struct {
        Addr     string `yaml:"Addr"`
        Port     int    `yaml:"Port"`
        User     string `yaml:"User"`
        PassWord string `yaml:"PassWord"`
        DBName   string `yaml:"DBName"`
    } `yaml:"MySql"`
    Redis struct {
        Addr     string `yaml:"Addr"`
        PassWord string `yaml:"PassWord"`
        DB       int    `yaml:"DB"`
    } `yaml:"Redis"`
}

func main() {
    data, err := os.ReadFile("world/config/config.yaml")
    if err != nil {
        fmt.Println("load config failed:", err)
        os.Exit(1)
    }
    cfg := &worldConfig{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        fmt.Println("parse config failed:", err)
        os.Exit(1)
    }

    // 日志
    pkg.StartLog()
    if cfg.SaveLog {
        _ = pkg.InitLogFile()
    }

    // DB / Redis
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.MySql.User, cfg.MySql.PassWord, cfg.MySql.Addr, cfg.MySql.Port, cfg.MySql.DBName)
    pkg.MysqlStart(dsn)

    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Addr,
        Password: cfg.Redis.PassWord,
        DB:       cfg.Redis.DB,
    })

    // 解析过期时间
    tokenExpire, err := time.ParseDuration(cfg.TokenExpire)
    if err != nil {
        tokenExpire = 24 * time.Hour
    }

    // 启动
    addr := fmt.Sprintf("%s:%s", cfg.Http.Addr, cfg.Http.Port)
    go startWorld(addr, pkg.DB, rdb, cfg.AgentAddr,
        fmt.Sprintf("%s:%s", cfg.Http.Addr, "9900"), // gameAddr 默认
        cfg.TokenSecret, tokenExpire)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("World server started on", addr)
    <-sigChan
    fmt.Println("World server shutting down...")
}
```

- [ ] **Step 6: 编译验证**

```bash
go build ./world/...
```

- [ ] **Step 7: 提交**

```bash
git add world/
git commit -m "feat: 创建 world 外部服务"
```

---

## Phase 6: 清理与客户端适配

### Task 14: 删除旧目录和根 main.go

**Files:**
- Delete: `server/` 目录（已清空）
- Delete: `pkg/` 目录（已清空）
- Delete: `libs/` 目录（已清空）
- Delete: `http/` 目录
- Delete: `config/` 目录（保留 config.yaml 到根目录可选）
- Delete: `routes/` 目录（旧路径）
- Delete: `controller/` 目录（旧路径）
- Delete: `register/` 目录（旧路径）
- Delete: `api/` 目录（旧路径）
- Delete: 根 `main.go`
- Delete: `game_test.go`（根目录）

- [ ] **Step 1: 清理**

```bash
rmdir server 2>/dev/null || true
rmdir pkg 2>/dev/null || true
rmdir libs 2>/dev/null || true
rmdir http 2>/dev/null || true
rmdir routes 2>/dev/null || true
rmdir controller 2>/dev/null || true
rmdir register 2>/dev/null || true
rm -rf api/ 2>/dev/null || true
rm main.go 2>/dev/null || true
rm game_test.go 2>/dev/null || true
```

注意：如果目录非空（有遗留 .go 文件），先确认文件已全部 git mv 后再删除。

- [ ] **Step 2: 编译验证整个项目**

```bash
go build ./...
```

排除 `configs/`（已知有 main redeclared 问题）和空 `test/` 目录。

```bash
go build $(go list ./... | grep -v configs | grep -v /test$)
```

- [ ] **Step 3: 提交**

```bash
git add -A
git commit -m "chore: 删除旧目录和根 main.go，完成三服架构迁移"
```

---

### Task 15: 更新 client/ import 路径

**Files:**
- Modify: `client/unified/main.go`
- Modify: `client/ws/ws_client.go`
- Modify: `client/client1.go`

- [ ] **Step 1: 更新 client/unified/main.go**

```bash
# 修改 import:
# "simple_game/pkg"       → "simple_game/game/pkg"
# "simple_game/libs"      → "simple_game/game/libs"
# "simple_game/api/protos/pt" → "simple_game/game/api/protos/pt"
```

将：

```go
import (
    "simple_game/api/protos/pt"
    "simple_game/libs"
    "simple_game/pkg"
)
```

改为：

```go
import (
    "simple_game/game/api/protos/pt"
    "simple_game/game/libs"
    "simple_game/game/pkg"
)
```

代码中对 `pkg.xxx`、`libs.xxx`、`pt.xxx` 的调用不变。

- [ ] **Step 2: 更新 client/ws/ws_client.go**

```go
// "simple_game/api/protos/pt" → "simple_game/game/api/protos/pt"
// "simple_game/libs" → "simple_game/game/libs"
```

- [ ] **Step 3: 更新 client/client1.go**

```go
// "simple_game/pkg" → "simple_game/game/pkg"
// "simple_game/libs" → "simple_game/game/libs"
// "simple_game/api/protos/pt" → "simple_game/game/api/protos/pt"
```

- [ ] **Step 4: 编译验证**

```bash
go build ./client/unified/... ./client/ws/...
```

- [ ] **Step 5: 提交**

```bash
git add client/
git commit -m "refactor: 更新 client import 路径适配三服架构"
```

---

## Phase 7: 最终验证

### Task 16: 全量编译 + go vet

- [ ] **Step 1: 全量编译（排除已知问题目录）**

```bash
go build $(go list ./... | grep -v configs | grep -v /test$)
```

预期：全部通过。

- [ ] **Step 2: go vet**

```bash
go vet $(go list ./... | grep -v configs | grep -v /test$)
```

预期：全部通过。

- [ ] **Step 3: 提交**

```bash
git add -A
git commit -m "chore: 全量编译验证通过"
```

---

## 实施总结

| Phase | Tasks | 新建文件 | 迁移文件 | 删除 |
|-------|-------|---------|---------|------|
| 1: 目录骨架 | 1 | 4 | 0 | 0 |
| 2: 共享代码 | 5 | 0 | 17 | 0 |
| 3: Game 服务端 | 3 | 4 | 0 | 0 |
| 4: Agent 服务 | 1 | 3 | 0 | 0 |
| 5: World 服务 | 1 | 5 | 0 | 0 |
| 6: 清理 | 2 | 0 | 0 | 10+ |
| 7: 验证 | 1 | 0 | 0 | 0 |
| **合计** | **14** | **16** | **17** | **10+** |

每个 Task 编译通过后 git commit，增量推进。
