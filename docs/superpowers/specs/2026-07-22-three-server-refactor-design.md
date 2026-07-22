# 三服架构重构设计

> 状态: 设计完成，待审查 | 日期: 2026-07-22

## 1. 目标

将当前单进程游戏服务器拆分为三个独立服务：

| 服务 | 职责 | 对外端口 |
|------|------|----------|
| **agent** | 网关：保持客户端长连接（TCP/WS），消息双向转发，不做业务逻辑 | 8888 (TCP) / 8889 (WS) |
| **game** | 逻辑服：运行 Actor 系统，处理全部游戏逻辑 | 内部 TCP + 9901 (gRPC) |
| **world** | 外部服：HTTP 登录认证、Token 签发、玩家注册中心 | 9902 (HTTP) |

## 2. 架构拓扑

```
┌──────────┐     HTTP /login     ┌──────────┐
│  Client   │ ──────────────────▶ │  World   │
│ (Browser) │ ◀──── JSON ─────── │  :9902   │
└─────┬─────┘    {token, agent}   └─────┬────┘
      │                                 │
      │ TCP / WebSocket                 │ Redis Write
      │                                 │ player_id→game_addr
      ▼                                 ▼
┌──────────┐    TCP 1:1 隧道       ┌──────────┐
│  Agent   │ ◀──────────────────▶  │   Game   │
│ :8888/89 │   TunnelMsg+PacketMsg │ Actor系统 │
└──────────┘                       │ :9901 gRPC│
                                    └──────────┘
                                           │
                                           ▼ Redis Read
                                      player_id→game_addr
                                      (用于 agent 路由查询)
```

## 3. 登录流程

```
Client              World                 Redis              Agent              Game
  │                   │                     │                  │                  │
  │ POST /login       │                     │                  │                  │
  │ {name, password}  │                     │                  │                  │
  │──────────────────▶│                     │                  │                  │
  │                   │ 验证账号密码          │                  │                  │
  │                   │ 加载玩家数据          │                  │                  │
  │                   │ 分配 Agent/Game      │                  │                  │
  │                   │─────────────────────▶│                  │                  │
  │                   │ SET player:name      │                  │                  │
  │                   │   agent_addr         │                  │                  │
  │                   │   game_addr          │                  │                  │
  │                   │◀─────────────────────│                  │                  │
  │                   │ 签发 JWT token       │                  │                  │
  │ {token,agent_addr}│                     │                  │                  │
  │◀──────────────────│                     │                  │                  │
  │                   │                     │                  │                  │
  │ Connect(token) ─────────────────────────────────────────▶│                  │
  │                   │                     │ GET player:name │                │
  │                   │                     │ → game_addr     │                  │
  │                   │                     │◀────────────────│                  │
  │                   │                     │                 │ Connect(token) ─▶│
  │                   │                     │                 │                  │ 验 token
  │                   │                     │                 │                  │ loadPlayer
  │                   │                     │                 │                  │ StartNewActor
  │◀────────────────────────────────────────── LoginRes ─────────────────────────│
  │                   │                     │                 │                  │
  │◀═════════════════════ 之后所有消息经 Agent 透传 ═════════════════════════════▶│
```

## 4. 目录结构

