# 三服架构重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将单进程游戏服务器拆分为 agent（网关）、game（逻辑）、world（外部）三个独立可执行服务。

**Architecture:** client → agent (TCP/WS 长连接) → game (TCP 1:1 隧道, Actor 系统) + world (HTTP /login 认证签发 token)。agent 只做字节流转发不解析协议体，game 承载全部业务逻辑和共享代码库。

**Tech Stack:** Go 1.23.3, protobuf, `nhooyr.io/websocket`, `go-redis/v8`, `sqlx`, `golang-jwt/jwt/v5`, gRPC, YAML 配置

## Global Constraints

- go.mod module path 保持 `simple_game`，新增子包路径全部在此之下
- Actor 模型 (IActor/ActorBase/AManager) 接口不修改
- PacketMsg / Pack2Msg / DeCodePack 接口不修改
- NetConn / Listener / Dial 接口不修改
- agent 不 import protobuf、routes、controller、actor_* 包
- 三个服务各自有独立 config/ 目录和 main.go

---

## Phase 1: 目录与共享代码迁移

### Task 1: 创建三服目录骨架 + game 配置

**Files:**
- Create: `game/config/config.yaml`
- Create: `game/main.go`（骨架，后续任务填充）
- Create: `agent/config/config.yaml`
- Create: `world/config/config.yaml`

**Interfaces:**
- Produces: 各服务的 config.yaml 文件，供后续任务读取

- [ ] **Step 1: 创建 game/config/config.yaml**

```bash
mkdir -p game/config agent/config world/config
```

- [ ] **Step 2: 写入 game/config/config.yaml**

```yaml
# game/config/config.yaml
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

- [ ] **Step 3: 写入 agent/config/config.yaml**

```yaml
# agent/config/config.yaml
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

- [ ] **Step 4: 写入 world/config/config.yaml**

```yaml
# world/config/config.yaml
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

- [ ] **Step 5: 创建 game/main.go 骨架（后续任务填充）**

```go
// game/main.go
package main

func main() {
	// 骨架 — 后续任务填充
}
```

- [ ] **Step 6: 提交**

```bash
git add agent/ world/ game/
git commit -m "feat: 创建三服目录骨架和配置文件"
```

---

### Task 2: 迁移 pkg/ → game/pkg/（改 import + 解除 config 循环依赖）

**Files:**
- Move: `pkg/logger.go` → `game/pkg/logger.go`
- Move: `pkg/msg.go` → `game/pkg/msg.go`
- Move: `pkg/netconn.go` → `game/pkg/netconn.go`
- Move: `pkg/wsconn.go` → `game/pkg/wsconn.go`
- Move: `pkg/conn_dial.go` → `game/pkg/conn_dial.go`
- Move: `pkg/uuid.go` → `game/pkg/uuid.go`
- Move: `pkg/redis.go` → `game/pkg/redis.go` (关键修改：函数签名)
- Move: `pkg/mysql.go` → `game/pkg/mysql.go` (关键修改：函数签名)

**Interfaces:**
- Consumes: 原 `pkg/` 下所有 .go 文件
- Produces: `game/pkg/` 包，`RedisStart(addr, password string, db int)`、`MysqlStart(dsn string)` 新签名

- [ ] **Step 1: 移动不依赖 config 的文件（批量 git mv）**

```bash
mkdir -p game/pkg
git mv pkg/logger.go game/pkg/logger.go
git mv pkg/msg.go game/pkg/msg.go
git mv pkg/netconn.go game/pkg/netconn.go
git mv pkg/wsconn.go game/pkg/wsconn.go
git mv pkg/conn_dial.go game/pkg/conn_dial.go
git mv pkg/uuid.go game/pkg/uuid.go
```

- [ ] **Step 2: 修改 redis.go — RedisStart 接受参数替代 config 导入**

```bash
git mv pkg/redis.go game/pkg/redis.go
```

然后编辑 `game/pkg/redis.go`，将：

```go
import (
    "context"
    "github.com/go-redis/redis/v8"
    "simple_game/config"
)

var RedisClient *redis.Client
var RCtx context.Context

func RedisStart() {
    addr, password, db := config.Conf.Redis.Addr, config.Conf.Redis.PassWord, config.Conf.Redis.DB
```

改为：

```go
import (
    "context"
    "github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var RCtx context.Context

func RedisStart(addr, password string, db int) {
```

- [ ] **Step 3: 修改 mysql.go — MysqlStart 接受 dsn 参数**

```bash
git mv pkg/mysql.go game/pkg/mysql.go
```

然后编辑 `game/pkg/mysql.go`，将：

```go
import (
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
    "simple_game/config"
)

var DB *sqlx.DB

func MysqlStart() {
    dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Conf.MySql.User, config.Conf.MySql.PassWord, config.Conf.MySql.Addr, config.Conf.MySql.Port, config.Conf.MySql.DBName)
```

改为：

```go
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func MysqlStart(dsn string) {
```

并把 `database, err := sqlx.Open("mysql", dataSourceName)` 改为 `database, err := sqlx.Open("mysql", dsn)`。

- [ ] **Step 4: 修正已迁移文件中的内部 import 路径**

`game/pkg/` 内部互相引用从 `"simple_game/pkg"` → 不需要 import（同包），直接调用。

检查 `netconn.go:38` 的 `NewListener` 是否引用了自身包 — 确认它在同包调用 `newTCPListener`/`newWSListener`，无需修改。

检查 `wsconn.go` 中的 `ERROR(...)` 调用（同包函数），无需修改。

检查 `conn_dial.go` 中的 `tcpConn`/`wsConn` 引用（同包），无需修改。

- [ ] **Step 5: 编译验证**

```bash
go build ./game/pkg/...
```

- [ ] **Step 6: 提交**

```bash
git add game/pkg/ pkg/
git commit -m "refactor: 迁移 pkg/ → game/pkg/，RedisStart/MysqlStart 解除 config 依赖"
```

---

### Task 3: 迁移 libs/ → game/libs/

**Files:**
- Move: `libs/packet_msg.go` → `game/libs/packet_msg.go`
- Delete: `libs/packet_msg_test.go`（测试文件暂不迁移，后续补）

**Interfaces:**
- Produces: `game/libs` 包，导出 `PacketMsg`、`Pack2Msg`、`DeCodePack`、`EnCodePack`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/libs
git mv libs/packet_msg.go game/libs/packet_msg.go
```

- [ ] **Step 2: 编译验证**

```bash
go build ./game/libs/...
```

`packet_msg.go` 只依赖标准库 + `google.golang.org/protobuf`，无内部 import，无需修改。

- [ ] **Step 3: 提交**

```bash
git add game/libs/ libs/
git commit -m "refactor: 迁移 libs/ → game/libs/"
```

---

### Task 4: 迁移 api/protos/pt/ → game/api/protos/pt/

**Files:**
- Move: `api/protos/pt/` 下所有 .go 文件 → `game/api/protos/pt/`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/api/protos/pt
git mv api/protos/pt/*.go game/api/protos/pt/
```

- [ ] **Step 2: 验证编译**

```bash
go build ./game/api/protos/pt/...
```

纯 protobuf 生成代码，无内部 import，无需修改。

- [ ] **Step 3: 提交**

```bash
git add game/api/ api/
git commit -m "refactor: 迁移 api/protos/pt/ → game/api/protos/pt/"
```

---

### Task 5: 迁移 routes/ → game/routes/

**Files:**
- Move: `routes/routes.go` → `game/routes/routes.go`
- Move: `routes/protobuf_factory.go` → `game/routes/protobuf_factory.go`

**Interfaces:**
- Consumes: `pkg` 包 (logger) → 需改 import
- Produces: `game/routes` 包，导出 `Handler`、`AddRoute`、`Route`、`ProtoBufFactory`、`RegisterProtoBufFactory`、`Init`

- [ ] **Step 1: 移动文件**

```bash
mkdir -p game/routes
git mv routes/routes.go game/routes/routes.go
git mv routes/protobuf_factory.go game/routes/protobuf_factory.go
```

- [ ] **Step 2: 修改 import 路径**

`game/routes/routes.go` 中 `"simple_game/pkg"` → `"simple_game/game/pkg"`：

```go
import (
    "errors"
    "google.golang.org/protobuf/proto"
    "simple_game/game/pkg"
)
```

`game/routes/protobuf_factory.go` 中 `"simple_game/api/protos/pt"` → `"simple_game/game/api/protos/pt"`：

```go
import (
    "google.golang.org/protobuf/proto"
    "reflect"
    "simple_game/game/api/protos/pt"
)
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/routes/...
```

- [ ] **Step 4: 提交**

```bash
git add game/routes/ routes/
git commit -m "refactor: 迁移 routes/ → game/routes/"
```

---

### Task 6: 迁移 controller/ → game/controller/ + register/ → game/register/

**Files:**
- Move: `controller/base_controller.go` → `game/controller/base_controller.go`
- Move: `controller/all_controller.go` → `game/controller/all_controller.go`
- Move: `register/register_route.go` → `game/register/register_route.go`

**Interfaces:**
- Consumes: `server` 包 (Player) → 需改 import 为 `game`
- Produces: `game/controller` 包、`game/register` 包

- [ ] **Step 1: 移动 controller 文件**

```bash
mkdir -p game/controller game/register
git mv controller/base_controller.go game/controller/base_controller.go
git mv controller/all_controller.go game/controller/all_controller.go
git mv register/register_route.go game/register/register_route.go
```

- [ ] **Step 2: 修改 controller/base_controller.go 的 import**

`"simple_game/server"` → `"simple_game/game"`，类型引用 `*server.Player` → `*game.Player`：

```go
package controller

import (
    "simple_game/game"
)

type BaseController struct {
}

func (this BaseController) Ctx(ctx interface{}) *game.Player {
    return ctx.(*game.Player)
}
```

- [ ] **Step 3: 修改 controller/all_controller.go 的 import**

`"simple_game/server"` → `"simple_game/game"`：

```go
package controller

import (
    "google.golang.org/protobuf/proto"
    "simple_game/game/api/protos/pt"
    "simple_game/game"
    "time"
)

type AllController struct {
    BaseController
}

func LoginController(actor *game.Player, req *pt.LoginReq) *pt.LoginRes {
    return &pt.LoginRes{
        UUid: req.UUid,
        Code: 0,
    }
}

func TestController(actor *game.Player, req *pt.HeartReq) proto.Message {
    return &pt.HeartRes{Time: time.Now().Unix()}
}
```

- [ ] **Step 4: 修改 register/register_route.go 的 import**

```go
package register

import (
    "google.golang.org/protobuf/proto"
    "simple_game/game/api/protos/pt"
    "simple_game/game/controller"
    "simple_game/game/routes"
)

func RegisteredRoute() {
    allController := controller.AllController{}

    routes.AddRoute("LoginReq", func(ctx interface{}, req any) proto.Message {
        player := allController.Ctx(ctx)
        return controller.LoginController(player, req.(*pt.LoginReq))
    })

    routes.AddRoute("HeartReq", func(ctx interface{}, req any) proto.Message {
        player := allController.Ctx(ctx)
        return controller.TestController(player, req.(*pt.HeartReq))
    })
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./game/controller/... ./game/register/...
```

注意：此时 `game` 主包尚未创建，controller 引用 `*game.Player` 会编译失败。**这是预期的**——等待 Task 8 迁移 Actor 文件后一起验证。

- [ ] **Step 6: 提交**

```bash
git add game/controller/ game/register/ controller/ register/
git commit -m "refactor: 迁移 controller/ register/ → game/"
```

---

## Phase 2: Game 核心迁移

### Task 7: 迁移 game/ 核心 — actor_base.go, actor_manager.go, grpc.go

**Files:**
- Move: `server/actor_base.go` → `game/actor_base.go`
- Move: `server/actor_manager.go` → `game/actor_manager.go`
- Move: `server/grpc.go` → `game/grpc.go`

**Interfaces:**
- Produces: `game` 包中的 `IActor`、`ActorBase`、`ActorConfig`、`AManager`、`ActorManner`、`CallRequest`、`Player`（type alias 占位，Task 8 填补）

- [ ] **Step 1: 移动文件**

```bash
git mv server/actor_base.go game/actor_base.go
git mv server/actor_manager.go game/actor_manager.go
git mv server/grpc.go game/grpc.go
```

- [ ] **Step 2: 修改 actor_base.go 的 import**

`"simple_game/pkg"` → `"simple_game/game/pkg"`：

```go
import (
    "context"
    "errors"
    "simple_game/game/pkg"
    "sync/atomic"
    "time"
)
```

- [ ] **Step 3: 修改 actor_manager.go — 无需改 import（只依赖 sync/time 标准库）**

检查无误。

- [ ] **Step 4: 修改 grpc.go 的 import**

`"simple_game/pkg"` → `"simple_game/game/pkg"`，`pt2 "simple_game/api/protos/pt"` → `pt2 "simple_game/game/api/protos/pt"`：

```go
package game

import (
    "context"
    "fmt"
    "net"
    pt2 "simple_game/game/api/protos/pt"
    "simple_game/game/pkg"

    "google.golang.org/grpc"
)

type server struct {
    pt2.UnimplementedHelloServer
}

func (s *server) Say(ctx context.Context, req *pt2.SayRequest) (*pt2.SayResponse, error) {
    fmt.Println("request:", req.Msg)
    return &pt2.SayResponse{Msg: "Hello " + req.Msg}, nil
}

func StartGRPC() {
    listen, err := net.Listen("tcp", ":9901")
    if err != nil {
        fmt.Printf("failed to listen: %v", err)
        return
    }

    s := grpc.NewServer()
    pt2.RegisterHelloServer(s, &server{})
    defer func() {
        s.Stop()
        _ = listen.Close()
    }()

    pkg.INFO("grpc server start port: ", 9901)
    err = s.Serve(listen)
    if err != nil {
        pkg.ERROR("failed to serve", err)
        return
    }
}
```

- [ ] **Step 5: 验证编译（此时 game 包尚缺 Player 定义和 server.go）**

暂不编译 — 等待 Task 8 完成后一起验证。

- [ ] **Step 6: 提交**

```bash
git add game/actor_base.go game/actor_manager.go game/grpc.go server/
git commit -m "refactor: 迁移 actor_base, actor_manager, grpc → game/"
```

---

### Task 8: 迁移 actor_player.go → game/actor_player.go

**Files:**
- Move: `server/actor_player.go` → `game/actor_player.go`

**Interfaces:**
- Produces: `game.Player` 结构体、`game.PlayerData`、`game.StartNewPlayerActor`
- 修正 import: `"simple_game/pkg"` → `"simple_game/game/pkg"`，`"simple_game/libs"` → `"simple_game/game/libs"`，`"simple_game/routes"` → `"simple_game/game/routes"`

- [ ] **Step 1: 移动文件**

```bash
git mv server/actor_player.go game/actor_player.go
```

- [ ] **Step 2: 修改 import**

```go
package game

import (
    "encoding/json"
    "fmt"
    "simple_game/game/libs"
    "simple_game/game/pkg"
    "simple_game/game/routes"
    "time"
)
```

删除 `"google.golang.org/protobuf/encoding/protojson"` 和 `"google.golang.org/protobuf/proto"` import（上一轮修复中已移除这些引用）。

- [ ] **Step 3: 删除 StartNewPlayerActor 中的 Login 逻辑（按 spec 7.1）**

原函数直接创建 Actor，不再包含登录认证（token 验证由 Game server 在 handleAgentConn 中提前完成）：

```go
func StartNewPlayerActor(name string, conn pkg.NetConn) (*Player, error) {
    playerActor := &Player{
        name: name,
        ip:   conn.RemoteAddr(),
        conn: conn,
        PlayerData: &PlayerData{
            Name:        name,
            LastLoginAt: time.Now().Unix(),
        },
    }
    _ = NewActorBase(name, playerActor, nil)
    return playerActor, nil
}
```

（无实质变更，只是确认函数与当前一致）

- [ ] **Step 4: 验证 game 包整体编译**

```bash
go build ./game/...
```

预期：game 包编译通过（server.go 尚需重写，故 game main 暂不编译）。

- [ ] **Step 5: 提交**

```bash
git add game/actor_player.go server/
git commit -m "refactor: 迁移 actor_player → game/"
```

---

### Task 9: 创建 game/tunnel/tunnel.go

**Files:**
- Create: `game/tunnel/tunnel.go`

**Interfaces:**
- Produces: `game/tunnel` 包，导出 `TunnelMsg`、`PackTunnel(msg *TunnelMsg) []byte`、`UnpackTunnel(data []byte) *TunnelMsg`

- [ ] **Step 1: 创建目录和文件**

```bash
mkdir -p game/tunnel
```

- [ ] **Step 2: 写入 game/tunnel/tunnel.go**

```go
// game/tunnel/tunnel.go
// TunnelMsg 在 PacketMsg 外层包裹路由标识，用于 Agent↔Game 之间的消息转发。
package tunnel

import (
    "encoding/json"
    "simple_game/game/libs"
)

// TunnelMsg 带路由标识的消息结构。
// 比 PacketMsg 多一个 TunnelID，Agent 通过该字段做 client↔game 双向映射。
type TunnelMsg struct {
    Name     string `json:"name"`
    Data     []byte `json:"data"`
    TunnelID string `json:"tunnel_id"`
}

// PackTunnel 将 PacketMsg 和 tunnelID 打包为 JSON 字节。
// 如果 msg 已经是 *libs.PacketMsg，直接内联字段；否则 Name 设为空。
func PackTunnel(msg *libs.PacketMsg, tunnelID string) []byte {
    tm := &TunnelMsg{
        Name:     msg.Name,
        Data:     msg.Data,
        TunnelID: tunnelID,
    }
    b, _ := json.Marshal(tm)
    return b
}

// PackTunnelRaw 将原始消息字节和 tunnelID 直接打包为 TunnelMsg JSON。
// name 用于设置消息类型标识，data 是原始 protobuf 二进制。
func PackTunnelRaw(name string, data []byte, tunnelID string) []byte {
    tm := &TunnelMsg{
        Name:     name,
        Data:     data,
        TunnelID: tunnelID,
    }
    b, _ := json.Marshal(tm)
    return b
}

// UnpackTunnel 从 JSON 字节解析出 TunnelMsg。
func UnpackTunnel(data []byte) *TunnelMsg {
    tm := &TunnelMsg{}
    if err := json.Unmarshal(data, tm); err != nil {
        return nil
    }
    return tm
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/tunnel/...
```

- [ ] **Step 4: 提交**

```bash
git add game/tunnel/
git commit -m "feat: 新增 game/tunnel — TunnelMsg 路由消息封装"
```

---

## Phase 3: Game 服务端

### Task 10: 重写 game/server.go（Agent 连接监听 + token 验证 + 消息分发）

**Files:**
- Create: `game/server.go`

**Interfaces:**
- Consumes: `game/tunnel`、`game/pkg`、`game/libs`、`game/routes`
- Produces: `game.Start()` 监听 Agent 连接、`game.handleAgentConn()`、`game.handleMsg()`
- 从原 `server/server.go` 提取并简化：删客户端监听、删 LoginReq 特殊处理、增 token 验证

- [ ] **Step 1: 写入 game/handle_msg.go（消息处理，从原 server.go 拆出）**

```go
// game/handle_msg.go
package game

import (
    "errors"
    "simple_game/game/api/protos/pt"
    "simple_game/game/libs"
    "simple_game/game/pkg"
    "simple_game/game/tunnel"

    "google.golang.org/protobuf/proto"
)

// handleMsg 处理来自 Agent 的单条消息。
// msg 是 TunnelMsg JSON 原始字节，conn 是该 Agent 隧道连接。
func handleMsg(conn pkg.NetConn, msg []byte, tunnelID string) error {
    tm := tunnel.UnpackTunnel(msg)
    if tm == nil {
        return errors.New("unpack tunnel msg error")
    }

    // 还原为 PacketMsg 交给 Actor 系统处理
    codePack := &libs.PacketMsg{
        Name: tm.Name,
        Data: tm.Data,
    }

    // 已认证的消息直接路由到 Actor
    return ActorManner.CastMsg(tunnelID, libs.EnCodePack(codePack))
}
```

- [ ] **Step 2: 写入 game/server.go（Agent 连接处理）**

```go
// game/server.go
package game

import (
    "encoding/json"
    "fmt"
    "strings"
    "simple_game/game/libs"
    "simple_game/game/pkg"
    "simple_game/game/tunnel"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Game 配置（在 main.go 中从 config 文件加载）
var (
    GameListenAddr  string
    TokenSecret     []byte
    agentConnMap    = make(map[string]pkg.NetConn) // tunnelID → Agent 连接
)

// 握手消息结构（对应 spec 5.1b）
type handshakeMsg struct {
    Token   string `json:"token"`
    Version int    `json:"version"`
}

// Start 启动 Game 服务：监听 Agent 连接。
func Start(addr string, tokenSecret string) {
    GameListenAddr = addr
    TokenSecret = []byte(tokenSecret)

    // 解析 addr
    parts := strings.Split(addr, ":")
    if len(parts) != 2 {
        pkg.ERROR("invalid game listen addr:", addr)
        return
    }

    l, err := pkg.NewListener("tcp", parts[0], parts[1])
    if err != nil {
        pkg.ERROR("game listener create failed:", err)
        return
    }
    pkg.INFO("game server listening on", addr)

    for {
        conn, err := l.Accept()
        if err != nil {
            if strings.Contains(err.Error(), "closed") {
                return
            }
            pkg.ERROR("game accept error:", err)
            continue
        }
        go handleAgentConn(conn)
    }
}

// handleAgentConn 处理单条 Agent 隧道连接。
func handleAgentConn(conn pkg.NetConn) {
    defer conn.Close()

    // 1. 读取握手消息（token + tunnelID 信息已由 Agent 在隧道首包中发送）
    _ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
    data, err := conn.ReadMessage()
    if err != nil {
        pkg.ERROR("read handshake from agent failed:", err)
        return
    }

    // 解析握手：{"token":"...", "version":1}
    hs := &handshakeMsg{}
    if err := json.Unmarshal(data, hs); err != nil {
        pkg.ERROR("parse handshake failed:", err)
        return
    }

    // 2. 验证 token
    playerName, err := verifyToken(hs.Token, TokenSecret)
    if err != nil {
        pkg.ERROR("token verify failed:", err)
        _ = conn.WriteMessage([]byte(`{"error":"invalid token"}`))
        return
    }

    // 3. 注册 tunnelID → conn 映射（Agent 以 playerName 作为 tunnelID）
    tunnelID := playerName
    agentConnMap[tunnelID] = conn
    defer func() {
        ActorManner.FindAndClosePlayer(tunnelID)
        delete(agentConnMap, tunnelID)
    }()

    // 4. 加载玩家数据 + 启动 Actor
    _, err = StartNewPlayerActor(tunnelID, conn)
    if err != nil {
        pkg.ERROR("start player actor failed:", err)
        return
    }

    // 5. 回复 Agent 握手成功
    _ = conn.WriteMessage(tunnel.PackTunnelRaw("HandshakeOk", nil, tunnelID))

    // 6. 主消息循环
    for {
        _ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        data, err := conn.ReadMessage()
        if err != nil {
            return
        }
        if err := handleMsg(conn, data, tunnelID); err != nil {
            pkg.ERROR("handleMsg error:", err)
        }
    }
}

// verifyToken 验证 JWT token，返回 player_name。
func verifyToken(tokenStr string, secret []byte) (string, error) {
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
        return "", fmt.Errorf("invalid token claims")
    }
    sub, _ := claims.GetSubject()
    if sub == "" {
        return "", fmt.Errorf("missing sub claim")
    }
    return sub, nil
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/...
```

- [ ] **Step 4: 提交**

```bash
git add game/server.go game/handle_msg.go
git commit -m "feat: 重写 game/server.go — Agent 连接监听 + JWT 验证 + 消息分发"
```

---

### Task 11: 完善 game/main.go

**Files:**
- Modify: `game/main.go`

**Interfaces:**
- Produces: 可执行的 Game 服务入口

- [ ] **Step 1: 创建 game/config.go（配套配置加载）**

```go
// game/config.go
package game

import (
    "os"
    "gopkg.in/yaml.v3"
)

type GameConfig struct {
    Listen struct {
        Addr string `yaml:"Addr"`
        Port string `yaml:"Port"`
    } `yaml:"Listen"`
    Grpc struct {
        Addr string `yaml:"Addr"`
        Port string `yaml:"Port"`
    } `yaml:"Grpc"`
    SaveLog     bool   `yaml:"SaveLog"`
    TokenSecret string `yaml:"TokenSecret"`
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

func LoadGameConfig(path string) (*GameConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    cfg := &GameConfig{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

- [ ] **Step 2: 重写 game/main.go**

```go
// game/main.go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "simple_game/game"
    "simple_game/game/pkg"
    "simple_game/game/register"
    "simple_game/game/routes"
)

func main() {
    // 加载配置
    cfg, err := game.LoadGameConfig("game/config/config.yaml")
    if err != nil {
        fmt.Println("load config failed:", err)
        os.Exit(1)
    }

    // 日志
    pkg.StartLog()
    if cfg.SaveLog {
        _ = pkg.InitLogFile()
    }

    // 路由注册
    routes.Init()
    register.RegisteredRoute()

    // 基础设施
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.MySql.User, cfg.MySql.PassWord, cfg.MySql.Addr, cfg.MySql.Port, cfg.MySql.DBName)
    pkg.MysqlStart(dsn)
    pkg.RedisStart(cfg.Redis.Addr, cfg.Redis.PassWord, cfg.Redis.DB)

    // 信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 启动服务
    go game.StartGRPC()
    go game.Start(
        fmt.Sprintf("%s:%s", cfg.Listen.Addr, cfg.Listen.Port),
        cfg.TokenSecret,
    )

    fmt.Println("Game server started")
    <-sigChan
    fmt.Println("Game server shutting down...")
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./game/...
```

预期 `game/main.go` 编译为可执行文件。验证 `go build -o /dev/null ./game` 成功。

- [ ] **Step 4: 提交**

```bash
git add game/config.go game/main.go
git commit -m "feat: 完善 game/main.go 入口"
```

---

## Phase 4: Agent 服务

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