```
simple_game/
├── agent/                     # 网关服（独立可执行）
│   ├── main.go
│   ├── config/config.yaml
│   ├── server.go              # accept 客户端、建连 Game、启动转发协程
│   └── session.go             # Session 结构体 + sessionMap
│
├── world/                     # 外部服（独立可执行）
│   ├── main.go
│   ├── config/config.yaml
│   ├── server.go              # HTTP 路由注册
│   ├── login.go               # POST /login handler
│   ├── token.go               # JWT 签发 / 验证
│   └── registry.go            # Redis 读写 player 注册信息
│
├── game/                      # 逻辑服 + 所有共享代码
│   ├── main.go
│   ├── config/config.yaml
│   ├── server.go              # 监听 Agent 连接、消息分发
│   ├── handle_msg.go          # handleMessage（从原 server.go 拆出）
│   ├── grpc.go                # gRPC 服务（保留）
│   ├── actor_base.go          # IActor + ActorBase
│   ├── actor_manager.go       # AManager + CastMsg/CallMsg
│   ├── actor_player.go        # Player Actor
│   ├── tunnel/
│   │   └── tunnel.go          # TunnelMsg 结构 + 打包/解包
│   ├── pkg/                   # 公共库（agent/world import）
│   │   ├── logger.go
│   │   ├── netconn.go
│   │   ├── wsconn.go
│   │   ├── conn_dial.go
│   │   ├── msg.go
│   │   ├── redis.go
│   │   ├── mysql.go
│   │   └── uuid.go
│   ├── libs/
│   │   └── packet_msg.go      # PacketMsg / Pack2Msg / DeCodePack
│   ├── api/protos/pt/
│   ├── routes/
│   ├── controller/
│   └── register/
│
└── client/                    # 测试客户端（import 路径适配后不变）
```

依赖方向：

```
agent ──import──▶ game/pkg, game/libs, game/api, game/tunnel
world ──import──▶ game/pkg, game/api
game  ──import──▶ game/pkg, game/libs, ...（内部）
```

## 5. 转发协议

### 5.1 Client ↔ Agent

与当前完全相同：TCP 4字节长度头分帧 / WS 文本帧，承载 `PacketMsg` JSON：

```json
{"name":"LoginReq","data":"<base64 of protobuf>"}
{"name":"HeartReq","data":""}
```

Agent 不解析内容，原封不动转发。

### 5.1b 客户端首条握手消息

客户端连上 Agent 后，第一条消息不是 PacketMsg，而是握手消息：

```json
{"token":"eyJhbGciOiJIUzI1NiIs...","version":1}
```

Agent 收到后：
1. 解出 `token` → 无需验证签名，只从中提取 `sub`（player_name）
2. `Redis GET player:<name>` → 拿到 `game_addr`
3. `pkg.Dial("tcp", game_addr)` → 建连 Game
4. 把 token 原文 + player_name 作为首条消息发给 Game（Game 侧验证签名 + 创建 Actor）
5. 存入 sessionMap：`player_name → Session{ClientConn, GameConn}`
6. 回复客户端 `{"status":"ok"}` (握手完成)
7. 启动双向转发

之后所有消息走 TunnelMsg 转发。

### 5.2 Agent ↔ Game

在 `PacketMsg` 外层包裹 `TunnelMsg`，增加路由标识字段：

```go
// game/tunnel/tunnel.go
type TunnelMsg struct {
    Name     string `json:"name"`
    Data     []byte `json:"data"`
    TunnelID string `json:"tunnel_id"` // 客户端/玩家唯一标识
}
```

**线上格式：**
```json
{"name":"LoginReq","data":"<base64>","tunnel_id":"player_12345678"}
```

**Agent 转发逻辑：**

```
客户端→Game: 收到 msg → 注入 tunnel_id → 写入 GameConn
Game→客户端: 收到 msg → 读取 tunnel_id → 查 sessionMap → 写入 ClientConn
```

Agent 全程只存取 `tunnel_id` 字段，不对 `name`/`data` 做任何解析。

## 6. Agent 设计

### 6.1 核心结构

```go
// agent/session.go
type Session struct {
    TunnelID   string
    ClientConn pkg.NetConn // 客户端连接（TCP 或 WS）
    GameConn   pkg.NetConn // 到 Game 的 TCP 连接
}

// agent/server.go
type AgentServer struct {
    gameAddr string
    listener pkg.Listener
    redis    *redis.Client
    sessions sync.Map // tunnel_id → *Session
}
```

### 6.2 连接生命周期

1. 客户端连上来，accept → 创建 `Session{ClientConn=conn}`
2. 读第一条消息，从中提取 `tunnel_id`
3. 用 `tunnel_id` 中的 player 信息查 Redis → 获取 `game_addr`
4. `pkg.Dial("tcp", game_addr)` 建连到 Game → `Session.GameConn = gameConn`
5. 把 tunnel_id 连同首条消息转发给 Game
6. 启动两个 goroutine：
   - `clientToGame`: `ClientConn.ReadMessage → 注入tunnel_id → GameConn.WriteMessage`
   - `gameToClient`: `GameConn.ReadMessage → 提取tunnel_id → sessions.Load → ClientConn.WriteMessage`
7. 任一端断开 → 关闭 Session → 关闭另一端连接

### 6.3 Agent 约束

- Agent 不 import `game/actor_*`、`game/routes`、`game/controller`
- Agent 不 import protobuf 包（不解析消息体）
- Agent 只依赖 `game/pkg`（NetConn/Dial/logger）、`game/libs`（DeCodePack 只用于取 tunnel_id）、`game/tunnel`

## 7. Game 设计

### 7.1 变更点

Game 是原 `server/` 包的迁移和精简：

| 项目 | 变更 |
|------|------|
| 客户端监听 | **删除**，改为监听 Agent 连接 |
| `handleConnection` | 简化：不再处理 Login 认证（token 由 World 签发） |
| `handleMessage` | 提取 `tunnel_id` 用于 Actor 管理 |
| `Player.conn` | 语义从"客户端连接"变为"到 Agent 的隧道连接"，代码不变 |
| `StartNewPlayerActor` | 移除登录逻辑，token 验证提前在 server.go 完成 |

### 7.2 Game 监听 Agent 连接

```go
// game/server.go
func Start() {
    l, _ := pkg.NewListener("tcp", config.Conf.Game.Addr, config.Conf.Game.Port)
    for {
        conn, _ := l.Accept()
        go handleAgentConn(conn)  // 类似原 handleConnection，但预期 TunnelMsg
    }
}

func handleAgentConn(conn pkg.NetConn) {
    defer conn.Close()
    // 读首条消息（含 token + tunnel_id）
    msg, _ := conn.ReadMessage()
    tunnel := tunnel.Unpack(msg)
    // 验证 token → 提取 player_id → 创建 Actor
    // 注册 tunnel_id → conn 映射（回包时用）
    // 消息循环：tunnel.Unpack → ActorManner.CastMsg → Actor.Handler → PackTunnel → conn.Write
}
```

### 7.3 Actor 系统（不变）

`actor_base.go`、`actor_manager.go`、`actor_player.go` 的核心逻辑不修改。Player 的 `conn` 字段指向 Agent 连接，`HandlerByClient` 通过 `conn.WriteMessage(pack)` 写回包，由 Game 端转发层自动注入 `tunnel_id` 返回给 Agent。

## 8. World 设计

### 8.1 HTTP 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/login` | 账号密码认证，返回 token + agent 地址 |
| GET | `/health` | 健康检查 |

### 8.2 POST /login 流程

```go
// world/login.go
func HandleLogin(w http.ResponseWriter, r *http.Request) {
    // 1. 解析 {name, password}
    // 2. 从 DB 验证账号密码（首次登录且密码匹配则允许）
    // 3. 从配置中选一个可用的 game_addr + agent_addr
    // 4. Redis SET player:{name} → {agent_addr, game_addr}
    // 5. 签发 JWT token: {player_id, exp}
    // 6. 返回 JSON: {"token": "...", "agent_addr": "127.0.0.1:8888"}
}
```

### 8.3 Redis 注册信息

```
Key:   player:<player_name>
Value: {"agent_addr":"...","game_addr":"..."}
TTL:   跟随 token 过期时间（或不过期，token 过期后由 Game 侧清理）
```

### 8.4 Token 设计

- 算法：HMAC-SHA256 JWT
- 密钥：配置在 world 和 game 各一份（world 签发，game 验证）
- Payload：`{ sub: player_name, iat, exp }`
- 过期：24 小时

## 9. 迁移清单

### 新建文件（11个）

- `agent/main.go`、`agent/config/config.yaml`、`agent/server.go`、`agent/session.go`
- `world/main.go`、`world/config/config.yaml`、`world/server.go`、`world/login.go`、`world/token.go`、`world/registry.go`
- `game/tunnel/tunnel.go`

### 迁移文件（仅改 import 路径）

| 原路径 | 新路径 |
|--------|--------|
| `pkg/*.go` | `game/pkg/*.go` |
| `libs/*.go` | `game/libs/*.go` |
| `api/protos/pt/` | `game/api/protos/pt/` |
| `routes/*.go` | `game/routes/*.go` |
| `controller/*.go` | `game/controller/*.go` |
| `register/*.go` | `game/register/*.go` |
| `server/server.go` | `game/server.go`（重写：去掉客户端监听，增加 Agent 监听 + token 验证） |
| `server/actor_base.go` | `game/actor_base.go` |
| `server/actor_manager.go` | `game/actor_manager.go` |
| `server/actor_player.go` | `game/actor_player.go` |
| `server/grpc.go` | `game/grpc.go` |

### 删除

| 路径 | 原因 |
|------|------|
| `server/` 目录 | 已拆入 agent/game |
| `pkg/` 目录 | 已迁入 game/pkg |
| `libs/` 目录 | 已迁入 game/libs |
| `http/` 目录 | 已迁入 world |
| `config/` 目录 | 已拆入各服 config |
| 根 `main.go` | 三个服务各自有 main.go |
| `routes/routes.go` | 旧包路径删除 |

### 不动

- `client/` — import 路径适配
- `doc/` — 文档保留
- `config/config.yaml` — 可选保留为全局参考

## 10. 配置示例

### agent/config/config.yaml

```yaml
Listeners:
  - Network: tcp
    Addr: "0.0.0.0"
    Port: "8888"
  - Network: ws
    Addr: "0.0.0.0"
    Port: "8889"

GameAddr: "127.0.0.1:9900"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

### game/config/config.yaml

```yaml
Listen:
  Addr: "127.0.0.1"
  Port: "9900"

Grpc:
  Addr: "127.0.0.1"
  Port: "9901"

SaveLog: false

TokenSecret: "shared-secret-key"

MySql:
  Addr: "127.0.0.1"
  Port: 3306
  User: "root"
  PassWord: "123456"
  DBName: "simple_game"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

### world/config/config.yaml

```yaml
Http:
  Addr: "127.0.0.1"
  Port: "9902"

AgentAddr: "127.0.0.1:8888"
TokenSecret: "shared-secret-key"
TokenExpire: 24h

SaveLog: false

MySql:
  Addr: "127.0.0.1"
  Port: 3306
  User: "root"
  PassWord: "123456"
  DBName: "simple_game"

Redis:
  Addr: "127.0.0.1:6379"
  PassWord: ""
  DB: 4
```

## 11. 启动方式

```bash
# 三个终端分别启动
go run ./game
go run ./agent
go run ./world

# 或者编译后部署
go build -o bin/game  ./game
go build -o bin/agent ./agent
go build -o bin/world ./world
```

启动顺序任意（world 依赖 Redis/MySQL，game 依赖 Redis/MySQL，agent 依赖 Redis），各服启动后即可接受请求。

## 12. 不变的设计原则

- Actor 模型：IActor → ActorBase → AManager 接口和机制不变
- 消息路由：`PacketMsg → routes.Route → controller` 链路不变
- 网络抽象：`NetConn / Listener / Dial` 接口不变
- 消息封装：`Pack2Msg / DeCodePack / PacketMsg` 不变
- tcpConn/wsConn 的双向透传能力直接复用，Agent 转发层不写新协议
